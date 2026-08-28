# modules/autoscaling/variables.tf

variable "project_name" {
  type        = string
  description = "Project name"
}

variable "stack_name" {
  type        = string
  description = "Stack name"
}

variable "cluster_name" {
  type        = string
  description = "ECS cluster name"
}

variable "service_name" {
  type        = string
  description = "ECS service name"
}

variable "enabled" {
  type        = bool
  description = "Enable autoscaling"
  default     = false
}

variable "min_capacity" {
  type        = number
  description = "Minimum capacity"
  default     = 1
}

variable "max_capacity" {
  type        = number
  description = "Maximum capacity"
  default     = 3
}

variable "cpu_target" {
  type        = number
  description = "CPU target utilization %"
  default     = 70
}

variable "memory_target" {
  type        = number
  description = "Memory target utilization %"
  default     = 80
}

variable "scale_in_cooldown" {
  type        = number
  description = "Scale in cooldown (seconds)"
  default     = 300
}

variable "scale_out_cooldown" {
  type        = number
  description = "Scale out cooldown (seconds)"
  default     = 60
}

variable "cpu_policy_name" {
  type        = string
  description = "CPU scaling policy name"
}

variable "memory_policy_name" {
  type        = string
  description = "Memory scaling policy name"
}

variable "scheduled_actions" {
  description = "Scheduled scaling actions"
  type = list(object({
    name             = string
    schedule         = string
    min_capacity     = number
    max_capacity     = number
    desired_capacity = number
    timezone         = string
  }))
  default = []
}