# modules/secrets/variables.tf

variable "project_name" {
  type        = string
  description = "Project name"
}

variable "stack_name" {
  type        = string
  description = "Stack name"
}

variable "secrets" {
  description = "Secrets to create in Secrets Manager"
  type = map(object({
    name                    = string
    description             = optional(string, "")
    value                   = string
    kms_key_id              = optional(string, null)
    recovery_window_in_days = optional(number, 30)
  }))
  default = {}
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply"
  default     = {}
}