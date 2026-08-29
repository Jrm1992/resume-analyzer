# Infrastructure as Code for resume-analyzer

Terraform modules for deploying resume-analyzer to ECS Fargate. Production runs on Floci (localstack), so every module works with both real AWS and LocalStack endpoints.

## Structure

```
infra/
├── main.tf                  # Root module: wires all modules together
├── variables.tf             # All stack inputs
├── outputs.tf               # Stack outputs
├── versions.tf              # Terraform + provider requirements, backend
├── environments/
│   └── prod.tfvars           # Floci (localstack) production values
└── modules/
    ├── iam/                  # Task + execution IAM roles
    ├── taskdef/              # ECS task definition (container defs, secrets)
    ├── cluster/              # ECS cluster
    ├── service/              # ECS Fargate service
    ├── alb/                  # Application Load Balancer + listener
    ├── loadbalancer/         # Target group + listener rule
    ├── autoscaling/          # Target-tracking CPU/memory scaling
    └── secrets/              # Secrets Manager secrets (optional)
```

## Deploy pipeline

Deploys happen via `.github/workflows/deploy.yml` on push to `main`:

1. Build & push multi-arch Docker image to `ghcr.io` (tagged with git SHA)
2. Runner (self-hosted, `floci-runner`) runs `terraform init/plan/apply` with `environments/prod.tfvars` and `image_tag` from the build job
3. State is persisted on the runner VM (`tfstate/terraform.tfstate`)

`.github/workflows/terraform-deploy.yml` runs `fmt/validate/plan` on every PR as a dry-run.

## Manual apply (on the Floci VM)

```bash
cd infra
terraform init
terraform plan -var-file=environments/prod.tfvars -var image_tag=<tag>
terraform apply -var-file=environments/prod.tfvars -var image_tag=<tag>
```

LocalStack credentials (any values work): `AWS_ACCESS_KEY_ID=test`, `AWS_SECRET_ACCESS_KEY=test`, `AWS_DEFAULT_REGION=us-east-1`.

## Configuration

All inputs are in `variables.tf`; per-environment values in `environments/prod.tfvars`. Key variables:

| Variable | Purpose |
|----------|---------|
| `image_repo` / `image_tag` | Container image (`ghcr.io/jrm1992/resume-analyzer:<tag>`) |
| `service` | ECS cluster name, service name, desired count |
| `container` | Port, CPU, memory for the Fargate task |
| `network` | VPC, subnets, security groups (LocalStack defaults in prod.tfvars) |
| `health_check` | ALB target group health check |
| `load_balancer` | Listener rule priority + host header |
| `env_vars` | Container environment variables |
| `config_secret_arn` | Secrets Manager ARN for the JSON config secret (LLM_API_KEY) |
| `scaling` | Auto-scaling (min/max, CPU/memory targets) |
| `dns` | Optional Route53 record |
| `iam` | Optional pre-existing role ARNs |

## Adding a new environment

```bash
# Copy environments/prod.tfvars to environments/staging.tfvars and edit values
terraform plan -var-file=environments/staging.tfvars -var image_tag=<tag>
```

Backend is `local` (state file on the VM). Moving to real AWS: uncomment the `s3` backend in `versions.tf`.
