# For AWS
resource "duplocloud_infrastructure_subnet" "aws-subnet" {
  name       = "mySubnet"
  infra_name = "myinfra"
  cidr_block = "10.34.1.0/24"
  type       = "private"
  zone       = "A"
}

# For Azure
resource "duplocloud_infrastructure_subnet" "az-subnet" {
  name              = "mySubnet"
  infra_name        = "myinfra"
  cidr_block        = "10.34.1.0/24"
  type              = "appgwsubnet"
  service_endpoints = ["Microsoft.Storage"]
}
