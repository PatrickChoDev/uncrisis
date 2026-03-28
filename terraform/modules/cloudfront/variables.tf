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
