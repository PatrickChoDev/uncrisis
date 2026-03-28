#!/usr/bin/env bash
# Start the ECS game server (scale desired_count to 1), wait for it to be
# running, then update the CloudFront ECS origin to the new task public IP.
# Run this before a game session. Pair with stop-server.sh when done.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TF_DIR="$(cd "${SCRIPT_DIR}/../terraform" && pwd)"

echo "==> Reading Terraform outputs..."
CLUSTER=$(tofu -chdir="${TF_DIR}" output -raw ecs_cluster_name)
SERVICE=$(tofu -chdir="${TF_DIR}" output -raw ecs_service_name)
DIST_ID=$(tofu -chdir="${TF_DIR}" output -raw cloudfront_distribution_id)
CF_URL=$(tofu -chdir="${TF_DIR}" output -raw cloudfront_url)
REGION=$(tofu -chdir="${TF_DIR}" output -raw aws_region 2>/dev/null || echo "${AWS_DEFAULT_REGION:-ap-southeast-1}")

echo "==> Scaling ECS service to 1 task..."
aws ecs update-service \
  --cluster "${CLUSTER}" \
  --service "${SERVICE}" \
  --desired-count 1 \
  --region "${REGION}" \
  --output json > /dev/null

echo "==> Waiting for task to be RUNNING (~60-90s)..."
aws ecs wait services-stable \
  --cluster "${CLUSTER}" \
  --services "${SERVICE}" \
  --region "${REGION}"

echo "==> Discovering task public IP..."
TASK_ARN=$(aws ecs list-tasks \
  --cluster "${CLUSTER}" \
  --service-name "${SERVICE}" \
  --desired-status RUNNING \
  --region "${REGION}" \
  --query "taskArns[0]" --output text)

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

echo "  Task: ${PUBLIC_IP} (${PUBLIC_DNS})"

# CloudFront requires a hostname — raw IP addresses are not accepted.
echo "==> Updating CloudFront ECS origin to ${PUBLIC_DNS}..."
DIST_CONFIG_JSON=$(aws cloudfront get-distribution-config --id "${DIST_ID}" --output json)
ETAG=$(echo "${DIST_CONFIG_JSON}" | python3 -c "import sys,json; print(json.load(sys.stdin)['ETag'])")
UPDATED_CONFIG=$(echo "${DIST_CONFIG_JSON}" | python3 -c "
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

aws cloudfront create-invalidation \
  --distribution-id "${DIST_ID}" \
  --paths "/sessions*" "/health" \
  --output json > /dev/null

echo ""
echo "Server is up. CloudFront propagation takes ~3-5 min."
echo "Game URL: ${CF_URL}"
echo ""
echo "Run ./scripts/stop-server.sh when done to avoid unnecessary charges."
