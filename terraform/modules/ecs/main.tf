data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# ── VPC — use default VPC for simplicity ─────────────────────────────────────
data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# ── ECS Cluster ───────────────────────────────────────────────────────────────
resource "aws_ecs_cluster" "this" {
  name = "${var.project_name}-cluster"
  tags = var.tags

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# ── CloudWatch log group ──────────────────────────────────────────────────────
resource "aws_cloudwatch_log_group" "game_server" {
  name              = "/ecs/${var.project_name}/game-server"
  retention_in_days = 3
  tags              = var.tags
}

# ── IAM — task execution role (pull image, write logs) ───────────────────────
data "aws_iam_policy_document" "ecs_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "execution" {
  name               = "${var.project_name}-ecs-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume.json
}

resource "aws_iam_role_policy_attachment" "execution_managed" {
  role       = aws_iam_role.execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# ── IAM — task role (SQS, DynamoDB, AppSync) ─────────────────────────────────
resource "aws_iam_role" "task" {
  name               = "${var.project_name}-ecs-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume.json
}

data "aws_iam_policy_document" "task" {
  statement {
    sid = "SQS"
    actions = [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
    ]
    resources = [var.sqs_queue_arn]
  }

  statement {
    sid = "DynamoDB"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:Query",
      "dynamodb:Scan",
    ]
    resources = [
      var.dynamodb_sessions_arn,
      var.dynamodb_votes_arn,
      var.dynamodb_scenarios_arn,
    ]
  }

  statement {
    sid       = "AppSync"
    actions   = ["appsync:GraphQL"]
    resources = ["arn:aws:appsync:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:apis/*"]
  }
}

resource "aws_iam_role_policy" "task" {
  role   = aws_iam_role.task.id
  policy = data.aws_iam_policy_document.task.json
}

# ── Security group ────────────────────────────────────────────────────────────
resource "aws_security_group" "game_server" {
  name        = "${var.project_name}-game-server-sg"
  description = "Allow inbound HTTP on 8080 and all outbound for the game server"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    description = "HTTP game server API"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

# ── Task definition ───────────────────────────────────────────────────────────
resource "aws_ecs_task_definition" "game_server" {
  family                   = "${var.project_name}-game-server"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.execution.arn
  task_role_arn            = aws_iam_role.task.arn
  tags                     = var.tags

  container_definitions = jsonencode([
    {
      name      = "game-server"
      image     = "${var.ecr_image_uri}:latest"
      essential = true

      environment = [
        { name = "AWS_REGION",         value = var.aws_region },
        { name = "SQS_QUEUE_URL",      value = var.sqs_queue_url },
        { name = "APPSYNC_ENDPOINT",   value = var.appsync_endpoint },
        { name = "APPSYNC_API_KEY",    value = var.appsync_api_key },
        { name = "DYNAMODB_SESSIONS",  value = "${var.project_name}-sessions" },
        { name = "DYNAMODB_VOTES",     value = "${var.project_name}-votes" },
        { name = "DYNAMODB_SCENARIOS", value = "${var.project_name}-scenarios" },
        { name = "ROUND_DURATION_SECS", value = "30" },
        { name = "TOTAL_ROUNDS",       value = "5" },
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.game_server.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "game-server"
        }
      }

      portMappings = [
        { containerPort = 8080, hostPort = 8080, protocol = "tcp" }
      ]
    }
  ])
}

# ── ECS Service ───────────────────────────────────────────────────────────────
resource "aws_ecs_service" "game_server" {
  name            = "${var.project_name}-game-server"
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.game_server.arn
  desired_count   = 1
  tags            = var.tags

  # Prefer Fargate Spot (~70% cheaper). Falls back to regular Fargate if Spot
  # capacity is unavailable (rare in ap-southeast-1 for a single small task).
  capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 100
    base              = 0
  }
  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
    base              = 0
  }

  network_configuration {
    subnets          = data.aws_subnets.default.ids
    security_groups  = [aws_security_group.game_server.id]
    assign_public_ip = true
  }

  lifecycle {
    ignore_changes = [task_definition, desired_count]
  }
}
