# Example: Importing an existing kubernetes secret
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the config map name
#
terraform import duplocloud_k8_secret.myapp v2/subscriptions/*TENANT_ID*/K8SecretApiV2/*NAME*

# An import has no configuration behind it, so it always runs as though Terraform manages
# the secret's contents: the values are read and written to state in plaintext, once,
# before `manage_secret_data = false` can take effect.  Subsequent reads mask them, but
# the values remain in the state history of a versioned remote backend.  Rotate the
# secret after importing, or scrub the state history, if that matters.
