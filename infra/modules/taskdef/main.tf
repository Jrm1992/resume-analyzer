# modules/taskdef/main.tf

resource "aws_cloudwatch_log_group" "main" {
  name              = var.log_group_name
  retention_in_days = 30
  tags              = var.tags
}

resource "aws_ecs_task_definition" "main" {
  family                   = var.task_definition_family
  container_definitions    = var.container_definitions
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory
  execution_role_arn       = var.execution_role_arn
  task_role_arn            = var.task_role_arn
  tags                     = var.tags
}

output "task_definition_arn" {
  description = "Task definition ARN"
  value       = aws_ecs_task_definition.main.arn
}

output "task_definition_family" {
  description = "Task definition family"
  value       = aws_ecs_task_definition.main.family
}

output "task_definition_revision" {
  description = "Task definition revision"
  value       = aws_ecs_task_definition.main.revision
}

output "log_group_name" {
  description = "CloudWatch log group name"
  value       = aws_cloudwatch_log_group.main.name
}