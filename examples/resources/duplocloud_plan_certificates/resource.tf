resource "duplocloud_plan_certificates" "myplan" {
  plan_id = "myplan"

  certificate {
    name = "my-cert"
    id   = "aws:foo:bar"
  }

  certificate {
    name = "another-cert"
    id   = "aws:foo:bar"
  }
}
