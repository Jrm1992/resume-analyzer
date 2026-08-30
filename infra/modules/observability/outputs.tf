output "grafana_port" {
  description = "Dedicated ALB listener port for Grafana — reachable at http://<any host resolving to the ALB>:<this port>, no Host header needed"
  value       = var.grafana.listener_port
}

output "loki_push_url" {
  description = "Loki push API URL, used by the app's FireLens sidecar"
  value       = "http://${var.loki.host_header}:${var.listener_port}/loki/api/v1/push"
}

output "loki_bucket_name" {
  description = "S3 bucket backing Loki's chunk/index storage"
  value       = aws_s3_bucket.loki_logs.bucket
}
