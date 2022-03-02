resource "duplocloud_tenant" "duplo-app" {
  account_name = "duplo-app"
  plan_id      = "default"
}

#ECS Service Autoscaling
resource "duplocloud_aws_appautoscaling_target" "asg-target" {
  tenant_id          = duplocloud_tenant.duplo-app.tenant_id
  max_capacity       = 4
  min_capacity       = 2
  resource_id        = "duploservices-duplo-app-ecs-service"
  scalable_dimension = "ecs:service:DesiredCount" # For ECS Service
  service_namespace  = "ecs"
}
