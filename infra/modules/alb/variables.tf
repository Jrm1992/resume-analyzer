# modules/alb/variables.tf

variable "project_name" {
  type        = string
  description = "Project name"
}

variable "stack_name" {
  type        = string
  description = "Stack name"
}

variable "alb_name" {
  type        = string
  description = "ALB name"
}

variable "internal" {
  type        = bool
  description = "Whether ALB is internal"
  default     = true
}

variable "vpc_id" {
  type        = string
  description = "VPC ID"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnet IDs"
}

variable "security_group_ids" {
  type        = list(string)
  description = "Security group IDs"
}

variable "listener_port" {
  type        = number
  description = "Listener port"
  default     = 80
}

variable "listener_protocol" {
  type        = string
  description = "Listener protocol"
  default     = "HTTP"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply"
  default     = {}
}