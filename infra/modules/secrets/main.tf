# modules/secrets/main.tf

# Create secrets in Secrets Manager if they don't exist
# This is optional - secrets can be created externally and referenced by ARN

resource "aws_secretsmanager_secret" "secrets" {
  for_each = var.secrets

  name                    = each.value.name
  description             = each.value.description
  kms_key_id              = each.value.kms_key_id
  recovery_window_in_days = each.value.recovery_window_in_days
  tags                    = merge(var.tags, { SecretName = each.key })
}

resource "aws_secretsmanager_secret_version" "secrets" {
  for_each = var.secrets

  secret_id     = aws_secretsmanager_secret.secrets[each.key].id
  secret_string = each.value.value
}

output "secret_arns" {
  description = "Map of secret names to ARNs"
  value       = { for k, v in aws_secretsmanager_secret.secrets : k => v.arn }
}

output "secret_names" {
  description = "Map of secret names to secret names"
  value       = { for k, v in aws_secretsmanager_secret.secrets : k => v.name }
}