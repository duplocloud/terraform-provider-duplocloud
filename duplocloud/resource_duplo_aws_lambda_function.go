package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsLambdaFunctionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the lambda function will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the lambda function cluster.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 64-MAX_DUPLOSERVICES_LENGTH),
				validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]*$`), "Invalid AWS lambda function name"),
			),
		},
		"description": {
			Description:  "A description of the lambda function.",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(0, 256),
		},
		"fullname": {
			Description: "The full name of the lambda function.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the lambda function.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"role": {
			Description: "The IAM role for the lambda function's execution.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"version": {
			Description: "The version of the lambda function.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"timeout": {
			Description:  "The execution time limit for the lambda function.",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      3,
			ValidateFunc: validation.IntBetween(1, 900),
		},
		"memory_size": {
			Description:  "The maximum amount of memory, in MB, that your lambda function is allowed to use at runtime.",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      128,
			ValidateFunc: validation.IntBetween(128, 10240),
		},
		"s3_bucket": {
			Description: "The S3 bucket where the lambda function package is located.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(3, 63),
				validation.StringMatch(regexp.MustCompile(`^[a-z0-9._-]*$`), "Invalid S3 bucket name"),
				validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "S3 bucket names cannot end with a dot"),
			),
		},
		"s3_key": {
			Description:  "The S3 key in the S3 bucket where the lambda function package is located.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 1024),
		},
		"environment": {
			Description: "Allow customization of the lambda execution environment.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"variables": {
						Description:      "Map of environment variables that are accessible from the function code during execution.",
						Type:             schema.TypeMap,
						Optional:         true,
						Elem:             &schema.Schema{Type: schema.TypeString},
						ValidateDiagFunc: validation.MapKeyMatch(regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9_])+$`), "Invalid environment variable name."),
					},
				},
			},
		},
		"runtime": {
			Description: "The [runtime](https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html) that the lambda function needs.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"nodejs", "nodejs4.3", "nodejs6.10", "nodejs8.10", "nodejs10.x", "nodejs12.x", "nodejs14.x",
				"java8", "java8.al2", "java11",
				"python2.7", "python3.6", "python3.7", "python3.8",
				"dotnetcore1.0", "dotnetcore2.0", "dotnetcore2.1", "dotnetcore3.1",
				"nodejs4.3-edge",
				"go1.x",
				"ruby2.5", "ruby2.7",
				"provided", "provided.al2",
			}, false),
		},
		"handler": {
			Description: "The [entrypoint](https://docs.aws.amazon.com/lambda/latest/dg/walkthrough-custom-events-create-test-function.html) of the lambda function in your code.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 128),
				validation.StringMatch(regexp.MustCompile(`^[^\s]*$`), "Invalid lambda function handler"),
			),
		},
		"last_modified": {
			Description: "A timestamp string of lambda's last modification time.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"source_code_hash": {
			Description: "The SHA 256 hash of the lambda functions's source code package.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"source_code_size": {
			Description: "The size in bytes of the lambda functions's source code package.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
	}
}

