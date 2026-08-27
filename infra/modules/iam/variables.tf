# modules/iam/variables.tf

variable "project_name" {
  type        = string
  description = "Project name"
}

variable "stack_name" {
  type        = string
  description = "Stack name"
}

variable "task_role_name" {
  type        = string
  description = "Task role name"
}

variable "execution_role_name" {
  type        = string
  description = "Execution role name"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply"
  default     = {}
}

variable "create_roles" {
  type        = bool
  description = "Whether to create roles (false = use existing ARNs)"
  default     = true
}

variable "existing_task_role_arn" {
  type        = string
  description = "Existing task role ARN (when create_roles = false)"
  default     = ""
}

variable "existing_execution_role_arn" {
  type        = string
  description = "Existing execution role ARN (when create_roles = false)"
  default     = ""
}

variable "task_policy_statements" {
  description = "Additional IAM policy statements for task role"
  type = list(object({
    sid       = string
    effect    = string
    actions   = list(string)
    resources = list(string)
    condition = optional(map(string))
  }))
  default = []
}