# Infrastructure as Code for resume-analyzer

This directory contains Pulumi Go infrastructure code for deploying resume-analyzer to AWS ECS Fargate, with support for Floci localstack.

## Structure

```
infra/
├── main.go                      # Entry point
├── Pulumi.yaml                  # Project configuration
├── Pulumi.dev.yaml              # Development stack (Floci localstack)
├── go.mod / go.sum              # Go module
└── internal/
    ├── config/                  # Stack configuration types & loading
    ├── iam/                     # IAM roles (task, execution)
    ├── taskdef/                 # ECS Task Definition
    ├── loadbalancer/            # ALB Target Group + Listener Rule
    ├── service/                 # ECS Fargate Service
    ├── scaling/                 # Application Auto Scaling
    └── tags/                    # Resource tagging helper
```

## Prerequisites

- Go 1.24+
- Pulumi CLI
- Floci localstack running on a VM (dev stack targets existing Floci)

## Deploy to Floci (dev)

Your Floci VM already has the network infrastructure. You need to provide these values in `Pulumi.dev.yaml`:

| Value | Description | Where to find |
|-------|-------------|---------------|
| `network.vpcId` | VPC ID in Floci | `aws ec2 describe-vpcs` |
| `network.subnetIds` | Subnet IDs (2+ AZs) | `aws ec2 describe-subnets` |
| `network.securityGroupId` | Security Group for ECS tasks | `aws ec2 describe-security-groups` |
| `loadBalancer.internal.listenerArn` | ALB Listener ARN | `aws elbv2 describe-listeners` |

```bash
cd infra
pulumi stack select dev  # or: pulumi stack init dev
# Edit Pulumi.dev.yaml with your Floci resource IDs
# Set LLM_API_KEY in envVars
pulumi up
```

## Configuration

All configuration is in `Pulumi.dev.yaml`. Key sections:

| Section | Purpose |
|---------|---------|
| `service` | ECS cluster, service name, desired count |
| `container` | Port, CPU, Memory for Fargate task |
| `network` | VPC, subnets, security group (ECS task SG) |
| `healthCheck` | ALB target group health check |
| `loadBalancer` | ALB listener ARN, priority, host header |
| `envVars` | Container environment variables |
| `scaling` | Auto-scaling (disabled on Floci) |
| `iam` | Optional pre-existing role ARNs |
| `dns` | Optional Route53 record |

## Localstack Provider

The dev stack uses an explicit AWS provider pointing to Floci:

```go
// main.go creates provider with:
Endpoints: aws.ProviderEndpointArray{
    aws.ProviderEndpointArgs{
        Ecs:                     pulumi.String("http://floci:4566"),
        Iam:                     pulumi.String("http://floci:4566"),
        Cloudwatch:              pulumi.String("http://floci:4566"),
        Elasticloadbalancing:    pulumi.String("http://floci:4566"),
        Route53:                 pulumi.String("http://floci:4566"),
        Ec2:                     pulumi.String("http://floci:4566"),
        Logs:                    pulumi.String("http://floci:4566"),
        Applicationautoscaling:  pulumi.String("http://floci:4566"),
    },
}
```

Environment variables for Pulumi CLI:
```bash
export AWS_ENDPOINT_URL=http://floci:4566
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1
```

## Adding New Stacks

```bash
pulumi stack init staging
# Copy Pulumi.dev.yaml to Pulumi.staging.yaml and update values
pulumi stack select staging
pulumi up
```

## Architecture

```
Floci Localstack (dev)          Real AWS (prod)
┌─────────────────────┐         ┌─────────────────────┐
│  ALB (port 80/443)  │         │  ALB (port 80/443)  │
│       │             │         │       │             │
│  Target Group       │         │  Target Group       │
│       │             │         │       │             │
└───────┼─────────────┘         └───────┼─────────────┘
        │                                 │
┌───────┴─────────────┐         ┌─────────┴─────────────┐
│  ECS Fargate Service│         │  ECS Fargate Service  │
│  (desiredCount: 1)  │         │  (auto-scaled 1-10)   │
│       │             │         │       │               │
│  Task Definition    │         │  Task Definition      │
│  - app container    │         │  - app container      │
│  - awslogs driver   │         │  - OTLP -> Grafana    │
└─────────────────────┘         └─────────────────────┘
```

## Key Differences: dev vs prod

| Aspect | dev (Floci) | prod (AWS) |
|--------|-------------|------------|
| Logs | CloudWatch (awslogs) | OTLP -> Grafana |
| Secrets | Environment variables | AWS Secrets Manager |
| DNS | None / local host | Route53 |
| Scaling | Disabled (not supported) | Full (1-10+) |
| Image | ghcr.io/repo:dev | ghcr.io/repo:sha/latest |

## Updating Image

The task definition uses `ghcr.io/{projectName}:{stackName}` as the image. Update by pushing a new image with that tag, then:

```bash
pulumi up  # Forces new task definition revision
```