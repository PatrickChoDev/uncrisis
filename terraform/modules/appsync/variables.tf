variable "project_name" {
  type = string
}

variable "sessions_table_arn" {
  type = string
}

variable "sessions_table_name" {
  type = string
}

variable "votes_table_arn" {
  type = string
}

variable "votes_table_name" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
