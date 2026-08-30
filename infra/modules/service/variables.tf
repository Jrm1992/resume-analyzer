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

variable "container_name" {
  type        = string
  description = "Name of the container (in the task definition) to attach to the load balancer"
  default     = "app"
}


variable "tags" {
  type        = map(string)
  description = "Tags to apply"
  default     = {}
}