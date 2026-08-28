# modules/taskdef/variables.tf

variable "project_name" {
  type        = string
  description = "Project name"
}

variable "stack_name" {
  type        = string
  description = "Stack name"
}

variable "task_definition_family" {
  type        = string
  description = "Task definition family name"
}

variable "log_group_name" {
  type        = string
  description = "CloudWatch log group name"
}

variable "container_definitions" {
  description = "Container definitions as list of maps"
  type        = any
}

variable "cpu" {
  type        = string
  description = "CPU units (e.g., 512)"
}

variable "memory" {
  type        = string
  description = "Memory in MiB (e.g., 1024)"
}

variable "execution_role_arn" {
  type        = string
  description = "Execution role ARN"
}

variable "task_role_arn" {
  type        = string
  description = "Task role ARN"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply"
  default     = {}
}