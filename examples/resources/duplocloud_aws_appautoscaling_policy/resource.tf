resource "duplocloud_tenant" "duplo-app" {
  account_name = "duplo-app"
  plan_id      = "default"
}

#ECS Service Autoscaling

resource "duplocloud_aws_appautoscaling_target" "asg-target" {
  tenant_id          = duplocloud_tenant.duplo-app.tenant_id
  max_capacity       = 4
  min_capacity       = 2
  resource_id        = "duploservices-duplo-app-ecs-service" # ECS Service Name
  scalable_dimension = "ecs:service:DesiredCount"            # For ECS Service
  service_namespace  = "ecs"
}


resource "duplocloud_aws_appautoscaling_policy" "asg-app-policy" {
  tenant_id          = duplocloud_tenant.duplo-app.tenant_id
  name               = "avg-cpu-utilization"
  policy_type        = "TargetTrackingScaling"
  resource_id        = duplocloud_aws_appautoscaling_target.asg-target.resource_id
  scalable_dimension = duplocloud_aws_appautoscaling_target.asg-target.scalable_dimension
  service_namespace  = duplocloud_aws_appautoscaling_target.asg-target.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization" //ECSServiceAverageMemoryUtilization
    }

    target_value = 40
  }
}
