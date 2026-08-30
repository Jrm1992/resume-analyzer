variable "project_name" {
  type        = string
  description = "Project name"
}

variable "stack_name" {
  type        = string
  description = "Stack name"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply"
  default     = {}
}

variable "cluster_name" {
  type        = string
  description = "ECS cluster name to deploy Loki and Grafana into"
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
  description = "Security group IDs (shared with the app service and ALB)"
}

variable "alb_listener_arn" {
  type        = string
  description = "ALB listener ARN to attach host-header routing rules to"
}

variable "listener_port" {
  type        = number
  description = "ALB listener port (used to build the Grafana/Loki URLs)"
}

variable "aws_region" {
  type        = string
  description = "AWS region"
}

variable "is_localstack" {
  type        = bool
  description = "Whether running against LocalStack/Floci"
}

variable "localstack_host" {
  type        = string
  description = "LocalStack/Floci hostname"
}

variable "loki" {
  description = "Loki container + routing configuration"
  type = object({
    image         = string
    cpu           = string
    memory        = string
    rule_priority = number
    host_header   = string
  })
}

variable "grafana" {
  description = "Grafana container + routing configuration"
  type = object({
    image            = string
    cpu              = string
    memory           = string
    rule_priority    = number
    host_header      = string
    admin_secret_arn = string
  })

  validation {
    condition     = length(var.grafana.admin_secret_arn) > 0
    error_message = "grafana.admin_secret_arn (Secrets Manager ARN holding GF_SECURITY_ADMIN_PASSWORD) is required."
  }
}
