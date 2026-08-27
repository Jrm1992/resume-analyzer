# modules/loadbalancer/main.tf

resource "aws_lb_target_group" "main" {
  name        = var.target_group_name
  port        = var.container_port
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    path                = var.health_check_path
    interval            = var.health_check_interval
    timeout             = var.health_check_timeout
    healthy_threshold   = 2
    unhealthy_threshold = var.health_check_retries
    matcher             = "200"
    protocol            = "HTTP"
    port                = "traffic-port"
  }

  tags = var.tags
}

resource "aws_lb_listener_rule" "main" {
  listener_arn = var.listener_arn
  priority     = var.rule_priority

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }

  condition {
    host_header {
      values = [var.host_header]
    }
  }
}

output "target_group_arn" {
  description = "Target group ARN"
  value       = aws_lb_target_group.main.arn
}

output "target_group_name" {
  description = "Target group name"
  value       = aws_lb_target_group.main.name
}

output "listener_rule_arn" {
  description = "Listener rule ARN"
  value       = aws_lb_listener_rule.main.arn
}