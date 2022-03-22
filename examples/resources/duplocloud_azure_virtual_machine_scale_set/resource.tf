resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_azure_virtual_machine_scale_set" "vmss" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "tstvmss"

  sku {
    tier     = "Standard"
    name     = "Standard_D1_v2"
    capacity = 2
  }

  os_profile {
    admin_password       = "DuploTest007"
    admin_username       = "duploadmin"
    computer_name_prefix = "tst"
  }

  storage_profile_image_reference {
    sku       = "2016-Datacenter"
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    version   = "latest"
  }

  network_profile {
    name    = "tstvmss"
    primary = true

    ip_configuration {
      name      = "tstvmss"
      subnet_id = "/subscriptions/143ffc59-9394-4ec6-8f5a-c408a238be62/resourceGroups/duploinfra-testdb/providers/Microsoft.Network/virtualNetworks/testdb/subnets/duploinfra-sub01"
    }
    ip_forwarding = true
  }

  upgrade_policy_mode    = "Manual"
  overprovision          = true
  single_placement_group = true
}
