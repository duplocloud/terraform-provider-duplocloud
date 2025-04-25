resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_ecs_task_definition" "myservice" {
  # ... see the documentation for the ECS task definition resource ...
}


# Deploy NGINX using ECS
resource "duplocloud_ecs_service" "myservice" {
  tenant_id       = duplocloud_tenant.myapp.tenant_id
  name            = "myecsservice"
  task_definition = duplocloud_ecs_task_definition.myservice.arn
  replicas        = 2
  load_balancer {
    lb_type              = 1
    port                 = "8080"
    external_port        = 80
    protocol             = "HTTP"
    enable_access_logs   = false
    drop_invalid_headers = true
    health_check_url     = "https://example.healthcheckurl.com/healthcheck"
    target_group_count   = 1
  }
}

# Example to create ecs service having asg as capacity_provider, for asg created with agent platform ecs
resource "duplocloud_ecs_service" "myservice" {
  tenant_id       = duplocloud_tenant.myapp.tenant_id
  name            = "myservice"
  task_definition = duplocloud_ecs_task_definition.myservice.arn
  replicas        = 1
  capacity_provider_strategy {
    base              = 0
    weight            = 1
    capacity_provider = "<asg-fullname>"
  }
}