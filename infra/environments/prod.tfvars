# environments/prod.tfvars
# Floci (localstack) as production environment

aws_region      = "us-east-1"
is_localstack   = true
localstack_host = "127.0.0.1"
project_name    = "resume-analyzer"
stack_name      = "prod"

# Credentials for localstack (any values work)
aws_access_key = "test"
aws_secret_key = "test"

service = {
  cluster_name  = "floci-cluster"
  service_name  = "resume-analyzer"
  desired_count = 2
}

container = {
  port   = 8080
  cpu    = "1024"
  memory = "2048"
}

# Floci localstack network IDs (default VPC in LocalStack)
network = {
  vpc_id             = "vpc-default"
  subnet_ids         = ["subnet-default-a", "subnet-default-b", "subnet-default-c"]
  security_group_ids = ["sg-default"]
}

health_check = {
  path     = "/healthz"
  interval = 30
  timeout  = 10
  retries  = 3
}

# Floci ALB listener (rule_priority and host_header only; listener_arn created by ALB module)
load_balancer = {
  internal = {
    listener_arn  = ""
    rule_priority = 10
    host_header   = "resume-analyzer.floci"
  }
}

env_vars = [
  { name = "LLM_MODEL", value = "gpt-4o-mini" },
  { name = "LLM_MAX_TOKENS", value = "4000" },
  { name = "LLM_TIMEOUT_SEC", value = "120" },
  { name = "LLM_RESPONSE_FORMAT", value = "json_object" },
  { name = "MAX_PDF_MB", value = "10" },
  { name = "PORT", value = "8080" },
  { name = "WORKERS", value = "4" },
  { name = "QUEUE_CAPACITY", value = "200" },
  { name = "JOB_TTL_MIN", value = "60" }
]

config_secret_arn = "arn:aws:secretsmanager:us-east-1:000000000000:secret:resume-analyzer/config-2G7SXF"

secrets = {}

secrets_to_create = {}

scaling = {
  enabled            = true
  min_capacity       = 2
  max_capacity       = 10
  cpu_target         = 70
  memory_target      = 80
  scale_in_cooldown  = 300
  scale_out_cooldown = 60
  scheduled_actions  = []
}

dns = {
  zone_name   = ""
  record_name = ""
}

iam = {
  task_role_arn      = ""
  execution_role_arn = ""
}