variable "aws_region" {
  description = "AWS region to deploy resources into"
  type        = string
  default     = "ap-southeast-1"
}

variable "project_name" {
  description = "Prefix used for all resource names"
  type        = string
  default     = "uncrisis"
}

variable "environment" {
  description = "Deployment environment (dev / staging / prod)"
  type        = string
  default     = "dev"
}
