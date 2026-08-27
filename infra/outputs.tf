output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = var.service.cluster_name
}

output "ecs_service_name" {
  description = "ECS service name"
  value       = module.service.service_name
}

output "task_definition_arn" {
  description = "Task definition ARN"
  value       = module.taskdef.task_definition_arn
}

output "task_definition_family" {
  description = "Task definition family"
  value       = module.taskdef.task_definition_family
}

output "task_role_arn" {
  description = "Task role ARN"
  value       = module.iam.task_role_arn
}

output "execution_role_arn" {
  description = "Execution role ARN"
  value       = module.iam.execution_role_arn
}

output "target_group_arn" {
  description = "ALB target group ARN"
  value       = module.loadbalancer.target_group_arn
}

output "listener_rule_arn" {
  description = "ALB listener rule ARN"
  value       = module.loadbalancer.listener_rule_arn
}

output "log_group_name" {
  description = "CloudWatch log group name"
  value       = module.taskdef.log_group_name
}

output "service_url" {
  description = "Service URL (if DNS configured)"
  value       = var.dns.record_name != "" ? "https://${var.dns.record_name}.${var.dns.zone_name}" : null
}