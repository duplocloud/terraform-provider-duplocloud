locals {
  tenant_id = "3a0b2ea5-7403-4765-ad6e-8771ca8fa0fd"
}

resource "duplocloud_k8_secret_provider_class" "spc" {
  tenant_id       = local.tenant_id
  name            = "dev-secret"
  secret_provider = "aws"
  annotations = {
    "a1" = "v1"
    "a2" = "v2"
  }
  labels = {
    "l1" = "v1"
    "l2" = "v2"
  }

  secret_object {
    name = "dev-secret-spc"
    type = "Opaque"
    data {
      key         = "ADP_CONSUMER_APPLICATION_ID"
      object_name = "ADP_CONSUMER_APPLICATION_ID"
    }
    data {
      key         = "ADP_CONSUMER_KEY"
      object_name = "ADP_CONSUMER_KEY"
    }
    data {
      key         = "ADP_CONSUMER_SECRET"
      object_name = "ADP_CONSUMER_SECRET"
    }
  }

  parameters = yamlencode(
    [
      {
        "objectName" : "duploservices-dev02-secret",
        "objectType" : "secretsmanager",
        "jmesPath" : [
          {
            "path" : "ADP_CONSUMER_APPLICATION_ID",
            "objectAlias" : "ADP_CONSUMER_APPLICATION_ID"
          },
          {
            "path" : "ADP_CONSUMER_KEY",
            "objectAlias" : "ADP_CONSUMER_KEY"
          },
          {
            "path" : "ADP_CONSUMER_SECRET",
            "objectAlias" : "ADP_CONSUMER_SECRET"
          }
        ]
      }
    ]
  )
}
