# modules/alb/main.tf

resource "aws_lb" "main" {
  name               = var.alb_name
  internal           = var.internal
  load_balancer_type = "application"
  security_groups    = var.security_group_ids
  subnets            = var.subnet_ids
  ip_address_type    = "ipv4"

  enable_deletion_protection = false

  tags = var.tags
}

resource "aws_lb_listener" "main" {
  load_balancer_arn = aws_lb.main.arn
  port              = var.listener_port
  protocol          = var.listener_protocol

  default_action {
    type = "fixed-response"
    fixed_response {
      status_code  = "404"
      content_type = "text/plain"
      message_body = "Not Found"
    }
  }

  tags = var.tags
}
