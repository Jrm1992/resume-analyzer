package main

import (
	"resume-analyzer/infra/internal/config"
	"resume-analyzer/infra/internal/iam"
	"resume-analyzer/infra/internal/loadbalancer"
	"resume-analyzer/infra/internal/scaling"
	"resume-analyzer/infra/internal/service"
	"resume-analyzer/infra/internal/taskdef"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg, err := config.Load(ctx)
		if err != nil {
			return err
		}

		var provider *aws.Provider
		if cfg.IsLocalstack {
			provider, err = aws.NewProvider(ctx, "aws", &aws.ProviderArgs{
				Region: pulumi.String("us-east-1"),
				Endpoints: aws.ProviderEndpointArray{
					aws.ProviderEndpointArgs{
						Ecs:                    pulumi.String("http://floci:4566"),
						Iam:                    pulumi.String("http://floci:4566"),
						Cloudwatch:             pulumi.String("http://floci:4566"),
						Elasticloadbalancing:   pulumi.String("http://floci:4566"),
						Route53:                pulumi.String("http://floci:4566"),
						Ec2:                    pulumi.String("http://floci:4566"),
						Logs:                   pulumi.String("http://floci:4566"),
						Applicationautoscaling: pulumi.String("http://floci:4566"),
					},
				},
				SkipCredentialsValidation: pulumi.Bool(true),
				SkipMetadataApiCheck:      pulumi.Bool(true),
				SkipRequestingAccountId:   pulumi.Bool(true),
				S3UsePathStyle:            pulumi.Bool(true),
			})
			if err != nil {
				return err
			}
		}

		opts := pulumi.Provider(provider)

		iamResult, err := iam.New(ctx, cfg, opts)
		if err != nil {
			return err
		}

		tdResult, err := taskdef.New(ctx, &taskdef.Input{
			Cfg:           cfg,
			ExecutionRole: iamResult.ExecutionRole,
			TaskRole:      iamResult.TaskRole,
			Opts:          []pulumi.ResourceOption{opts},
		})
		if err != nil {
			return err
		}

		lbResult, err := loadbalancer.New(ctx, cfg, opts)
		if err != nil {
			return err
		}

		svcResult, err := service.New(ctx, &service.Input{
			Cfg:                 cfg,
			TaskDefinitionArn:   tdResult.TaskDefinition.Arn,
			InternalTargetGroup: lbResult.InternalTargetGroup,
			Opts:                []pulumi.ResourceOption{opts},
		})
		if err != nil {
			return err
		}

		// Skip auto-scaling on localstack (not supported)
		if !cfg.IsLocalstack {
			_, err = scaling.New(ctx, &scaling.Input{
				Cfg:     cfg,
				Service: svcResult.Service,
				Opts:    []pulumi.ResourceOption{opts},
			})
			if err != nil {
				return err
			}
		}

		if cfg.DNS.ZoneName != "" && cfg.DNS.RecordName != "" {
			if err := provisionDNS(ctx, cfg, opts); err != nil {
				return err
			}
		}

		exportOutputs(ctx, iamResult, tdResult, lbResult, svcResult)
		return nil
	})
}

func provisionDNS(ctx *pulumi.Context, cfg *config.StackConfig, opts pulumi.ResourceOption) error {
	zone, err := route53.LookupZone(ctx, &route53.LookupZoneArgs{Name: pulumi.StringRef(cfg.DNS.ZoneName)}, nil)
	if err != nil {
		return err
	}

	_, err = route53.NewRecord(ctx, "dns-record", &route53.RecordArgs{
		ZoneId: pulumi.String(zone.ZoneId),
		Name:   pulumi.String(cfg.DNS.RecordName),
		Type:   pulumi.String("CNAME"),
		Ttl:    pulumi.Int(60),
		Records: pulumi.StringArray{
			pulumi.String("internal-alb-placeholder"),
		},
	}, opts)
	return err
}

func exportOutputs(
	ctx *pulumi.Context,
	iamResult *iam.Result,
	tdResult *taskdef.Result,
	lbResult *loadbalancer.Result,
	svcResult *service.Result,
) {
	ctx.Export("taskRoleArn", iamResult.TaskRole.Arn)
	ctx.Export("executionRoleArn", iamResult.ExecutionRole.Arn)
	ctx.Export("taskDefinitionArn", tdResult.TaskDefinition.Arn)
	ctx.Export("logGroupName", tdResult.LogGroup.Name)
	ctx.Export("targetGroupArn", lbResult.InternalTargetGroup.Arn)
	ctx.Export("serviceName", svcResult.Service.Name)
	ctx.Export("serviceArn", svcResult.Service.Arn)
}
