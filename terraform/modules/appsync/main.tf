resource "aws_appsync_graphql_api" "this" {
  name                = "${var.project_name}-api"
  authentication_type = "API_KEY"
  schema              = file("${path.module}/schema.graphql")
  tags                = var.tags
}

resource "aws_appsync_api_key" "this" {
  api_id  = aws_appsync_graphql_api.this.id
  expires = timeadd(timestamp(), "8760h") # 1 year
}

# ── IAM role that lets AppSync read/write DynamoDB ──────────────────────────

data "aws_iam_policy_document" "appsync_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["appsync.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "appsync_dynamo" {
  name               = "${var.project_name}-appsync-dynamo"
  assume_role_policy = data.aws_iam_policy_document.appsync_assume.json
}

data "aws_iam_policy_document" "appsync_dynamo" {
  statement {
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem",
      "dynamodb:Query",
      "dynamodb:Scan",
    ]
    resources = [
      var.sessions_table_arn,
      var.votes_table_arn,
    ]
  }
}

resource "aws_iam_role_policy" "appsync_dynamo" {
  role   = aws_iam_role.appsync_dynamo.id
  policy = data.aws_iam_policy_document.appsync_dynamo.json
}

# ── Data sources ─────────────────────────────────────────────────────────────

resource "aws_appsync_datasource" "sessions" {
  api_id           = aws_appsync_graphql_api.this.id
  name             = "SessionsTable"
  type             = "AMAZON_DYNAMODB"
  service_role_arn = aws_iam_role.appsync_dynamo.arn

  dynamodb_config {
    table_name = var.sessions_table_name
  }
}

resource "aws_appsync_datasource" "votes" {
  api_id           = aws_appsync_graphql_api.this.id
  name             = "VotesTable"
  type             = "AMAZON_DYNAMODB"
  service_role_arn = aws_iam_role.appsync_dynamo.arn

  dynamodb_config {
    table_name = var.votes_table_name
  }
}

# None data source — needed for local resolvers that drive subscriptions
resource "aws_appsync_datasource" "none" {
  api_id = aws_appsync_graphql_api.this.id
  name   = "NoneDS"
  type   = "NONE"
}

# ── Resolvers ─────────────────────────────────────────────────────────────────

# Query.getSession
resource "aws_appsync_resolver" "get_session" {
  api_id      = aws_appsync_graphql_api.this.id
  type        = "Query"
  field       = "getSession"
  data_source = aws_appsync_datasource.sessions.name

  request_template = <<-VTL
    {
      "version": "2017-02-28",
      "operation": "GetItem",
      "key": {
        "sessionId": $util.dynamodb.toDynamoDBJson($ctx.args.sessionId)
      }
    }
  VTL

  response_template = "$util.toJson($ctx.result)"
}

# Mutation.submitVote — writes directly to SQS via the frontend;
# the game server fires updateGameState, which triggers subscriptions.
# Here we just persist the raw vote to DynamoDB as an audit record.
resource "aws_appsync_resolver" "submit_vote" {
  api_id      = aws_appsync_graphql_api.this.id
  type        = "Mutation"
  field       = "submitVote"
  data_source = aws_appsync_datasource.votes.name

  request_template = <<-VTL
    {
      "version": "2017-02-28",
      "operation": "PutItem",
      "key": {
        "sessionId": $util.dynamodb.toDynamoDBJson($ctx.args.input.sessionId),
        "playerId":  $util.dynamodb.toDynamoDBJson($ctx.args.input.playerId)
      },
      "attributeValues": {
        "choice":    $util.dynamodb.toDynamoDBJson($ctx.args.input.choice),
        "round":     $util.dynamodb.toDynamoDBJson($ctx.args.input.round),
        "timestamp": $util.dynamodb.toDynamoDBJson($util.time.nowISO8601())
      }
    }
  VTL

  response_template = "$util.toJson($ctx.result)"
}

# Mutation.updateGameState — called by the ECS game server; drives subscriptions
resource "aws_appsync_resolver" "update_game_state" {
  api_id      = aws_appsync_graphql_api.this.id
  type        = "Mutation"
  field       = "updateGameState"
  data_source = aws_appsync_datasource.none.name

  request_template = <<-VTL
    {
      "version": "2018-05-29",
      "payload": $util.toJson($ctx.args.input)
    }
  VTL

  response_template = "$util.toJson($ctx.result)"
}

# Subscription.onGameStateUpdated
resource "aws_appsync_resolver" "on_game_state_updated" {
  api_id      = aws_appsync_graphql_api.this.id
  type        = "Subscription"
  field       = "onGameStateUpdated"
  data_source = aws_appsync_datasource.none.name

  request_template  = "{\"version\":\"2018-05-29\",\"payload\":{}}"
  response_template = "$util.toJson($ctx.result)"
}
