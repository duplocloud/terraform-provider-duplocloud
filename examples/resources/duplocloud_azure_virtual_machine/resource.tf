resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_virtual_machine" "az_vm" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  friendly_name = "test-vm"

  image_id       = "16.04-LTS;Canonical;UbuntuServer"
  capacity       = "Standard_D2s_v3"
  agent_platform = 0 # Duplo native container agent

  admin_username = "azureuser"
  admin_password = "Root!12345"
  disk_size_gb   = 50
  subnet_id      = "duploinfra-default"

  minion_tags {
    key   = "AllocationTags"
    value = "test-host"
  }

  tags {
    key   = "CreatedBy"
    value = "duplo"
  }

  tags {
    key   = "Owner"
    value = "duplo"
  }
}
