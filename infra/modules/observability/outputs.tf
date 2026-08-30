output "grafana_url" {
  description = "Grafana URL (via the internal ALB host-header rule)"
  value       = "http://${var.grafana.host_header}:${var.listener_port}"
}

output "loki_push_url" {
  description = "Loki push API URL, used by the app's FireLens sidecar"
  value       = "http://${var.loki.host_header}:${var.listener_port}/loki/api/v1/push"
}

output "loki_bucket_name" {
  description = "S3 bucket backing Loki's chunk/index storage"
  value       = aws_s3_bucket.loki_logs.bucket
}
