# modules/loadbalancer/variables.tf

variable "project_name" {
  type        = string
  description = "Project name"
}

variable "stack_name" {
  type        = string
  description = "Stack name"
}

variable "target_group_name" {
  type        = string
  description = "Target group name"
}

variable "container_port" {
  type        = number
  description = "Container port"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID"
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

variable "health_check_path" {
  type        = string
  description = "Health check path"
}

variable "health_check_interval" {
  type        = number
  description = "Health check interval (seconds)"
}

variable "health_check_timeout" {
  type        = number
  description = "Health check timeout (seconds)"
}

variable "health_check_retries" {
  type        = number
  description = "Health check retries"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply"
  default     = {}
}