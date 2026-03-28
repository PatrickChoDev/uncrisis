variable "bucket_name" {
  description = "S3 bucket name"
  type        = string
}

variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}
