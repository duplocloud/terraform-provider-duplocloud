resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Create a task definition for NGINX using ECS
resource "duplocloud_ecs_task_definition" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  family    = "duploservices-default-myservice"
  container_definitions = jsonencode([{
    Name  = "default"
    Image = "nginx:latest"
    Environment = [
      { Name = "NGINX_HOST", Value = "foo" }
    ]
    ContainerMappings = [
      {
        ContainerPorts = "80",
        HostPort       = "80",
        Protocol = {
          Value = "tcp"
        }
      }
    ]
  }])
  cpu                      = "256"
  memory                   = "1024"
  requires_compatibilities = ["FARGATE"]
}
