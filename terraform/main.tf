terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# ──────────────────────────────────────────────
# ECR — container registry for the game server
# ──────────────────────────────────────────────
module "ecr" {
  source         = "./modules/ecr"
  repository_name = "${var.project_name}-game-server"
  tags           = local.tags
}

# ──────────────────────────────────────────────
# S3 — static frontend hosting
# ──────────────────────────────────────────────
module "s3" {
  source       = "./modules/s3"
  bucket_name  = "${var.project_name}-frontend-${var.environment}"
  tags         = local.tags
}

# ──────────────────────────────────────────────
# CloudFront — CDN for the React frontend
# ──────────────────────────────────────────────
module "cloudfront" {
  source              = "./modules/cloudfront"
  s3_bucket_id        = module.s3.bucket_id
  s3_bucket_domain    = module.s3.bucket_regional_domain_name
  s3_bucket_arn       = module.s3.bucket_arn
  project_name        = var.project_name
  tags                = local.tags
  # ecs_task_ip is intentionally left as default ("placeholder.invalid") here.
  # deploy-backend.sh updates the CloudFront origin directly via AWS CLI after
  # each ECS deployment, and lifecycle.ignore_changes keeps tofu from resetting it.
}

# ──────────────────────────────────────────────
# DynamoDB — sessions, scenarios, votes
# ──────────────────────────────────────────────
module "dynamodb" {
  source       = "./modules/dynamodb"
  project_name = var.project_name
  tags         = local.tags
}

# ──────────────────────────────────────────────
# SQS — vote queue
# ──────────────────────────────────────────────
module "sqs" {
  source       = "./modules/sqs"
  project_name = var.project_name
  tags         = local.tags
}

# ──────────────────────────────────────────────
# AppSync — GraphQL API with WebSocket subscriptions
# ──────────────────────────────────────────────
module "appsync" {
  source              = "./modules/appsync"
  project_name        = var.project_name
  sessions_table_arn  = module.dynamodb.sessions_table_arn
  sessions_table_name = module.dynamodb.sessions_table_name
  votes_table_arn     = module.dynamodb.votes_table_arn
  votes_table_name    = module.dynamodb.votes_table_name
  tags                = local.tags
}

# ──────────────────────────────────────────────
# ECS — Fargate game server
# ──────────────────────────────────────────────
module "ecs" {
  source                = "./modules/ecs"
  project_name          = var.project_name
  environment           = var.environment
  aws_region            = var.aws_region
  ecr_image_uri         = module.ecr.repository_url
  sqs_queue_url         = module.sqs.vote_queue_url
  sqs_queue_arn         = module.sqs.vote_queue_arn
  dynamodb_sessions_arn = module.dynamodb.sessions_table_arn
  dynamodb_votes_arn    = module.dynamodb.votes_table_arn
  appsync_endpoint      = module.appsync.graphql_url
  appsync_api_key       = module.appsync.api_key
  tags                  = local.tags
}

# ──────────────────────────────────────────────
# S3 update — grant CloudFront OAC read access
# ──────────────────────────────────────────────
resource "aws_s3_bucket_policy" "frontend" {
  bucket = module.s3.bucket_id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowCloudFrontServicePrincipal"
        Effect    = "Allow"
        Principal = { Service = "cloudfront.amazonaws.com" }
        Action    = "s3:GetObject"
        Resource  = "${module.s3.bucket_arn}/*"
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = module.cloudfront.distribution_arn
          }
        }
      }
    ]
  })
}

locals {
  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}
