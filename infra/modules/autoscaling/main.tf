# modules/autoscaling/main.tf

# Service scalable target
resource "aws_appautoscaling_target" "ecs_service" {
  count = var.enabled ? 1 : 0

  max_capacity       = var.max_capacity
  min_capacity       = var.min_capacity
  resource_id        = "service/${var.cluster_name}/${var.service_name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

# CPU scaling policy
resource "aws_appautoscaling_policy" "cpu" {
  count = var.enabled ? 1 : 0

  name               = var.cpu_policy_name
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_service[0].resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_service[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_service[0].service_namespace

  target_tracking_scaling_policy_configuration {
    target_value       = var.cpu_target
    scale_in_cooldown  = var.scale_in_cooldown
    scale_out_cooldown = var.scale_out_cooldown
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
  }
}

# Memory scaling policy
resource "aws_appautoscaling_policy" "memory" {
  count = var.enabled ? 1 : 0

  name               = var.memory_policy_name
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_service[0].resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_service[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_service[0].service_namespace

  target_tracking_scaling_policy_configuration {
    target_value       = var.memory_target
    scale_in_cooldown  = var.scale_in_cooldown
    scale_out_cooldown = var.scale_out_cooldown
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
  }
}

# Scheduled scaling actions
resource "aws_appautoscaling_scheduled_action" "scheduled" {
  for_each = { for action in var.scheduled_actions : action.name => action }

  name               = each.value.name
  service_namespace  = "ecs"
  resource_id        = "service/${var.cluster_name}/${var.service_name}"
  scalable_dimension = "ecs:service:DesiredCount"
  schedule           = each.value.schedule
  timezone           = each.value.timezone

  scalable_target_action {
    min_capacity = each.value.min_capacity
    max_capacity = each.value.max_capacity
  }
}

output "cpu_policy_arn" {
  description = "CPU scaling policy ARN"
  value       = var.enabled ? aws_appautoscaling_policy.cpu[0].arn : null
}

output "memory_policy_arn" {
  description = "Memory scaling policy ARN"
  value       = var.enabled ? aws_appautoscaling_policy.memory[0].arn : null
}