package duplosdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploEnabled is a generic flag holder
type DuploEnabled struct {
	Enabled bool `json:"Enabled,omitempty"`
}

// DuploStringValue is a generic value holder
type DuploStringValue struct {
	Value string `json:"Value,omitempty"`
}

// DuploName is a generic name holder
type DuploName struct {
	Name string `json:"Name,omitempty"`
}

// DuploKeyStringValue is a generic key value holder
type DuploKeyStringValue struct {
	Key   string `json:"Key"`
	Value string `json:"Value,omitempty"`
}

// DuploNameStringValue is a generic name value holder
type DuploNameStringValue struct {
	Name  string `json:"Name"`
	Value string `json:"Value,omitempty"`
}

func duploKeyValueSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}
