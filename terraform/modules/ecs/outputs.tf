output "cluster_name" {
  value = aws_ecs_cluster.this.name
}

output "service_name" {
  value = aws_ecs_service.game_server.name
}

output "task_definition_arn" {
  value = aws_ecs_task_definition.game_server.arn
}
