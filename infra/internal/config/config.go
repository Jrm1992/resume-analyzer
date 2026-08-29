package config

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type ServiceConfig struct {
	ServiceName  string `json:"serviceName"`
	ClusterName  string `json:"clusterName"`
	DesiredCount int    `json:"desiredCount"`
}

type ContainerConfig struct {
	Port   int    `json:"port"`
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type NetworkConfig struct {
	VpcId           string   `json:"vpcId"`
	SubnetIds       []string `json:"subnetIds"`
	SecurityGroupId string   `json:"securityGroupId"`
}

type HealthCheckConfig struct {
	Path     string `json:"path"`
	Interval int    `json:"interval"`
	Timeout  int    `json:"timeout"`
	Retries  int    `json:"retries"`
}

type ListenerConfig struct {
	ListenerArn  string `json:"listenerArn"`
	RulePriority int    `json:"rulePriority"`
	HostHeader   string `json:"hostHeader"`
}

type LoadBalancerConfig struct {
	Internal ListenerConfig `json:"internal"`
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ScalingMetricConfig struct {
	TargetValue      float64 `json:"targetValue"`
	ScaleInCooldown  int     `json:"scaleInCooldown"`
	ScaleOutCooldown int     `json:"scaleOutCooldown"`
}

type ScalingConfig struct {
	Enabled          bool                    `json:"enabled"`
	MinCapacity      int                     `json:"minCapacity"`
	MaxCapacity      int                     `json:"maxCapacity"`
	CPU              ScalingMetricConfig     `json:"cpu"`
	Memory           ScalingMetricConfig     `json:"memory"`
	ScheduledActions []ScheduledActionConfig `json:"scheduledActions"`
}

type ScheduledActionConfig struct {
	Name            string `json:"name"`
	Schedule        string `json:"schedule"`
	MinCapacity     int    `json:"minCapacity"`
	MaxCapacity     int    `json:"maxCapacity"`
	DesiredCapacity *int   `json:"desiredCapacity,omitempty"`
	Timezone        string `json:"timezone"`
}

type DNSConfig struct {
	ZoneName   string `json:"zoneName"`
	RecordName string `json:"recordName"`
}

type IAMConfig struct {
	TaskRoleArn      string `json:"taskRoleArn"`
	ExecutionRoleArn string `json:"executionRoleArn"`
}

type StackConfig struct {
	ProjectName  string             `json:"projectName"`
	StackName    string             `json:"stackName"`
	IsLocalstack bool               `json:"isLocalstack"`
	Service      ServiceConfig      `json:"service"`
	Container    ContainerConfig    `json:"container"`
	Network      NetworkConfig      `json:"network"`
	HealthCheck  HealthCheckConfig  `json:"healthCheck"`
	LoadBalancer LoadBalancerConfig `json:"loadBalancer"`
	EnvVars      []EnvVar           `json:"envVars"`
	Scaling      ScalingConfig      `json:"scaling"`
	DNS          DNSConfig          `json:"dns"`
	IAM          IAMConfig          `json:"iam"`
}

func loadCoreObjects(cfg *config.Config, sc *StackConfig) error {
	objs := map[string]interface{}{
		"service":      &sc.Service,
		"container":    &sc.Container,
		"network":      &sc.Network,
		"healthCheck":  &sc.HealthCheck,
		"loadBalancer": &sc.LoadBalancer,
		"scaling":      &sc.Scaling,
		"dns":          &sc.DNS,
		"iam":          &sc.IAM,
	}

	for key, target := range objs {
		if err := cfg.TryObject(key, target); err != nil {
			return fmt.Errorf("config %s: %w", key, err)
		}
	}

	var envVarsRaw []EnvVar
	if err := cfg.TryObject("envVars", &envVarsRaw); err != nil {
		return fmt.Errorf("config envVars: %w", err)
	}
	sc.EnvVars = envVarsRaw

	var saRaw []ScheduledActionConfig
	if err := cfg.TryObject("scaling.scheduledActions", &saRaw); err != nil {
		return fmt.Errorf("config scaling.scheduledActions: %w", err)
	}
	sc.Scaling.ScheduledActions = saRaw

	return nil
}

func applyHealthCheckDefaults(hc *HealthCheckConfig) {
	if hc.Path == "" {
		hc.Path = "/health"
	}
	if hc.Interval == 0 {
		hc.Interval = 30
	}
	if hc.Timeout == 0 {
		hc.Timeout = 10
	}
	if hc.Retries == 0 {
		hc.Retries = 3
	}
}

func applyScalingDefaults(s *ScalingConfig) {
	if s.MinCapacity == 0 {
		s.MinCapacity = 1
	}
	if s.MaxCapacity == 0 {
		s.MaxCapacity = 3
	}
	if s.CPU.TargetValue == 0 {
		s.CPU.TargetValue = 70
	}
	if s.CPU.ScaleInCooldown == 0 {
		s.CPU.ScaleInCooldown = 300
	}
	if s.CPU.ScaleOutCooldown == 0 {
		s.CPU.ScaleOutCooldown = 60
	}
	if s.Memory.TargetValue == 0 {
		s.Memory.TargetValue = 80
	}
	if s.Memory.ScaleInCooldown == 0 {
		s.Memory.ScaleInCooldown = 300
	}
	if s.Memory.ScaleOutCooldown == 0 {
		s.Memory.ScaleOutCooldown = 60
	}
	for i := range s.ScheduledActions {
		if s.ScheduledActions[i].Timezone == "" {
			s.ScheduledActions[i].Timezone = "UTC"
		}
	}
}

func applyContainerDefaults(c *ContainerConfig) {
	if c.CPU == "" {
		c.CPU = "512"
	}
	if c.Memory == "" {
		c.Memory = "1024"
	}
}

func Load(ctx *pulumi.Context) (*StackConfig, error) {
	cfg := config.New(ctx, "")

	sc := &StackConfig{
		ProjectName:  cfg.Get("projectName"),
		StackName:    cfg.Get("stackName"),
		IsLocalstack: cfg.GetBool("isLocalstack"),
	}

	if sc.ProjectName == "" {
		sc.ProjectName = "resume-analyzer"
	}
	if sc.StackName == "" {
		sc.StackName = "dev"
	}
	if sc.StackName == "dev" && !cfg.GetBool("isLocalstack") {
		sc.IsLocalstack = true
	}

	if err := loadCoreObjects(cfg, sc); err != nil {
		return nil, err
	}

	applyHealthCheckDefaults(&sc.HealthCheck)
	applyScalingDefaults(&sc.Scaling)
	applyContainerDefaults(&sc.Container)

	return sc, nil
}
