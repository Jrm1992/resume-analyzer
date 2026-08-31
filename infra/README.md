# Infrastructure as Code for resume-analyzer

Terraform modules for deploying resume-analyzer to ECS Fargate. Production runs on Floci (localstack), so every module works with both real AWS and LocalStack endpoints.

## Structure

```
infra/
├── main.tf                  # Root module: wires all modules together
├── variables.tf             # All stack inputs
├── outputs.tf               # Stack outputs
├── versions.tf              # Terraform + provider requirements, backend
├── environments/
│   └── prod.tfvars           # Floci (localstack) production values
└── modules/
    ├── iam/                  # Task + execution IAM roles
    ├── taskdef/              # ECS task definition (container defs, secrets)
    ├── cluster/              # ECS cluster
    ├── service/              # ECS Fargate service
    ├── alb/                  # Application Load Balancer + listener
    ├── loadbalancer/         # Target group + listener rule
    ├── autoscaling/          # Target-tracking CPU/memory scaling
    └── secrets/              # Secrets Manager secrets (optional)
```

## Deploy pipeline

Deploys happen via `.github/workflows/deploy.yml` on push to `main`:

1. Build & push multi-arch Docker image to `ghcr.io` (tagged with git SHA)
2. Runner (self-hosted, `floci-runner`) runs `terraform init/plan/apply` with `environments/prod.tfvars` and `image_tag` from the build job
3. State is persisted on the runner VM (`tfstate/terraform.tfstate`)

`.github/workflows/terraform-deploy.yml` runs `fmt/validate/plan` on every PR as a dry-run.

## Manual apply (on the Floci VM)

```bash
cd infra
terraform init
terraform plan -var-file=environments/prod.tfvars -var image_tag=<tag>
terraform apply -var-file=environments/prod.tfvars -var image_tag=<tag>
```

LocalStack credentials (any values work): `AWS_ACCESS_KEY_ID=test`, `AWS_SECRET_ACCESS_KEY=test`, `AWS_DEFAULT_REGION=us-east-1`.

## Configuration

All inputs are in `variables.tf`; per-environment values in `environments/prod.tfvars`. Key variables:

| Variable | Purpose |
|----------|---------|
| `image_repo` / `image_tag` | Container image (`ghcr.io/jrm1992/resume-analyzer:<tag>`) |
| `service` | ECS cluster name, service name, desired count |
| `container` | Port, CPU, memory for the Fargate task |
| `network` | VPC, subnets, security groups (LocalStack defaults in prod.tfvars) |
| `health_check` | ALB target group health check |
| `load_balancer` | Listener rule priority + host header |
| `config_secret_arn` | Secrets Manager ARN for the JSON config secret (all app config, incl. LLM_API_KEY) |
| `scaling` | Auto-scaling (min/max, CPU/memory targets) |
| `dns` | Optional Route53 record |
| `iam` | Optional pre-existing role ARNs |
| `observability_enabled` | Deploy Loki + Grafana and switch the app to ship logs there instead of CloudWatch |
| `grafana_admin_secret_arn` | Secrets Manager ARN holding the Grafana admin password (required when `observability_enabled = true`) |
| `loki_container` / `grafana_container` | Image + CPU/memory sizing for each |
| `load_balancer.loki` | ALB rule priority + host header for Loki (internal consumers only) |
| `grafana_listener_port` | Dedicated ALB listener port for Grafana (default `3001` — 3000 collides with the Floci UI), no host header needed |

## Adding a new environment

```bash
# Copy environments/prod.tfvars to environments/staging.tfvars and edit values
terraform plan -var-file=environments/staging.tfvars -var image_tag=<tag>
```

## Observability (Loki + Grafana)

Set `observability_enabled = true` to deploy Loki (S3-backed storage) and Grafana as their own ECS services behind the same internal ALB. Grafana gets its own dedicated ALB listener (`grafana_listener_port`, default `3001` — 3000 collides with the Floci UI) that forwards directly to it — no host header involved, so it works over a plain port-forward (e.g. a Tailscale tunnel bound to a fixed port). Loki stays on the existing host-header scheme on the shared listener, since it's only ever called internally (Grafana's datasource, Promtail), never by a human browser.

Before enabling it:

1. Create the Grafana admin password secret and point `grafana_admin_secret_arn` at it — it's a plain string secret, not JSON.
2. Point DNS (or `/etc/hosts` on the Floci VM) for `load_balancer.loki.host_header` at the ALB, the same way `load_balancer.internal.host_header` is already resolved for the app today.
3. Make sure `grafana_listener_port` is reachable through however you access the Floci VM (e.g. forward that port over your Tailscale tunnel).
4. `terraform apply` — this creates a new S3 bucket (`<project>-<stack>-loki-logs`) for Loki's chunk/index storage.

Grafana comes up at `http://<any host that reaches the ALB>:<grafana_listener_port>` with a Loki datasource pre-provisioned.

### Shipping app logs to Loki

The app's task keeps the plain `awslogs` driver (→ CloudWatch, unchanged) — Floci ([floci-io/floci](https://github.com/floci-io/floci)) doesn't implement ECS FireLens (it never wires a container's `awsfirelens` log driver to a sidecar; every container always runs on Docker's stock `json-file` driver underneath, regardless of what the task definition says), so a FireLens/Fluent Bit sidecar is a dead end here.

Instead, a real Promtail reads the app container's logs straight off the Docker daemon on the Floci host, via `docker_sd_configs` against `/var/run/docker.sock` — this works because Floci runs ECS tasks as literal Docker containers on a host you control, unlike real Fargate. Run it as a plain Compose stack on the Floci VM (outside Terraform — it needs the host's Docker socket, which no ECS task definition can get):

```bash
cd infra/promtail
docker compose up -d
```

It discovers containers by the `org.opencontainers.image.title=resume-analyzer` image label (set by `docker/metadata-action` in the build workflow) so it only ships the app's own logs, not Loki/Grafana/Floci's internal containers, and joins Floci's `floci-localstack_default` Docker network to resolve `loki.floci` — adjust that network name in `docker-compose.yml` if your Floci setup names it differently (`docker inspect <any floci-ecs container> --format '{{json .NetworkSettings.Networks}}'` shows it).

Backend is `local` (state file on the VM). Moving to real AWS: uncomment the `s3` backend in `versions.tf`.
