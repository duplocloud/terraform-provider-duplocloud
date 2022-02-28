resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_oci_containerengine_node_pool" "myOciNodePool" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  name          = "tf-test"
  node_shape    = "VM.Standard2.1"
  node_image_id = "ocid1.image.oc1.ap-mumbai-1.aaaaaaaagosxifkwha6a6pi2fxx4idf3te3icdsf7z6jar2sxls6xycnehna"

  node_config_details {
    size = 1

    placement_configs {
      availability_domain = "uwFr:AP-MUMBAI-1-AD-1"
      subnet_id           = "ocid1.subnet.oc1.ap-mumbai-1.aaaaaaaasz36nwww2zygjn7arpuq4fbz3z22kn6adlalldvld3b5nu6afuxa"
    }

    freeform_tags = {
      test = "123"
    }
  }

}

