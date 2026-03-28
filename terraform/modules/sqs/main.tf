# Dead-letter queue for unprocessable votes
resource "aws_sqs_queue" "vote_dlq" {
  name                      = "${var.project_name}-vote-dlq"
  message_retention_seconds = 345600 # 4 days
  tags                      = var.tags
}

# Main vote queue
resource "aws_sqs_queue" "vote" {
  name                       = "${var.project_name}-votes"
  visibility_timeout_seconds = 60
  message_retention_seconds  = 86400 # 1 day
  receive_wait_time_seconds  = 20    # long-polling

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.vote_dlq.arn
    maxReceiveCount     = 3
  })

  tags = var.tags
}
