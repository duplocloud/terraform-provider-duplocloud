locals {
  tenant_id = "72916492-b69f-492e-8f64-957ad211aca1"
  cert_arn  = "arn:aws:acm:us-west-2:708951311304:certificate/829b32dc-d106-4229-a96d-123456789"
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
    {
      "objects" : [
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
    }
  )

}
