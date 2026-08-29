package iam

import (
	"fmt"

	"resume-analyzer/infra/internal/config"
	"resume-analyzer/infra/internal/tags"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Result struct {
	TaskRole      *iam.Role
	ExecutionRole *iam.Role
}

func New(ctx *pulumi.Context, cfg *config.StackConfig, opts ...pulumi.ResourceOption) (*Result, error) {
	taskRole, err := newTaskRole(ctx, cfg, opts...)
	if err != nil {
		return nil, fmt.Errorf("create task role: %w", err)
	}

	executionRole, err := newExecutionRole(ctx, cfg, opts...)
	if err != nil {
		return nil, fmt.Errorf("create execution role: %w", err)
	}

	return &Result{
		TaskRole:      taskRole,
		ExecutionRole: executionRole,
	}, nil
}

func newTaskRole(ctx *pulumi.Context, cfg *config.StackConfig, opts ...pulumi.ResourceOption) (*iam.Role, error) {
	assumeRolePolicy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": { "Service": "ecs-tasks.amazonaws.com" },
			"Action": "sts:AssumeRole"
		}]
	}`

	role, err := iam.NewRole(ctx, "task-role", &iam.RoleArgs{
		Name:             pulumi.String(fmt.Sprintf("%s-%s-task", cfg.ProjectName, cfg.StackName)),
		AssumeRolePolicy: pulumi.String(assumeRolePolicy),
		Description:      pulumi.String(fmt.Sprintf("Task role for %s %s", cfg.ProjectName, cfg.StackName)),
		Tags:             tags.Base(cfg.ProjectName, cfg.StackName, map[string]string{"Role": "task"}),
	}, opts...)
	if err != nil {
		return nil, err
	}

	policyDoc := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Action": [
				"secretsmanager:GetSecretValue",
				"ssm:GetParameter",
				"ssm:GetParameters"
			],
			"Resource": "*"
		}]
	}`

	_, err = iam.NewRolePolicy(ctx, "task-policy", &iam.RolePolicyArgs{
		Role:   role.Name,
		Policy: pulumi.String(policyDoc),
	}, opts...)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func newExecutionRole(ctx *pulumi.Context, cfg *config.StackConfig, opts ...pulumi.ResourceOption) (*iam.Role, error) {
	assumeRolePolicy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": { "Service": "ecs-tasks.amazonaws.com" },
			"Action": "sts:AssumeRole"
		}]
	}`

	role, err := iam.NewRole(ctx, "execution-role", &iam.RoleArgs{
		Name:             pulumi.String(fmt.Sprintf("%s-%s-execution", cfg.ProjectName, cfg.StackName)),
		AssumeRolePolicy: pulumi.String(assumeRolePolicy),
		Description:      pulumi.String(fmt.Sprintf("Execution role for %s %s", cfg.ProjectName, cfg.StackName)),
		Tags:             tags.Base(cfg.ProjectName, cfg.StackName, map[string]string{"Role": "execution"}),
	}, opts...)
	if err != nil {
		return nil, err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, "execution-policy-attachment", &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
	}, opts...)
	if err != nil {
		return nil, err
	}

	return role, nil
}
