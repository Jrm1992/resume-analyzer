# modules/service/main.tf

resource "aws_ecs_service" "main" {
  name             = var.service_name
  cluster          = var.cluster_name
  task_definition  = var.task_definition_arn
  desired_count    = var.desired_count
  launch_type      = "FARGATE"
  platform_version = "LATEST"
  deployment_controller {
    type = "ECS"
  }

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = var.security_group_ids
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.target_group_arn
    container_name   = "app"
    container_port   = var.container_port
  }

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50

  tags = var.tags

}


output "service_name" {
  description = "ECS service name"
  value       = aws_ecs_service.main.name
}

output "service_arn" {
  description = "ECS service ARN"
  value       = aws_ecs_service.main.id
}