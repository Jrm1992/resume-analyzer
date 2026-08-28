# modules/alb/outputs.tf

output "alb_arn" {
  description = "ALB ARN"
  value       = aws_lb.main.arn
}

output "alb_dns_name" {
  description = "ALB DNS name"
  value       = aws_lb.main.dns_name
}

output "listener_arn" {
  description = "Listener ARN"
  value       = aws_lb_listener.main.arn
}