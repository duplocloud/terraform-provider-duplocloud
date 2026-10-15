package duplocloud

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// credentialNamePattern matches attribute names that usually hold credential
// material. It is deliberately broad; everything it catches is either marked
// Sensitive or listed in reviewedNonCredentials below. The trailing s? matters:
// without it a plural name such as auth_tokens or api_keys slips through.
var credentialNamePattern = regexp.MustCompile(
	`(^|_)(password|passwd|pwd|secret|token|credential|private_key|access_key|api_key|apikey|master_key|connection_string)s?($|_)`,
)

// reviewedNonCredentials are attributes the pattern catches that do not hold
// secret material. Each entry is a reviewed decision, not a convenience.
// Adding a name here should be as deliberate as adding Sensitive: true.
var reviewedNonCredentials = map[string]string{
	// Identifiers and names, not key material.
	"secret_name":             "name of a Kubernetes Secret",
	"secret_namespace":        "namespace of a Kubernetes Secret",
	"secret_type":             "Kubernetes Secret type, e.g. Opaque",
	"secret_version":          "version identifier",
	"client_secret_version":   "version identifier",
	"secret_annotations":      "Kubernetes metadata",
	"secret_labels":           "Kubernetes metadata",
	"secret_provider":         "CSI provider name",
	"secret_object":           "CSI secret object reference",
	"secret_file":             "file name within a mounted Secret",
	"secret_ref":              "reference to a Secret, not its contents",
	"secret_key_ref":          "reference to a key within a Secret",
	"node_publish_secret_ref": "reference to a Secret",
	"private_registry_secret": "name of a dockerconfigjson Secret",
	"image_pull_secrets":      "list of Secret names",
	"os_profile_secrets":      "list of certificate references",
	"kms_encryption_key":      "KMS key identifier or ARN, not key material",

	// Caller-supplied tokens that are names or idempotency keys.
	"creation_token": "EFS idempotency token / caller-supplied name",

	// Auth configuration, not the credential itself.
	"authentication_strategy":             "auth mode selector",
	"authorization_type":                  "auth mode selector",
	"authorizer_id":                       "identifier",
	"authorization_endpoint":              "URL",
	"token_endpoint":                      "URL",
	"authentication_request_extra_params": "non-secret request params",
	"on_unauthenticated_request":          "behaviour selector",
	"authenticate_cognito":                "nested config block",
	"authenticate_oidc":                   "nested config block",
	"oauth_scopes":                        "scope strings",
	"password_authentication":             "Enabled/Disabled selector",
	"oidc_token":                          "nested config block, not a token value",
	"oauth_token":                         "nested config block, not a token value",
	"service_account_token":               "nested projection config",

	// Public certificate data.
	"certificate_authority_data":         "public CA certificate",
	"cluster_certificate_authority_data": "public CA certificate",

	// Known and accepted: Lambda Alexa event-source token. Not established as
	// a secret in this context.
	"event_source_token": "Lambda Alexa event source token",
}

// TestCredentialAttributesAreSensitive asserts that every attribute whose name
// indicates credential material is marked Sensitive: true, so Terraform renders
// it as (sensitive value) rather than printing it in plan output. It covers the
// provider configuration block, every resource and every data source.
func TestCredentialAttributesAreSensitive(t *testing.T) {
	p := Provider()

	var findings []string
	visit := func(path, attr string, s *schema.Schema) {
		if s.Sensitive {
			return
		}
		if !holdsCredentialText(s) {
			return
		}
		if !credentialNamePattern.MatchString(attr) {
			return
		}
		if _, ok := reviewedNonCredentials[attr]; ok {
			return
		}
		findings = append(findings, fmt.Sprintf("%s -> %s", path, attr))
	}

	// The provider's own configuration block is covered alongside the resources
	// and data sources, so duplo_token cannot lose its flag unnoticed.
	walkSchema("provider configuration", p.Schema, visit)
	for name, r := range p.ResourcesMap {
		walkSchema(fmt.Sprintf("resource %q", name), r.Schema, visit)
	}
	for name, r := range p.DataSourcesMap {
		walkSchema(fmt.Sprintf("data source %q", name), r.Schema, visit)
	}

	if len(findings) > 0 {
		sort.Strings(findings)
		t.Errorf("%d attribute(s) look like credentials but are not marked Sensitive: true.\n"+
			"Add Sensitive: true, or add the name to reviewedNonCredentials with a reason.\n\n%s",
			len(findings), strings.Join(findings, "\n"))
	}
}

// holdsCredentialText reports whether an attribute can carry a credential as
// its own value: a string, a map, or a list/set of strings. Terraform renders
// the elements of a list or set in plan output too, so a list of strings leaks
// just as readily as a bare string. A list or set of nested blocks carries no
// value of its own, so it is skipped here — walkSchema descends into those and
// checks the attributes inside instead.
func holdsCredentialText(s *schema.Schema) bool {
	switch s.Type {
	case schema.TypeString, schema.TypeMap:
		return true
	case schema.TypeList, schema.TypeSet:
		elem, ok := s.Elem.(*schema.Schema)
		return ok && elem.Type == schema.TypeString
	default:
		return false
	}
}

// walkSchema visits every attribute in a schema map, recursing into nested
// blocks so attributes inside Elem resources are covered too.
func walkSchema(path string, m map[string]*schema.Schema, fn func(path, attr string, s *schema.Schema)) {
	for attr, s := range m {
		fn(path, attr, s)
		if nested, ok := s.Elem.(*schema.Resource); ok && nested != nil {
			walkSchema(path+"."+attr, nested.Schema, fn)
		}
	}
}
