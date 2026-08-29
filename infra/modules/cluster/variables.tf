variable "cluster_name" {
  type        = string
  description = "ECS cluster name"
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply"
  default     = {}
}

variable "manage_capacity_providers" {
  type        = bool
  description = "Whether to manage cluster capacity providers"
  default     = false
}
