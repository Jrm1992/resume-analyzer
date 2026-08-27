# modules/iam/main.tf

# Task Role - for the ECS task to access AWS resources
resource "aws_iam_role" "task" {
  count = var.create_roles ? 1 : 0

  name               = var.task_role_name
  assume_role_policy = data.aws_iam_policy_document.task_assume_role.json
  description        = "Task role for ${var.project_name} ${var.stack_name}"
  tags               = var.tags
}

data "aws_iam_policy_document" "task_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "task" {
  count = var.create_roles ? 1 : 0

  name   = "${var.task_role_name}-policy"
  role   = aws_iam_role.task[0].id
  policy = data.aws_iam_policy_document.task_policy.json
}

data "aws_iam_policy_document" "task_policy" {
  dynamic "statement" {
    for_each = var.task_policy_statements
    content {
      sid       = statement.value.sid
      effect    = statement.value.effect
      actions   = statement.value.actions
      resources = statement.value.resources
      dynamic "condition" {
        for_each = statement.value.condition != null ? [statement.value.condition] : []
        content {
          test     = condition.value.test
          variable = condition.value.variable
          values   = condition.value.values
        }
      }
    }
  }

  # Default statements for secrets and SSM
  statement {
    sid    = "SecretsManagerAndSSM"
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue",
      "ssm:GetParameter",
      "ssm:GetParameters"
    ]
    resources = ["*"]
  }
}

# Execution Role - for ECS to pull images, write logs, etc.
resource "aws_iam_role" "execution" {
  count = var.create_roles ? 1 : 0

  name               = var.execution_role_name
  assume_role_policy = data.aws_iam_policy_document.execution_assume_role.json
  description        = "Execution role for ${var.project_name} ${var.stack_name}"
  tags               = var.tags
}

data "aws_iam_policy_document" "execution_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "execution" {
  count = var.create_roles ? 1 : 0

  role       = aws_iam_role.execution[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Outputs
output "task_role_arn" {
  description = "Task role ARN"
  value       = var.create_roles ? aws_iam_role.task[0].arn : var.existing_task_role_arn
}

output "task_role_name" {
  description = "Task role name"
  value       = var.create_roles ? aws_iam_role.task[0].name : var.existing_task_role_arn
}

output "execution_role_arn" {
  description = "Execution role ARN"
  value       = var.create_roles ? aws_iam_role.execution[0].arn : var.existing_execution_role_arn
}

output "execution_role_name" {
  description = "Execution role name"
  value       = var.create_roles ? aws_iam_role.execution[0].name : var.existing_execution_role_arn
}