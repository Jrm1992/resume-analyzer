package service

import (
	"fmt"

	"resume-analyzer/infra/internal/config"
	"resume-analyzer/infra/internal/tags"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Input struct {
	Cfg                 *config.StackConfig
	TaskDefinitionArn   pulumi.StringOutput
	InternalTargetGroup *lb.TargetGroup
	Opts                []pulumi.ResourceOption
}

type Result struct {
	Service *ecs.Service
}

func New(ctx *pulumi.Context, in *Input) (*Result, error) {
	cfg := in.Cfg
	opts := in.Opts

	if cfg.Network.SecurityGroupId == "" {
		return nil, fmt.Errorf("network.securityGroupId (ALB SG) is required")
	}

	subnetPulumi := make(pulumi.StringArray, len(cfg.Network.SubnetIds))
	for i, s := range cfg.Network.SubnetIds {
		subnetPulumi[i] = pulumi.String(s)
	}

	svc, err := ecs.NewService(ctx, "ecs-service", &ecs.ServiceArgs{
		Name:                 pulumi.String(cfg.Service.ServiceName),
		Cluster:              pulumi.String(cfg.Service.ClusterName),
		TaskDefinition:       in.TaskDefinitionArn,
		DesiredCount:         pulumi.Int(cfg.Service.DesiredCount),
		LaunchType:           pulumi.String("FARGATE"),
		EnableEcsManagedTags: pulumi.Bool(true),
		NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
			AssignPublicIp: pulumi.Bool(false),
			Subnets:        subnetPulumi,
			SecurityGroups: pulumi.StringArray{pulumi.String(cfg.Network.SecurityGroupId)},
		},
		LoadBalancers: ecs.ServiceLoadBalancerArray{
			&ecs.ServiceLoadBalancerArgs{
				TargetGroupArn: in.InternalTargetGroup.Arn,
				ContainerName:  pulumi.String("app"),
				ContainerPort:  pulumi.Int(cfg.Container.Port),
			},
		},
		DeploymentCircuitBreaker: &ecs.ServiceDeploymentCircuitBreakerArgs{
			Enable:   pulumi.Bool(true),
			Rollback: pulumi.Bool(true),
		},
		DeploymentMinimumHealthyPercent: pulumi.Int(100),
		DeploymentMaximumPercent:        pulumi.Int(200),
		WaitForSteadyState:              pulumi.Bool(true),
		Tags:                            tags.Base(cfg.ProjectName, cfg.StackName),
	}, opts...)
	if err != nil {
		return nil, err
	}

	return &Result{Service: svc}, nil
}