// Resource for managing an AWS lambda function
func resourceAwsLambdaFunction() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_lambda_function` manages an AWS lambda function in Duplo.",

		ReadContext:   resourceAwsLambdaFunctionRead,
		CreateContext: resourceAwsLambdaFunctionCreate,
		UpdateContext: resourceAwsLambdaFunctionUpdate,
		DeleteContext: resourceAwsLambdaFunctionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsLambdaFunctionSchema(),
	}
}

/// READ resource
func resourceAwsLambdaFunctionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseAwsLambdaFunctionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLambdaFunctionRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, clientErr := c.LambdaFunctionGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("") // object missing
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s lambda function '%s': %s", tenantID, name, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	flattenAwsLambdaConfiguration(d, &duplo.Configuration)
	// d.Set("s3_bucket", duplo.Code.S3Bucket)
	// d.Set("s3_key", duplo.Code.S3Key)

	log.Printf("[TRACE] resourceAwsLambdaFunctionRead(%s, %s): end", tenantID, name)
	return nil
}

/// CREATE resource
func resourceAwsLambdaFunctionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsLambdaFunctionCreate(%s, %s): start", tenantID, name)

	// Create the request object.
	rq := duplosdk.DuploLambdaCreateRequest{
		FunctionName: name,
		Handler:      d.Get("handler").(string),
		Description:  d.Get("description").(string),
		Timeout:      d.Get("timeout").(int),
		MemorySize:   d.Get("memory_size").(int),
		Code: duplosdk.DuploLambdaCode{
			S3Bucket: d.Get("s3_bucket").(string),
			S3Key:    d.Get("s3_key").(string),
		},
	}
	if v, ok := d.GetOk("runtime"); ok && v != nil && v.(string) != "" {
		rq.Runtime = &duplosdk.DuploStringValue{Value: v.(string)}
	}
	environment, err := getOptionalBlockAsMap(d, "environment")
	if err != nil {
		return diag.FromErr(err)
	}
	rq.Environment = expandAwsLambdaEnvironment(environment)

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	_, err = c.LambdaFunctionCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s lambda function '%s': %s", tenantID, name, err)
	}

	// Wait for Duplo to be able to return the cluster's details.
	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "lambda function", id, func() (interface{}, duplosdk.ClientError) {
		return c.LambdaFunctionGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsLambdaFunctionRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsLambdaFunctionCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsLambdaFunctionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseAwsLambdaFunctionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLambdaFunctionUpdate(%s, %s): start", tenantID, name)

	// Check what changes are needed
	needsCode := needsAwsLambdaFunctionCodeUpdate(d)
	needsConfig := needsAwsLambdaFunctionConfigUpdate(d)
	c := m.(*duplosdk.Client)

	// Optionally update lambda configuration.
	if needsConfig {
		err = updateAwsLambdaFunctionConfig(tenantID, name, d, c)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Optionally update lambda function code.
	if needsCode {
		err = updateAwsLambdaFunctionCode(tenantID, name, d, c)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Read the latest state.
	diags := resourceAwsLambdaFunctionRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsLambdaFunctionUpdate(%s, %s): end", tenantID, name)
	return diags
}

/// DELETE resource
func resourceAwsLambdaFunctionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseAwsLambdaFunctionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLambdaFunctionDelete(%s, %s): start", tenantID, name)

	// Delete the function.
	c := m.(*duplosdk.Client)
	clientErr := c.LambdaFunctionDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s lambda function '%s': %s", tenantID, name, clientErr)
	}

	// Wait up to 60 seconds for Duplo to delete the function.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "lambda function", id, func() (interface{}, duplosdk.ClientError) {
		return c.LambdaFunctionGet(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsLambdaFunctionDelete(%s, %s): end", tenantID, name)
	return nil
}

func parseAwsLambdaFunctionIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAwsLambdaConfiguration(d *schema.ResourceData, duplo *duplosdk.DuploLambdaConfiguration) {
	d.Set("fullname", duplo.FunctionName)
	d.Set("arn", duplo.FunctionArn)
	d.Set("role", duplo.Role)
	d.Set("description", duplo.Description)
	d.Set("last_modified", duplo.LastModified)
	d.Set("source_code_hash", duplo.CodeSha256)
	d.Set("source_code_size", duplo.CodeSize)
	d.Set("memory_size", duplo.MemorySize)
	d.Set("timeout", duplo.Timeout)
	d.Set("handler", duplo.Handler)
	d.Set("version", duplo.Version)
	if duplo.Runtime != nil {
		d.Set("runtime", duplo.Runtime.Value)
	}
	d.Set("environment", flattenAwsLambdaEnvironment(duplo.Environment))
}

func flattenAwsLambdaEnvironment(environment *duplosdk.DuploLambdaEnvironment) []interface{} {
	env := []interface{}{}

	if environment != nil && environment.Variables != nil && len(environment.Variables) > 0 {
		vars := map[string]interface{}{}
		for k, v := range environment.Variables {
			vars[k] = v
		}
		env = []interface{}{map[string]interface{}{"variables": vars}}
	}

	return env
}

func expandAwsLambdaEnvironment(environment map[string]interface{}) *duplosdk.DuploLambdaEnvironment {
	var env *duplosdk.DuploLambdaEnvironment = nil

	if environment != nil {
		if v, ok := environment["variables"]; ok && v != nil && len(v.(map[string]interface{})) > 0 {
			env = &duplosdk.DuploLambdaEnvironment{Variables: map[string]string{}}
			for k, v := range v.(map[string]interface{}) {
				if v == nil {
					env.Variables[k] = ""
				} else {
					env.Variables[k] = v.(string)
				}
			}
		}
	}

	return env
}

func updateAwsLambdaFunctionConfig(tenantID, name string, d *schema.ResourceData, c *duplosdk.Client) error {
	rq := duplosdk.DuploLambdaConfigurationRequest{
		FunctionName: name,
		Handler:      d.Get("handler").(string),
		Description:  d.Get("description").(string),
		Timeout:      d.Get("timeout").(int),
		MemorySize:   d.Get("memory_size").(int),
	}

	if v, ok := d.GetOk("runtime"); ok && v != nil && v.(string) != "" {
		rq.Runtime = &duplosdk.DuploStringValue{Value: v.(string)}
	}

	environment, err := getOptionalBlockAsMap(d, "environment")
	if err != nil {
		return err
	}
	rq.Environment = expandAwsLambdaEnvironment(environment)

	err = c.LambdaFunctionUpdateConfiguration(tenantID, &rq)

	// TODO: Wait for the changes to be applied.

	return err
}

func updateAwsLambdaFunctionCode(tenantID, name string, d *schema.ResourceData, c *duplosdk.Client) error {
	rq := duplosdk.DuploLambdaUpdateRequest{
		FunctionName: name,
		S3Bucket:     d.Get("s3_bucket").(string),
		S3Key:        d.Get("s3_key").(string),
	}
	err := c.LambdaFunctionUpdate(tenantID, &rq)

	// TODO: Wait for the changes to be applied.

	return err
}

func needsAwsLambdaFunctionCodeUpdate(d *schema.ResourceData) bool {
	return d.HasChange("s3_bucket") ||
		d.HasChange("s3_key")
}

func needsAwsLambdaFunctionConfigUpdate(d *schema.ResourceData) bool {
	return d.HasChange("handler") ||
		d.HasChange("runtime") ||
		d.HasChange("description") ||
		d.HasChange("timeout") ||
		d.HasChange("memory_size") ||
		d.HasChange("environment")
}
