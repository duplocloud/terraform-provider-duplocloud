resource "duplocloud_user" "myuser" {
  username                = "me@abc.com"
  roles                   = ["User", "Administrator", "SignupUser", "SecurityAdmin"]
  is_readonly             = false
  reallocate_vpn_address  = false
  regenerate_vpn_password = false
}
