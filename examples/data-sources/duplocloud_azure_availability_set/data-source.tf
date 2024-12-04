data "duplocloud_azure_availability_set" "st" {
  tenant_id = "tenant id"
  name      = "availability-set name"

}

output "out" {
  value = {
    availability_set_id = data.duplocloud_azure_availability_set.st.availability_set_id
  }
}