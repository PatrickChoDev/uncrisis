#!/usr/bin/env bash
# Stop the ECS game server (scale desired_count to 0) to avoid charges
# during idle periods. Run start-server.sh before the next game session.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TF_DIR="$(cd "${SCRIPT_DIR}/../terraform" && pwd)"

echo "==> Reading Terraform outputs..."
CLUSTER=$(tofu -chdir="${TF_DIR}" output -raw ecs_cluster_name)
SERVICE=$(tofu -chdir="${TF_DIR}" output -raw ecs_service_name)
REGION=$(tofu -chdir="${TF_DIR}" output -raw aws_region 2>/dev/null || echo "${AWS_DEFAULT_REGION:-ap-southeast-1}")

echo "==> Scaling ECS service to 0 tasks..."
aws ecs update-service \
  --cluster "${CLUSTER}" \
  --service "${SERVICE}" \
  --desired-count 0 \
  --region "${REGION}" \
  --output json > /dev/null

echo ""
echo "Game server stopped. No Fargate charges while at 0 tasks."
echo "Run ./scripts/start-server.sh before the next session."
