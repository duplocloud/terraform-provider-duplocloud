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
    PortMappings = [
      {
        ContainerPort = "80",
        HostPort      = "80",
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

# Example to set runtime platform
resource "duplocloud_ecs_task_definition" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  family    = "def02"
  container_definitions = jsonencode([{
    Name  = "contain01"
    Image = "nginx:latest"
    Environment = [
      { Name = "NGINX_HOST", Value = "foo" }
    ]
    PortMappings = [
      {
        ContainerPort = 80,
        HostPort      = 80,
        Protocol = {
          Value = "tcp"
        }
      }
    ]
  }])
  cpu                      = "1024"
  memory                   = "3072"
  requires_compatibilities = ["FARGATE"]
  prevent_tf_destroy       = "false"
  runtime_platform {
    cpu_architecture        = "X86_64"
    operating_system_family = "Linux"
  }
  network_mode = "awsvpc"
}

//Example to set volumes with EFS
resource "duplocloud_ecs_task_definition" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  family    = "newecs-task-def"
  container_definitions = jsonencode([{
    Name  = "default"
    Image = "nginx:latest",
    port  = "80"
    Environment = [
      { Name = "NGINX_HOST", Value = "foo" }
    ]
    PortMappings = [
      {
        ContainerPort = "80",
        HostPort      = "80",
        Protocol = {
          Value = "tcp"
        }
      }
    ]
  }])
  cpu                      = "256"
  memory                   = "1024"
  requires_compatibilities = ["FARGATE"]
  volumes = jsonencode([
    {
      EfsVolumeConfiguration = {
        AuthorizationConfig = {
          AccessPointId = "fsap-0123456789abcdef0"
          Iam = {
            Value = "ENABLED"
          }
        }
        FileSystemId  = "fs-0123456789abcdef0"
        RootDirectory = "/"
        TransitEncryption = {
          Value = "ENABLED"
        }
      }
      Name = "efs-vol"
    }

  ])
}