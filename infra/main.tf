provider "aws" {
  region                      = var.aws_region
  access_key                  = var.aws_access_key
  secret_key                  = var.aws_secret_key
  skip_credentials_validation = var.is_localstack
  skip_metadata_api_check     = var.is_localstack
  skip_requesting_account_id  = var.is_localstack

  endpoints {
    ec2                    = var.is_localstack ? "http://${var.localstack_host}:4566" : null
    ecs                    = var.is_localstack ? "http://${var.localstack_host}:4566" : null
    elasticloadbalancing   = var.is_localstack ? "http://${var.localstack_host}:4566" : null
    iam                    = var.is_localstack ? "http://${var.localstack_host}:4566" : null
    logs                   = var.is_localstack ? "http://${var.localstack_host}:4566" : null
    secretsmanager         = var.is_localstack ? "http://${var.localstack_host}:4566" : null
    s3                     = var.is_localstack ? "http://${var.localstack_host}:4566" : null
    sts                    = var.is_localstack ? "http://${var.localstack_host}:4566" : null
    applicationautoscaling = var.is_localstack ? "http://${var.localstack_host}:4566" : null
  }
}
data "aws_region" "current" {}

locals {
  project_name = var.project_name
  stack_name   = var.stack_name
  full_name    = "${var.project_name}-${var.stack_name}"
  tags = merge(
    {
      Project     = var.project_name
      Stack       = var.stack_name
      ManagedBy   = "terraform"
      Environment = var.stack_name
    },
    var.extra_tags
  )

  # IAM role names
  task_role_name      = "${local.full_name}-task"
  execution_role_name = "${local.full_name}-execution"

  # Scaling policy names
  cpu_policy_name    = "${local.full_name}-cpu-scaling"
  memory_policy_name = "${local.full_name}-memory-scaling"

  # Target group name
  target_group_name = local.full_name

  # Whether to create IAM roles (vs use existing)
  create_iam_roles = var.iam.task_role_arn == "" && var.iam.execution_role_arn == ""
}

# ============================================================
# IAM Module
# ============================================================
module "iam" {
  source = "./modules/iam"

  project_name                = var.project_name
  stack_name                  = var.stack_name
  task_role_name              = local.task_role_name
  execution_role_name         = local.execution_role_name
  tags                        = local.tags
  create_roles                = local.create_iam_roles
  existing_task_role_arn      = var.iam.task_role_arn
  existing_execution_role_arn = var.iam.execution_role_arn
}

# ============================================================
# Secrets Module (optional - creates secrets in Secrets Manager)
# ============================================================
module "secrets" {
  source       = "./modules/secrets"
  project_name = var.project_name
  stack_name   = var.stack_name
  secrets      = var.secrets_to_create
  tags         = local.tags
}

# ============================================================
# Task Definition Module
# ============================================================
module "taskdef" {
  source = "./modules/taskdef"

  project_name           = var.project_name
  stack_name             = var.stack_name
  task_definition_family = local.full_name
  log_group_name         = "/ecs/${local.full_name}"
  container_definitions  = local.container_definitions
  cpu                    = var.container.cpu
  memory                 = var.container.memory
  execution_role_arn     = module.iam.execution_role_arn
  task_role_arn          = module.iam.task_role_arn
  tags                   = local.tags
}

# ============================================================
# Load Balancer Module
# ============================================================
module "loadbalancer" {
  source = "./modules/loadbalancer"

  project_name          = var.project_name
  stack_name            = var.stack_name
  target_group_name     = local.target_group_name
  container_port        = var.container.port
  vpc_id                = var.network.vpc_id
  listener_arn          = var.load_balancer.internal.listener_arn
  rule_priority         = var.load_balancer.internal.rule_priority
  host_header           = var.load_balancer.internal.host_header
  health_check_path     = var.health_check.path
  health_check_interval = var.health_check.interval
  health_check_timeout  = var.health_check.timeout
  health_check_retries  = var.health_check.retries
  tags                  = local.tags
}

# ============================================================
# ECS Service Module
# ============================================================
module "service" {
  source = "./modules/service"

  project_name        = var.project_name
  stack_name          = var.stack_name
  service_name        = var.service.service_name
  cluster_name        = var.service.cluster_name
  task_definition_arn = module.taskdef.task_definition_arn
  desired_count       = var.service.desired_count
  subnet_ids          = var.network.subnet_ids
  security_group_ids  = var.network.security_group_ids
  container_port      = var.container.port
  target_group_arn    = module.loadbalancer.target_group_arn
  listener_arn        = var.load_balancer.internal.listener_arn
  rule_priority       = var.load_balancer.internal.rule_priority
  host_header         = var.load_balancer.internal.host_header
  tags                = local.tags
}

# ============================================================
# Autoscaling Module
# ============================================================
module "autoscaling" {
  source = "./modules/autoscaling"

  project_name       = var.project_name
  stack_name         = var.stack_name
  cluster_name       = var.service.cluster_name
  service_name       = var.service.service_name
  enabled            = var.scaling.enabled
  min_capacity       = var.scaling.min_capacity
  max_capacity       = var.scaling.max_capacity
  cpu_target         = var.scaling.cpu_target
  memory_target      = var.scaling.memory_target
  scale_in_cooldown  = var.scaling.scale_in_cooldown
  scale_out_cooldown = var.scaling.scale_out_cooldown
  cpu_policy_name    = local.cpu_policy_name
  memory_policy_name = local.memory_policy_name
  scheduled_actions  = var.scaling.scheduled_actions
}

# ============================================================
# Locals for container definitions
# ============================================================
locals {
  # Build secrets list with optional config_secret_arn
  secrets_list = concat(
    [for name, arn in var.secrets : { name = name, valueFrom = arn }],
    var.config_secret_arn != "" ? [{ name = "CONFIG_SECRET", valueFrom = var.config_secret_arn }] : []
  )

  container_definitions = jsonencode([
    {
      portMappings = [
        {
          containerPort = var.container.port
          hostPort      = var.container.port
          protocol      = "tcp"
        }
      ]
      environment = [
        for env in var.env_vars : {
          name  = env.name
          value = env.value
        }
      ]
      secrets = local.secrets_list
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = "/ecs/${local.full_name}"
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:${var.container.port}${var.health_check.path} || exit 1"]
        interval    = var.health_check.interval
        timeout     = var.health_check.timeout
        retries     = var.health_check.retries
        startPeriod = 10
      }
    }
  ])
}