output "vote_queue_url" {
  value = aws_sqs_queue.vote.url
}

output "vote_queue_arn" {
  value = aws_sqs_queue.vote.arn
}

output "vote_dlq_url" {
  value = aws_sqs_queue.vote_dlq.url
}
