output "sessions_table_arn" {
  value = aws_dynamodb_table.sessions.arn
}

output "sessions_table_name" {
  value = aws_dynamodb_table.sessions.name
}

output "votes_table_arn" {
  value = aws_dynamodb_table.votes.arn
}

output "votes_table_name" {
  value = aws_dynamodb_table.votes.name
}

output "scenarios_table_arn" {
  value = aws_dynamodb_table.scenarios.arn
}

output "scenarios_table_name" {
  value = aws_dynamodb_table.scenarios.name
}
