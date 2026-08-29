resource "aws_ecs_cluster" "main" {
  name = var.cluster_name
  tags = var.tags
}

resource "aws_ecs_cluster_capacity_providers" "main" {
  count        = var.manage_capacity_providers ? 1 : 0
  cluster_name = aws_ecs_cluster.main.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 1
    capacity_provider = "FARGATE"
  }
}

output "cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.main.name
}

output "cluster_arn" {
  description = "ECS cluster ARN"
  value       = aws_ecs_cluster.main.arn
}
