variable "project_name" {
  type = string
}

variable "s3_bucket_id" {
  type = string
}

variable "s3_bucket_domain" {
  type = string
}

variable "s3_bucket_arn" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}

variable "ecs_task_ip" {
  description = "ECS task public IP — set to 'placeholder.invalid' on first apply, then managed by deploy-backend.sh"
  type        = string
  default     = "placeholder.invalid"
}
