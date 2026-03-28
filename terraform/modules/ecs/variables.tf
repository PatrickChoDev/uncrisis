variable "project_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "aws_region" {
  type = string
}

variable "ecr_image_uri" {
  type = string
}

variable "sqs_queue_url" {
  type = string
}

variable "sqs_queue_arn" {
  type = string
}

variable "dynamodb_sessions_arn" {
  type = string
}

variable "dynamodb_votes_arn" {
  type = string
}

variable "appsync_endpoint" {
  type = string
}

variable "appsync_api_key" {
  type      = string
  sensitive = true
}

variable "tags" {
  type    = map(string)
  default = {}
}
