output "cloudfront_url" {
  description = "Public URL of the React frontend"
  value       = "https://${module.cloudfront.domain_name}"
}

output "appsync_graphql_url" {
  description = "AppSync GraphQL endpoint (HTTP + WebSocket)"
  value       = module.appsync.graphql_url
}

output "appsync_api_key" {
  description = "AppSync API key (rotate regularly)"
  value       = module.appsync.api_key
  sensitive   = true
}

output "sqs_vote_queue_url" {
  description = "SQS URL that the game server polls"
  value       = module.sqs.vote_queue_url
}

output "ecr_repository_url" {
  description = "ECR URL — push your game server Docker image here"
  value       = module.ecr.repository_url
}

output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the ECS Fargate service"
  value       = module.ecs.service_name
}

output "s3_bucket_name" {
  description = "S3 bucket name for the frontend (use with aws s3 sync)"
  value       = module.s3.bucket_id
}

output "cloudfront_distribution_id" {
  description = "CloudFront distribution ID (use for cache invalidation)"
  value       = module.cloudfront.distribution_id
}
