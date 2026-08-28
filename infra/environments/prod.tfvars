# environments/prod.tfvars
# Floci (localstack) as production environment

aws_region      = "us-east-1"
is_localstack   = true
localstack_host = "floci"

project_name = "resume-analyzer"
stack_name   = "prod"

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

# Floci localstack network IDs (adjust to your Floci VM)
network = {
  vpc_id             = "vpc-0flocilocalstack"
  subnet_ids         = ["subnet-0flocilocalstack1", "subnet-0flocilocalstack2"]
  security_group_ids = ["sg-0flocilocalstack"]
}

health_check = {
  path     = "/health"
  interval = 30
  timeout  = 10
  retries  = 3
}

# Floci ALB listener (adjust to your Floci setup)
load_balancer = {
  internal = {
    listener_arn  = "arn:aws:elasticloadbalancing:us-east-1:000000000000:listener/app/floci-alb/xxxxxxxx/xxxxxxxx"
    rule_priority = 10
    host_header   = "resume-analyzer.floci"
  }
}

env_vars = [
  { name = "LLM_BASE_URL", value = "https://api.openai.com/v1" },
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

# Secret ARN in Floci localstack Secrets Manager
secrets = {
  LLM_API_KEY = "arn:aws:secretsmanager:us-east-1:000000000000:secret:resume-analyzer/prod/llm-api-key-xxxxx"
}

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