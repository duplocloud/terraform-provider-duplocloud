package duplocloud

import (
	"github.com/stretchr/testify/assert"
	"terraform-provider-duplocloud/duplosdk"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestMapImageConfig(t *testing.T) {
	// Mock ResourceData with valid image_config
	d := schema.TestResourceDataRaw(t, awsLambdaFunctionSchema(), map[string]interface{}{
		"image_config": []interface{}{
			map[string]interface{}{
				"command":           []interface{}{"echo", "hello"},
				"entry_point":       []interface{}{"entry1", "entry2"},
				"working_directory": "testDir",
			},
		},
	})

	// Initialize a DuploLambdaConfigurationRequest object
	rq := &duplosdk.DuploLambdaConfigurationRequest{}

	// Call mapImageConfig
	err := mapImageConfig(d, rq)
	assert.Nil(t, err)

	// Assertions
	assert.Equal(t, []string{"echo", "hello"}, rq.ImageConfig.Command)
	assert.Equal(t, []string{"entry1", "entry2"}, rq.ImageConfig.EntryPoint)
	assert.Equal(t, "testDir", rq.ImageConfig.WorkingDir)
}
