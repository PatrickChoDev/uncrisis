# Sessions table — one item per active game room
resource "aws_dynamodb_table" "sessions" {
  name         = "${var.project_name}-sessions"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "sessionId"

  attribute {
    name = "sessionId"
    type = "S"
  }

  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }

  tags = var.tags
}

# Votes table — one item per player-per-round
resource "aws_dynamodb_table" "votes" {
  name         = "${var.project_name}-votes"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "sessionId"
  range_key    = "playerId"

  attribute {
    name = "sessionId"
    type = "S"
  }

  attribute {
    name = "playerId"
    type = "S"
  }

  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }

  tags = var.tags
}

# Scenarios table — seeded crisis scenario bank
resource "aws_dynamodb_table" "scenarios" {
  name         = "${var.project_name}-scenarios"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "scenarioId"

  attribute {
    name = "scenarioId"
    type = "S"
  }

  tags = var.tags
}
