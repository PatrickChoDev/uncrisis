#!/usr/bin/env bash
# Deploy the React frontend: build, sync to S3, invalidate CloudFront cache.
# Run from any directory — the script resolves its own location.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TF_DIR="${REPO_ROOT}/terraform"
FRONTEND_DIR="${REPO_ROOT}/frontend"

# ── Read outputs from Terraform state ────────────────────────────────────────
echo "==> Reading Terraform outputs..."
BUCKET=$(tofu -chdir="${TF_DIR}" output -raw s3_bucket_name)
DIST_ID=$(tofu -chdir="${TF_DIR}" output -raw cloudfront_distribution_id)
CF_URL=$(tofu -chdir="${TF_DIR}" output -raw cloudfront_url)

echo "  Bucket:  ${BUCKET}"
echo "  Dist ID: ${DIST_ID}"
echo "  URL:     ${CF_URL}"

# ── Always regenerate .env.local from Terraform outputs ──────────────────────
# Recreating it every time ensures AppSync URL/key and CloudFront URL stay in
# sync with the current Terraform state (important after tofu apply recreates
# the AppSync API with a new ID).
echo "==> Regenerating frontend/.env.local from Terraform outputs..."
cat > "${FRONTEND_DIR}/.env.local" <<EOF
VITE_APPSYNC_ENDPOINT=$(tofu -chdir="${TF_DIR}" output -raw appsync_graphql_url)
VITE_APPSYNC_API_KEY=$(tofu -chdir="${TF_DIR}" output -raw appsync_api_key)
VITE_APPSYNC_REGION=ap-southeast-1
VITE_GAME_SERVER_URL=${CF_URL}
EOF

# ── Install dependencies ──────────────────────────────────────────────────────
echo "==> Installing npm dependencies..."
npm --prefix "${FRONTEND_DIR}" install

# ── Build ─────────────────────────────────────────────────────────────────────
echo "==> Building frontend..."
npm --prefix "${FRONTEND_DIR}" run build

# ── Upload to S3 ──────────────────────────────────────────────────────────────
echo "==> Syncing to S3..."
aws s3 sync "${FRONTEND_DIR}/dist/" "s3://${BUCKET}/" --delete

# ── Invalidate CloudFront cache ───────────────────────────────────────────────
echo "==> Invalidating CloudFront cache..."
INVALIDATION_ID=$(aws cloudfront create-invalidation \
  --distribution-id "${DIST_ID}" \
  --paths "/*" \
  --query "Invalidation.Id" \
  --output text)

echo ""
echo "Done. Invalidation ${INVALIDATION_ID} is in progress (~30s)."
echo "Frontend live at: ${CF_URL}"
