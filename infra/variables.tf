variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "aws_access_key" {
  description = "AWS access key ID"
  type        = string
  sensitive   = true
}

variable "aws_secret_key" {
  description = "AWS secret access key"
  type        = string
  sensitive   = true
}

variable "is_localstack" {
  description = "Whether running against localstack/Floci"
  type        = bool
  default     = true
}

variable "localstack_host" {
  description = "Localstack/Floci hostname"
  type        = string
  default     = "floci"
}

variable "project_name" {
  description = "Project name"
  type        = string
  default     = "resume-analyzer"
}

variable "stack_name" {
  description = "Stack name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "image_tag" {
  description = "Docker image tag to deploy"
  type        = string
  default     = "latest"
}

variable "extra_tags" {
  description = "Extra tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# Service configuration

# Service configuration
variable "service" {
  description = "ECS service configuration"
  type = object({
    cluster_name  = string
    service_name  = string
    desired_count = number
  })
  default = {
    cluster_name  = "floci-cluster"
    service_name  = "resume-analyzer"
    desired_count = 1
  }
}

# Container configuration
variable "container" {
  description = "Container configuration"
  type = object({
    port   = number
    cpu    = string
    memory = string
  })
  default = {
    port   = 8080
    cpu    = "512"
    memory = "1024"
  }
}

# Network configuration
variable "network" {
  description = "Network configuration"
  type = object({
    vpc_id             = string
    subnet_ids         = list(string)
    security_group_ids = list(string)
  })
  default = {
    vpc_id             = "vpc-0flocilocalstack"
    subnet_ids         = ["subnet-0flocilocalstack1", "subnet-0flocilocalstack2"]
    security_group_ids = ["sg-0flocilocalstack"]
  }
}

# Health check configuration
variable "health_check" {
  description = "Health check configuration"
  type = object({
    path     = string
    interval = number
    timeout  = number
    retries  = number
  })
  default = {
    path     = "/health"
    interval = 30
    timeout  = 10
    retries  = 3
  }
}

# Load balancer configuration
variable "load_balancer" {
  description = "Load balancer configuration"
  type = object({
    internal = object({
      listener_arn  = string
      rule_priority = number
      host_header   = string
    })
  })
  default = {
    internal = {
      listener_arn  = "arn:aws:elasticloadbalancing:us-east-1:000000000000:listener/app/resume-analyzer-dev-alb/xxxxxxxx/xxxxxxxx"
      rule_priority = 10
      host_header   = "resume-analyzer.localhost"
    }
  }
}

# Environment variables (non-secret)
variable "env_vars" {
  description = "Environment variables for the container (non-secret values)"
  type        = list(object({ name = string, value = string }))
  default = [
    { name = "LLM_BASE_URL", value = "https://api.openai.com/v1" },
    { name = "LLM_MODEL", value = "gpt-4o-mini" },
    { name = "LLM_MAX_TOKENS", value = "4000" },
    { name = "LLM_TIMEOUT_SEC", value = "120" },
    { name = "LLM_RESPONSE_FORMAT", value = "json_object" },
    { name = "MAX_PDF_MB", value = "10" },
    { name = "PORT", value = "8080" },
    { name = "WORKERS", value = "2" },
    { name = "QUEUE_CAPACITY", value = "100" },
    { name = "JOB_TTL_MIN", value = "60" }
  ]
}

# Secrets configuration (references to Secrets Manager ARNs)
variable "secrets" {
  description = "Secrets to inject as environment variables (name -> secret ARN)"
  type        = map(string)
  default     = {}
}

# Secrets to create in Secrets Manager (optional - for initial setup)
variable "secrets_to_create" {
  description = "Secrets to create in Secrets Manager (name -> {name, description, value, kms_key_id, recovery_window_in_days})"
  type = map(object({
    name                    = string
    description             = optional(string, "")
    value                   = string
    kms_key_id              = optional(string, null)
    recovery_window_in_days = optional(number, 30)
  }))
  default = {}
}

# Scaling configuration
variable "scaling" {
  description = "Auto scaling configuration"
  type = object({
    enabled            = bool
    min_capacity       = number
    max_capacity       = number
    cpu_target         = number
    memory_target      = number
    scale_in_cooldown  = number
    scale_out_cooldown = number
    scheduled_actions = list(object({
      name             = string
      schedule         = string
      min_capacity     = number
      max_capacity     = number
      desired_capacity = number
      timezone         = string
    }))
  })
  default = {
    enabled            = false
    min_capacity       = 1
    max_capacity       = 3
    cpu_target         = 70
    memory_target      = 80
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
    scheduled_actions  = []
  }
}

# DNS configuration (optional)
variable "dns" {
  description = "DNS configuration"
  type = object({
    zone_name   = string
    record_name = string
  })
  default = {
    zone_name   = ""
    record_name = ""
  }
}

# IAM role ARNs (for existing roles, empty = create new)
variable "iam" {
  description = "IAM role ARNs (empty = create new)"
  type = object({
    task_role_arn      = string
    execution_role_arn = string
  })
  default = {
    task_role_arn      = ""
    execution_role_arn = ""
  }
}