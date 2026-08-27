# modules/service/variables.tf

variable "project_name" {
  type        = string
  description = "Project name"
}

variable "stack_name" {
  type        = string
  description = "Stack name"
}

variable "service_name" {
  type        = string
  description = "ECS service name"
}

variable "cluster_name" {
  type        = string
  description = "ECS cluster name"
}

variable "task_definition_arn" {
  type        = string
  description = "Task definition ARN"
}

variable "desired_count" {
  type        = number
  description = "Desired task count"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnet IDs"
}

variable "security_group_ids" {
  type        = list(string)
  description = "Security group IDs"
}

variable "container_port" {
  type        = number
  description = "Container port"
}

variable "target_group_arn" {
  type        = string
  description = "Target group ARN"
}

variable "listener_arn" {
  type        = string
  description = "ALB listener ARN"
}

variable "rule_priority" {
  type        = number
  description = "Listener rule priority"
}

variable "host_header" {
  type        = string
  description = "Host header for routing"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply"
  default     = {}
}