package scaling

import (
	"fmt"

	"resume-analyzer/infra/internal/config"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/appautoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Result struct {
	ScalableTarget   *appautoscaling.Target
	CPUPolicy        *appautoscaling.Policy
	MemoryPolicy     *appautoscaling.Policy
	ScheduledActions []*appautoscaling.ScheduledAction
}

type Input struct {
	Cfg     *config.StackConfig
	Service *ecs.Service
	Opts    []pulumi.ResourceOption
}

func New(ctx *pulumi.Context, in *Input) (*Result, error) {
	cfg := in.Cfg
	opts := in.Opts
	sc := cfg.Scaling

	resourceID := pulumi.All(in.Service.Cluster, in.Service.Name).ApplyT(
		func(args []interface{}) string {
			return fmt.Sprintf("service/%s/%s", args[0], args[1])
		},
	).(pulumi.StringOutput)

	scalableTarget, err := appautoscaling.NewTarget(ctx, "ecs-scalable-target", &appautoscaling.TargetArgs{
		MaxCapacity:       pulumi.Int(sc.MaxCapacity),
		MinCapacity:       pulumi.Int(sc.MinCapacity),
		ResourceId:        resourceID,
		ScalableDimension: pulumi.String("ecs:service:DesiredCount"),
		ServiceNamespace:  pulumi.String("ecs"),
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar scalable target: %w", err)
	}

	cpuPolicy, err := appautoscaling.NewPolicy(ctx, "ecs-cpu-scaling-policy", &appautoscaling.PolicyArgs{
		PolicyType:        pulumi.String("TargetTrackingScaling"),
		ResourceId:        resourceID,
		ServiceNamespace:  pulumi.String("ecs"),
		ScalableDimension: pulumi.String("ecs:service:DesiredCount"),
		TargetTrackingScalingPolicyConfiguration: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationArgs{
			TargetValue: pulumi.Float64(sc.CPU.TargetValue),
			PredefinedMetricSpecification: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationPredefinedMetricSpecificationArgs{
				PredefinedMetricType: pulumi.String("ECSServiceAverageCPUUtilization"),
			},
			ScaleInCooldown:  pulumi.Int(sc.CPU.ScaleInCooldown),
			ScaleOutCooldown: pulumi.Int(sc.CPU.ScaleOutCooldown),
		},
	}, append(opts, pulumi.DependsOn([]pulumi.Resource{scalableTarget}))...)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar política de CPU scaling: %w", err)
	}

	memoryPolicy, err := appautoscaling.NewPolicy(ctx, "ecs-memory-scaling-policy", &appautoscaling.PolicyArgs{
		PolicyType:        pulumi.String("TargetTrackingScaling"),
		ResourceId:        resourceID,
		ServiceNamespace:  pulumi.String("ecs"),
		ScalableDimension: pulumi.String("ecs:service:DesiredCount"),
		TargetTrackingScalingPolicyConfiguration: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationArgs{
			TargetValue: pulumi.Float64(sc.Memory.TargetValue),
			PredefinedMetricSpecification: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationPredefinedMetricSpecificationArgs{
				PredefinedMetricType: pulumi.String("ECSServiceAverageMemoryUtilization"),
			},
			ScaleInCooldown:  pulumi.Int(sc.Memory.ScaleInCooldown),
			ScaleOutCooldown: pulumi.Int(sc.Memory.ScaleOutCooldown),
		},
	}, append(opts, pulumi.DependsOn([]pulumi.Resource{scalableTarget}))...)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar política de Memory scaling: %w", err)
	}

	scheduledActions, err := newScheduledActions(ctx, cfg, resourceID, opts...)
	if err != nil {
		return nil, err
	}

	return &Result{
		ScalableTarget:   scalableTarget,
		CPUPolicy:        cpuPolicy,
		MemoryPolicy:     memoryPolicy,
		ScheduledActions: scheduledActions,
	}, nil
}

func newScheduledActions(
	ctx *pulumi.Context,
	cfg *config.StackConfig,
	resourceID pulumi.StringOutput,
	opts ...pulumi.ResourceOption,
) ([]*appautoscaling.ScheduledAction, error) {
	var actions []*appautoscaling.ScheduledAction

	for _, sa := range cfg.Scaling.ScheduledActions {
		args := &appautoscaling.ScheduledActionArgs{
			Name:               pulumi.String(fmt.Sprintf("%s-%s-%s", cfg.ProjectName, cfg.StackName, sa.Name)),
			ResourceId:         resourceID,
			ServiceNamespace:   pulumi.String("ecs"),
			ScalableDimension:  pulumi.String("ecs:service:DesiredCount"),
			Schedule:           pulumi.String(sa.Schedule),
			Timezone:           pulumi.String(sa.Timezone),
			ScalableTargetAction: &appautoscaling.ScheduledActionScalableTargetActionArgs{
				MinCapacity: pulumi.Int(sa.MinCapacity),
				MaxCapacity: pulumi.Int(sa.MaxCapacity),
			},
		}

		action, err := appautoscaling.NewScheduledAction(ctx, fmt.Sprintf("scheduled-%s", sa.Name), args, opts...)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar scheduled action %s: %w", sa.Name, err)
		}
		actions = append(actions, action)
	}

	return actions, nil
}