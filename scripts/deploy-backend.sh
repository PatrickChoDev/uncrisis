#!/usr/bin/env bash
# Deploy the Go game server: build Docker image, push to ECR, force ECS redeploy,
# then update the CloudFront ECS origin to the new task public IP.
# The frontend URL (VITE_GAME_SERVER_URL) never changes — it always points to CloudFront.
# Run from any directory — the script resolves its own location.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TF_DIR="${REPO_ROOT}/terraform"
BACKEND_DIR="${REPO_ROOT}/backend"

# ── Read outputs from Terraform state ────────────────────────────────────────
echo "==> Reading Terraform outputs..."
ECR_URL=$(tofu -chdir="${TF_DIR}" output -raw ecr_repository_url)
CLUSTER=$(tofu -chdir="${TF_DIR}" output -raw ecs_cluster_name)
SERVICE=$(tofu -chdir="${TF_DIR}" output -raw ecs_service_name)
DIST_ID=$(tofu -chdir="${TF_DIR}" output -raw cloudfront_distribution_id)
CF_URL=$(tofu -chdir="${TF_DIR}" output -raw cloudfront_url)
REGION=$(tofu -chdir="${TF_DIR}" output -raw aws_region 2>/dev/null || echo "${AWS_DEFAULT_REGION:-ap-southeast-1}")

echo "  ECR:          ${ECR_URL}"
echo "  Cluster:      ${CLUSTER}"
echo "  Service:      ${SERVICE}"
echo "  CloudFront:   ${DIST_ID}"
echo "  Region:       ${REGION}"

# ── Docker login to ECR ───────────────────────────────────────────────────────
echo "==> Authenticating with ECR..."
aws ecr get-login-password --region "${REGION}" \
  | docker login --username AWS --password-stdin "${ECR_URL}"

# ── Build image (linux/amd64 to match Fargate) ───────────────────────────────
echo "==> Building Docker image..."
docker build --platform linux/amd64 -t uncrisis-game-server "${BACKEND_DIR}"

# ── Tag and push ──────────────────────────────────────────────────────────────
echo "==> Pushing image to ECR..."
docker tag uncrisis-game-server:latest "${ECR_URL}:latest"
docker push "${ECR_URL}:latest"

# ── Force ECS service redeploy ────────────────────────────────────────────────
echo "==> Forcing ECS service redeployment..."
aws ecs update-service \
  --cluster "${CLUSTER}" \
  --service "${SERVICE}" \
  --force-new-deployment \
  --region "${REGION}" \
  --output json > /dev/null

# ── Wait for the new task to be RUNNING ──────────────────────────────────────
echo "==> Waiting for service to stabilize (this takes ~60-90s)..."
aws ecs wait services-stable \
  --cluster "${CLUSTER}" \
  --services "${SERVICE}" \
  --region "${REGION}"

# ── Discover the new task's public IP ────────────────────────────────────────
echo "==> Discovering new task public IP..."

TASK_ARN=$(aws ecs list-tasks \
  --cluster "${CLUSTER}" \
  --service-name "${SERVICE}" \
  --desired-status RUNNING \
  --region "${REGION}" \
  --query "taskArns[0]" \
  --output text)

if [[ -z "${TASK_ARN}" || "${TASK_ARN}" == "None" ]]; then
  echo "ERROR: No running task found in service ${SERVICE}. Check ECS console for errors."
  exit 1
fi

ENI_ID=$(aws ecs describe-tasks \
  --cluster "${CLUSTER}" \
  --tasks "${TASK_ARN}" \
  --region "${REGION}" \
  --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value | [0]" \
  --output text)

ENI_INFO=$(aws ec2 describe-network-interfaces \
  --network-interface-ids "${ENI_ID}" \
  --region "${REGION}" \
  --query "NetworkInterfaces[0].Association.{ip:PublicIp,dns:PublicDnsName}" \
  --output json)

PUBLIC_IP=$(echo "${ENI_INFO}"  | python3 -c "import sys,json; print(json.load(sys.stdin)['ip'])")
PUBLIC_DNS=$(echo "${ENI_INFO}" | python3 -c "import sys,json; print(json.load(sys.stdin)['dns'])")

if [[ -z "${PUBLIC_IP}" || "${PUBLIC_IP}" == "None" ]]; then
  echo "ERROR: Could not determine public IP for task ${TASK_ARN}."
  exit 1
fi

echo "  New game server: ${PUBLIC_IP} (${PUBLIC_DNS})"

# ── Update CloudFront ECS origin to the new DNS hostname ─────────────────────
# CloudFront requires a hostname — raw IP addresses are not accepted.
echo "==> Updating CloudFront origin to ${PUBLIC_DNS}..."

DIST_CONFIG_JSON=$(aws cloudfront get-distribution-config \
  --id "${DIST_ID}" \
  --output json)

ETAG=$(echo "${DIST_CONFIG_JSON}" | python3 -c "import sys,json; print(json.load(sys.stdin)['ETag'])")

UPDATED_CONFIG=$(echo "${DIST_CONFIG_JSON}" \
  | python3 -c "
import sys, json
data = json.load(sys.stdin)['DistributionConfig']
for origin in data['Origins']['Items']:
    if origin['Id'].startswith('ECS-'):
        origin['DomainName'] = '${PUBLIC_DNS}'
print(json.dumps(data))
")

aws cloudfront update-distribution \
  --id "${DIST_ID}" \
  --if-match "${ETAG}" \
  --distribution-config "${UPDATED_CONFIG}" \
  --output json > /dev/null

echo "  CloudFront origin updated. Propagation takes ~3-5 minutes."

# ── Invalidate /sessions* so any cached responses are cleared ────────────────
echo "==> Creating CloudFront invalidation for /sessions*..."
aws cloudfront create-invalidation \
  --distribution-id "${DIST_ID}" \
  --paths "/sessions*" "/health" \
  --output json > /dev/null

echo ""
echo "Done."
echo "  Game server:     ${PUBLIC_IP} (proxied via CloudFront)"
echo "  Frontend URL:    ${CF_URL}"
echo ""
echo "Note: CloudFront propagation takes ~3-5 min. The game server URL"
echo "      is always ${CF_URL} — no need to update frontend/.env.local."
