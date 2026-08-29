package taskdef

import (
	"encoding/json"
	"fmt"

	"resume-analyzer/infra/internal/config"
	"resume-analyzer/infra/internal/tags"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecs"
	awsiam "github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Input struct {
	Cfg           *config.StackConfig
	ExecutionRole *awsiam.Role
	TaskRole      *awsiam.Role
	Opts          []pulumi.ResourceOption
}

type Result struct {
	TaskDefinition *ecs.TaskDefinition
	LogGroup       *cloudwatch.LogGroup
}

func New(ctx *pulumi.Context, in *Input) (*Result, error) {
	cfg := in.Cfg
	opts := in.Opts

	logGroupName := fmt.Sprintf("/ecs/%s-%s", cfg.ProjectName, cfg.StackName)
	logGroup, err := cloudwatch.NewLogGroup(ctx, "log-group", &cloudwatch.LogGroupArgs{
		Name: pulumi.String(logGroupName),
	}, opts...)
	if err != nil {
		return nil, err
	}

	envList := make([]map[string]string, len(cfg.EnvVars))
	for i, e := range cfg.EnvVars {
		envList[i] = map[string]string{
			"name":  e.Name,
			"value": e.Value,
		}
	}
	envJSON, err := json.Marshal(envList)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar envVars: %w", err)
	}

	portMappingsJSON, err := json.Marshal([]map[string]int{
		{"containerPort": cfg.Container.Port, "hostPort": cfg.Container.Port, "protocol": 6},
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar portMappings: %w", err)
	}

	def := map[string]any{
		"name":         "app",
		"image":        pulumi.Sprintf("ghcr.io/%s:%s", cfg.ProjectName, cfg.StackName),
		"essential":    true,
		"portMappings": json.RawMessage(portMappingsJSON),
		"environment":  json.RawMessage(envJSON),
		"logConfiguration": map[string]any{
			"logDriver": "awslogs",
			"options": map[string]string{
				"awslogs-group":         logGroupName,
				"awslogs-region":        "us-east-1",
				"awslogs-stream-prefix": "ecs",
			},
		},
		"healthCheck": map[string]any{
			"command":     []string{"CMD-SHELL", fmt.Sprintf("curl -f http://localhost:%d%s || exit 1", cfg.Container.Port, cfg.HealthCheck.Path)},
			"interval":    cfg.HealthCheck.Interval,
			"timeout":     cfg.HealthCheck.Timeout,
			"retries":     cfg.HealthCheck.Retries,
			"startPeriod": 10,
		},
	}

	containerDefJSON, err := json.Marshal([]any{def})
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar container definition: %w", err)
	}

	taskDef, err := ecs.NewTaskDefinition(ctx, "task-definition", &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String(fmt.Sprintf("%s-%s", cfg.ProjectName, cfg.StackName)),
		ContainerDefinitions:    pulumi.String(containerDefJSON),
		NetworkMode:             pulumi.String("awsvpc"),
		RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
		Cpu:                     pulumi.String(cfg.Container.CPU),
		Memory:                  pulumi.String(cfg.Container.Memory),
		ExecutionRoleArn:        in.ExecutionRole.Arn,
		TaskRoleArn:             in.TaskRole.Arn,
		Tags:                    tags.Base(cfg.ProjectName, cfg.StackName),
	}, append(opts, pulumi.DependsOn([]pulumi.Resource{logGroup}))...)
	if err != nil {
		return nil, err
	}

	return &Result{
		TaskDefinition: taskDef,
		LogGroup:       logGroup,
	}, nil
}
