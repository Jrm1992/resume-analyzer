package loadbalancer

import (
	"fmt"

	"resume-analyzer/infra/internal/config"
	"resume-analyzer/infra/internal/tags"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Result struct {
	InternalTargetGroup *lb.TargetGroup
}

func New(ctx *pulumi.Context, cfg *config.StackConfig, opts ...pulumi.ResourceOption) (*Result, error) {
	internalTG, err := newTargetGroup(ctx, cfg, "internal", opts...)
	if err != nil {
		return nil, err
	}

	if err := newListenerRule(ctx, cfg, "internal", cfg.LoadBalancer.Internal, internalTG, opts...); err != nil {
		return nil, err
	}

	return &Result{InternalTargetGroup: internalTG}, nil
}

func newTargetGroup(ctx *pulumi.Context, cfg *config.StackConfig, lbType string, opts ...pulumi.ResourceOption) (*lb.TargetGroup, error) {
	hc := cfg.HealthCheck

	name := fmt.Sprintf("%s-%s-%s", cfg.ProjectName, cfg.StackName, lbType[:3])
	if len(name) > 32 {
		name = name[:32]
	}

	tg, err := lb.NewTargetGroup(ctx, fmt.Sprintf("%s-tg", lbType), &lb.TargetGroupArgs{
		Name:                pulumi.String(name),
		Port:                pulumi.Int(cfg.Container.Port),
		Protocol:            pulumi.String("HTTP"),
		TargetType:          pulumi.String("ip"),
		VpcId:               pulumi.String(cfg.Network.VpcId),
		DeregistrationDelay: pulumi.Int(30),
		HealthCheck: &lb.TargetGroupHealthCheckArgs{
			Path:               pulumi.String(hc.Path),
			Interval:           pulumi.Int(hc.Interval),
			Timeout:            pulumi.Int(hc.Timeout),
			HealthyThreshold:   pulumi.Int(2),
			UnhealthyThreshold: pulumi.Int(hc.Retries),
			Matcher:            pulumi.String("200-299"),
		},
		Tags: tags.Base(cfg.ProjectName, cfg.StackName, map[string]string{
			"Type": lbType,
		}),
	}, opts...)
	if err != nil {
		return nil, err
	}

	return tg, nil
}

func newListenerRule(
	ctx *pulumi.Context,
	cfg *config.StackConfig,
	lbType string,
	listenerCfg config.ListenerConfig,
	tg *lb.TargetGroup,
	opts ...pulumi.ResourceOption,
) error {
	_, err := lb.NewListenerRule(ctx, fmt.Sprintf("%s-listener-rule", lbType), &lb.ListenerRuleArgs{
		ListenerArn: pulumi.String(listenerCfg.ListenerArn),
		Priority:    pulumi.Int(listenerCfg.RulePriority),
		Actions: lb.ListenerRuleActionArray{
			&lb.ListenerRuleActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: tg.Arn,
			},
		},
		Conditions: lb.ListenerRuleConditionArray{
			&lb.ListenerRuleConditionArgs{
				HostHeader: &lb.ListenerRuleConditionHostHeaderArgs{
					Values: pulumi.StringArray{pulumi.String(listenerCfg.HostHeader)},
				},
			},
		},
		Tags: tags.Base(cfg.ProjectName, cfg.StackName, map[string]string{
			"Type": lbType,
		}),
	}, opts...)
	return err
}