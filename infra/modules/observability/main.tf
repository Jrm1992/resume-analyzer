locals {
  full_name = "${var.project_name}-${var.stack_name}"

  loki_s3_endpoint_override = var.is_localstack ? indent(6, "endpoint: \"${var.localstack_host}:4566\"\ninsecure: true") : ""

  loki_config = <<-YAML
    auth_enabled: false
    server:
      http_listen_port: 3100
      grpc_listen_port: 9095
    common:
      path_prefix: /loki
      replication_factor: 1
      storage:
        s3:
          bucketnames: ${aws_s3_bucket.loki_logs.bucket}
          region: ${var.aws_region}
          s3forcepathstyle: true
          ${local.loki_s3_endpoint_override}
      ring:
        kvstore:
          store: inmemory
    schema_config:
      configs:
        - from: 2024-01-01
          store: tsdb
          object_store: s3
          schema: v13
          index:
            prefix: index_
            period: 24h
    limits_config:
      retention_period: 0s
  YAML

  loki_command = "mkdir -p /etc/loki && cat > /etc/loki/local-config.yaml <<'CFG'\n${local.loki_config}\nCFG\nexec /usr/bin/loki -config.file=/etc/loki/local-config.yaml"

  grafana_datasource = <<-YAML
    apiVersion: 1
    datasources:
      - name: Loki
        type: loki
        access: proxy
        url: http://${var.loki.host_header}:${var.listener_port}
        isDefault: true
        editable: false
  YAML

  grafana_command = "mkdir -p /etc/grafana/provisioning/datasources && cat > /etc/grafana/provisioning/datasources/loki.yaml <<'CFG'\n${local.grafana_datasource}\nCFG\nexec /run.sh"

  loki_container_definitions = jsonencode([
    {
      name       = "app"
      image      = var.loki.image
      essential  = true
      entryPoint = ["/bin/sh", "-c"]
      command    = [local.loki_command]
      portMappings = [
        { containerPort = 3100, hostPort = 3100, protocol = "tcp" }
      ]
      environment = []
      secrets     = []
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = "/ecs/${local.full_name}-loki"
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  grafana_container_definitions = jsonencode([
    {
      name       = "app"
      image      = var.grafana.image
      essential  = true
      entryPoint = ["/bin/sh", "-c"]
      command    = [local.grafana_command]
      portMappings = [
        { containerPort = 3000, hostPort = 3000, protocol = "tcp" }
      ]
      environment = []
      secrets = [
        { name = "GF_SECURITY_ADMIN_PASSWORD", valueFrom = var.grafana.admin_secret_arn }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = "/ecs/${local.full_name}-grafana"
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])
}

resource "aws_s3_bucket" "loki_logs" {
  bucket        = "${local.full_name}-loki-logs"
  force_destroy = true
  tags          = var.tags
}

module "loki_iam" {
  source = "../iam"

  project_name        = var.project_name
  stack_name          = var.stack_name
  task_role_name      = "${local.full_name}-loki-task"
  execution_role_name = "${local.full_name}-loki-execution"
  tags                = var.tags

  task_policy_statements = [
    {
      sid    = "LokiS3Storage"
      effect = "Allow"
      actions = [
        "s3:ListBucket",
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
      ]
      resources = [
        aws_s3_bucket.loki_logs.arn,
        "${aws_s3_bucket.loki_logs.arn}/*",
      ]
      condition = null
    }
  ]
}

module "grafana_iam" {
  source = "../iam"

  project_name        = var.project_name
  stack_name          = var.stack_name
  task_role_name      = "${local.full_name}-grafana-task"
  execution_role_name = "${local.full_name}-grafana-execution"
  tags                = var.tags
}

module "loki_taskdef" {
  source = "../taskdef"

  project_name           = var.project_name
  stack_name             = var.stack_name
  task_definition_family = "${local.full_name}-loki"
  log_group_name         = "/ecs/${local.full_name}-loki"
  container_definitions  = local.loki_container_definitions
  cpu                    = var.loki.cpu
  memory                 = var.loki.memory
  execution_role_arn     = module.loki_iam.execution_role_arn
  task_role_arn          = module.loki_iam.task_role_arn
  tags                   = var.tags
}

module "loki_lb" {
  source = "../loadbalancer"

  project_name          = var.project_name
  stack_name            = var.stack_name
  target_group_name     = "${local.full_name}-loki"
  container_port        = 3100
  vpc_id                = var.vpc_id
  listener_arn          = var.alb_listener_arn
  rule_priority         = var.loki.rule_priority
  host_header           = var.loki.host_header
  health_check_path     = "/ready"
  health_check_interval = 30
  health_check_timeout  = 10
  health_check_retries  = 3
  tags                  = var.tags
}

module "loki_service" {
  source = "../service"

  project_name        = var.project_name
  stack_name          = var.stack_name
  service_name        = "${local.full_name}-loki"
  cluster_name        = var.cluster_name
  task_definition_arn = module.loki_taskdef.task_definition_arn
  desired_count       = 1
  subnet_ids          = var.subnet_ids
  security_group_ids  = var.security_group_ids
  container_port      = 3100
  target_group_arn    = module.loki_lb.target_group_arn
  tags                = var.tags
}

module "grafana_taskdef" {
  source = "../taskdef"

  project_name           = var.project_name
  stack_name             = var.stack_name
  task_definition_family = "${local.full_name}-grafana"
  log_group_name         = "/ecs/${local.full_name}-grafana"
  container_definitions  = local.grafana_container_definitions
  cpu                    = var.grafana.cpu
  memory                 = var.grafana.memory
  execution_role_arn     = module.grafana_iam.execution_role_arn
  task_role_arn          = module.grafana_iam.task_role_arn
  tags                   = var.tags
}

resource "aws_lb_target_group" "grafana" {
  name        = "${local.full_name}-grafana"
  port        = 3000
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    path                = "/api/health"
    interval            = 30
    timeout             = 10
    healthy_threshold   = 2
    unhealthy_threshold = 3
    matcher             = "200"
    protocol            = "HTTP"
    port                = "traffic-port"
  }

  tags = var.tags
}

resource "aws_lb_listener" "grafana" {
  load_balancer_arn = var.alb_arn
  port              = var.grafana.listener_port
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.grafana.arn
  }

  tags = var.tags
}

module "grafana_service" {
  source = "../service"

  project_name        = var.project_name
  stack_name          = var.stack_name
  service_name        = "${local.full_name}-grafana"
  cluster_name        = var.cluster_name
  task_definition_arn = module.grafana_taskdef.task_definition_arn
  desired_count       = 1
  subnet_ids          = var.subnet_ids
  security_group_ids  = var.security_group_ids
  container_port      = 3000
  target_group_arn    = aws_lb_target_group.grafana.arn
  tags                = var.tags
}
