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
				validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]*$`), "Invalid AWS lambda function name"),
			),
		},
		"package_type": {
			Description:  "The type of lambda package.  Must be `Zip` or `Image`.  Defaults to `Zip`.",
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice([]string{"Zip", "Image"}, false),
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
			Description: "The S3 bucket where the lambda function package is located. Used (and required) only when `package_type` is `\"Zip\"`.",
			Type:        schema.TypeString,
			Optional:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(3, 63),
				validation.StringMatch(regexp.MustCompile(`^[a-z0-9._-]*$`), "Invalid S3 bucket name"),
				validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "S3 bucket names cannot end with a dot"),
			),
		},
		"s3_key": {
			Description:  "The S3 key in the S3 bucket where the lambda function package is located. Used (and required) only when `package_type` is `\"Zip\"`.",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(1, 1024),
		},
		"image_uri": {
			Description:  "The docker image that holds the lambda function's code. Used (and required) only when `package_type` is `\"Image\"`. The image must be in a private ECR.",
			Type:         schema.TypeString,
			Optional:     true,
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
		"tracing_config": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"mode": {
						Description:  "Whether to sample and trace a subset of incoming requests with AWS X-Ray. Valid values are `PassThrough` and `Active`.",
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"PassThrough", "Active"}, false),
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
				"nodejs", "nodejs4.3", "nodejs6.10", "nodejs8.10", "nodejs10.x", "nodejs12.x", "nodejs14.x", "nodejs16.x", "nodejs18.x", "nodejs20.x",
				"java8", "java8.al2", "java11", "java17",
				"python2.7", "python3.6", "python3.7", "python3.8", "python3.9", "python3.10",
				"dotnetcore1.0", "dotnetcore2.0", "dotnetcore2.1", "dotnetcore3.1",
				"dotnet6", "dotnet7",
				"nodejs4.3-edge",
				"go1.x",
				"ruby2.5", "ruby2.7", "ruby3.2",
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
		"image_config": {
			Description: "Configuration for the Lambda function's container image",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"command": {
						Description: "The command that is passed to the container.",
						Type:        schema.TypeList,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"entry_point": {
						Description: "The entry point that is passed to the container.",
						Type:        schema.TypeList,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"working_directory": {
						Description: "The working directory that is passed to the container.",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
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
		"tags": {
			Description: "Map of tags to assign to the object.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"layers": {
			Description: "List of Lambda Layer Version ARNs (maximum of 5) to attach to your Lambda Function.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    5,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"ephemeral_storage": {
			Description:  "The Ephemeral Storage size, in MB, that your lambda function is allowed to use at runtime.",
			Type:         schema.TypeInt,
			Default:      512,
			Optional:     true,
			ValidateFunc: validation.IntBetween(512, 10240),
		},
		"dead_letter_config": {
			Description: "Dead letter queue configuration that specifies the queue or topic where Lambda sends asynchronous events when they fail processing.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"target_arn": {
						Description: "ARN of an SNS topic or SQS queue to notify when an invocation fails.",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
		},
		"architectures": {
			Description: "Instruction set architecture for your Lambda function. Valid values are `[x86_64]` and `[arm64]`. Default is `[x86_64]`",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
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

// READ resource
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
	d.Set("tags", duplo.Tags)

	if duplo.Configuration.DeadLetterConfig != nil && duplo.Configuration.DeadLetterConfig.TargetArn != "" {
		if err := d.Set("dead_letter_config", []interface{}{
			map[string]interface{}{
				"target_arn": string(duplo.Configuration.DeadLetterConfig.TargetArn),
			},
		}); err != nil {
			return diag.Errorf("setting dead_letter_config: %s", err)
		}
	}
	// d.Set("s3_bucket", duplo.Code.S3Bucket)
	// d.Set("s3_key", duplo.Code.S3Key)

	log.Printf("[TRACE] resourceAwsLambdaFunctionRead(%s, %s): end", tenantID, name)
	return nil
}

// CREATE resource
func resourceAwsLambdaFunctionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsLambdaFunctionCreate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	features, _ := c.AdminGetSystemFeatures()
	lambdaMaxLength := 64 - MAX_DUPLOSERVICES_LENGTH
	if features != nil && features.IsTagsBasedResourceMgmtEnabled {
		lambdaMaxLength = 64
	}
	if len(name) > lambdaMaxLength {
		return diag.Errorf("Expected length of lambda function name '%s' to be in the range (1 - %d)", name, lambdaMaxLength)
	}

	var command []string
	var entryPoint []string
	var workingDir string

	// Create the request object.
	rq := duplosdk.DuploLambdaCreateRequest{
		FunctionName: name,
		PackageType: &duplosdk.DuploStringValue{
			Value: getPackageType(d),
		},
		Description:      d.Get("description").(string),
		Timeout:          d.Get("timeout").(int),
		MemorySize:       d.Get("memory_size").(int),
		Code:             duplosdk.DuploLambdaCode{}, // initial assumption
		Tags:             expandAwsLambdaTags(d),
		EphemeralStorage: &duplosdk.DuploLambdaEphemeralStorage{},
	}
	if v, ok := getAsStringArray(d, "layers"); ok && v != nil {
		rq.Layers = v
	}

	if v, ok := getAsStringArray(d, "architectures"); ok && v != nil {
		rq.Architectures = v
	}

	if v, ok := d.GetOk("image_config"); ok && len(v.([]interface{})) > 0 {
		imageConfig := v.([]interface{})[0].(map[string]interface{})
		if value, exists := imageConfig["command"]; exists {
			command = expandStringList(value.([]interface{}))
		}
		if value, exists := imageConfig["entry_point"]; exists {
			entryPoint = expandStringList(value.([]interface{}))
		}
		if value, exists := imageConfig["working_directory"]; exists {
			workingDir = value.(string)
		}
	}

	// Handle the package type
	if rq.PackageType.Value == "Zip" {
		rq.Handler = d.Get("handler").(string)
		rq.Code.S3Bucket = d.Get("s3_bucket").(string)
		rq.Code.S3Key = d.Get("s3_key").(string)
		if v, ok := d.GetOk("runtime"); ok && v != nil && v.(string) != "" {
			rq.Runtime = &duplosdk.DuploStringValue{Value: v.(string)}
		}
	} else if rq.PackageType.Value == "Image" {
		rq.Code.ImageURI = d.Get("image_uri").(string)
		rq.ImageConfig = &duplosdk.DuploLambdaImageConfig{
			Command:    command,
			EntryPoint: entryPoint,
			WorkingDir: workingDir,
		}
	}

	if v, ok := d.GetOk("tracing_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		rq.TracingConfig = &duplosdk.DuploLambdaTracingConfig{
			Mode: duplosdk.DuploStringValue{Value: v.([]interface{})[0].(map[string]interface{})["mode"].(string)},
		}
	}

	environment, err := getOptionalBlockAsMap(d, "environment")
	if err != nil {
		return diag.FromErr(err)
	}
	rq.Environment = expandAwsLambdaEnvironment(environment)

	if v, ok := d.GetOk("ephemeral_storage"); ok && v != nil && v.(int) != 0 {
		rq.EphemeralStorage = &duplosdk.DuploLambdaEphemeralStorage{Size: v.(int)}
	}

	if v, ok := d.GetOk("dead_letter_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		rq.DeadLetterConfig = &duplosdk.DuploDeadLetterConfig{
			TargetArn: v.([]interface{})[0].(map[string]interface{})["target_arn"].(string),
		}
	}

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

// DELETE resource
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
	d.Set("layers", duplo.Layers)
	if duplo.EphemeralStorage != nil {
		d.Set("ephemeral_storage", duplo.EphemeralStorage.Size)
	}
	if duplo.Runtime != nil {
		d.Set("runtime", duplo.Runtime.Value)
	}
	if duplo.PackageType != nil {
		d.Set("package_type", duplo.PackageType.Value)
	}
	d.Set("environment", flattenAwsLambdaEnvironment(duplo.Environment))
	if duplo.TracingConfig != nil {
		d.Set("tracing_config", []interface{}{
			map[string]interface{}{
				"mode": string(duplo.TracingConfig.Mode.Value),
			},
		})
	}
	if duplo.DeadLetterConfig != nil {
		d.Set("dead_letter_config", []interface{}{
			map[string]interface{}{
				"target_arn": duplo.DeadLetterConfig.TargetArn,
			},
		})
	}
	if duplo.ImageConfig != nil {
		imageConfig := map[string]interface{}{
			"command":           duplo.ImageConfig.Command,
			"entry_point":       duplo.ImageConfig.EntryPoint,
			"working_directory": duplo.ImageConfig.WorkingDir,
		}
		d.Set("image_config", []interface{}{imageConfig})
	}
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

func expandAwsLambdaTags(d *schema.ResourceData) map[string]string {
	tags := map[string]string{}
	if v, ok := d.GetOk("tags"); ok && v != nil && len(v.(map[string]interface{})) > 0 {
		for k, v := range v.(map[string]interface{}) {
			if v == nil {
				tags[k] = ""
			} else {
				tags[k] = v.(string)
			}
		}
	}
	return tags
}

func updateAwsLambdaFunctionConfig(tenantID, name string, d *schema.ResourceData, c *duplosdk.Client) error {
	log.Printf("[TRACE] updateAwsLambdaFunctionConfig(%s): start", name)
	rq := duplosdk.DuploLambdaConfigurationRequest{
		FunctionName: d.Get("fullname").(string),
		Handler:      d.Get("handler").(string),
		Description:  d.Get("description").(string),
		Timeout:      d.Get("timeout").(int),
		MemorySize:   d.Get("memory_size").(int),
		Tags:         expandAwsLambdaTags(d),
	}

	if v, ok := getAsStringArray(d, "layers"); ok && v != nil {
		rq.Layers = v
	}

	// Handle the package type
	if getPackageType(d) == "Zip" {
		rq.Handler = d.Get("handler").(string)
		if v, ok := d.GetOk("runtime"); ok && v != nil && v.(string) != "" {
			rq.Runtime = &duplosdk.DuploStringValue{Value: v.(string)}
		}
	}

	if v, ok := d.GetOk("ephemeral_storage"); ok && v != nil && v.(int) != 0 {
		rq.EphemeralStorage = &duplosdk.DuploLambdaEphemeralStorage{Size: v.(int)}
	}

	if v, ok := d.GetOk("tracing_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		rq.TracingConfig = &duplosdk.DuploLambdaTracingConfig{
			Mode: duplosdk.DuploStringValue{Value: v.([]interface{})[0].(map[string]interface{})["mode"].(string)},
		}
	}

	if v, ok := d.GetOk("dead_letter_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		rq.DeadLetterConfig = &duplosdk.DuploDeadLetterConfig{
			TargetArn: v.([]interface{})[0].(map[string]interface{})["target_arn"].(string),
		}
	}

	err := mapImageConfig(d, &rq)
	if err != nil {
		return err
	}

	environment, err := getOptionalBlockAsMap(d, "environment")
	if err != nil {
		return err
	}
	rq.Environment = expandAwsLambdaEnvironment(environment)

	err = c.LambdaFunctionUpdateConfiguration(tenantID, &rq)

	// TODO: Wait for the changes to be applied.
	log.Printf("[TRACE] updateAwsLambdaFunctionConfig(%s): end", name)
	return err
}

func mapImageConfig(d *schema.ResourceData, rq *duplosdk.DuploLambdaConfigurationRequest) error {
	if imageConfigRaw, ok := d.GetOk("image_config"); ok {
		imageConfigList, ok := imageConfigRaw.([]interface{})
		if !ok || len(imageConfigList) == 0 {
			return nil
		}

		imageConfig, ok := imageConfigList[0].(map[string]interface{})
		if !ok {
			return fmt.Errorf("image_config list element is not a valid map")
		}

		// Check if ImageConfig is nil and if so, initialize it
		if rq.ImageConfig == nil {
			rq.ImageConfig = &duplosdk.DuploLambdaImageConfig{}
		}

		if value, exists := imageConfig["command"]; exists {
			commandList, ok := value.([]interface{})
			if !ok {
				return fmt.Errorf("command in image_config is not a valid list")
			}
			rq.ImageConfig.Command = expandStringList(commandList)
		}

		if value, exists := imageConfig["entry_point"]; exists {
			entryPointList, ok := value.([]interface{})
			if !ok {
				return fmt.Errorf("entry_point in image_config is not a valid list")
			}
			rq.ImageConfig.EntryPoint = expandStringList(entryPointList)
		}

		if value, exists := imageConfig["working_directory"]; exists {
			workingDir, ok := value.(string)
			if !ok {
				return fmt.Errorf("working_directory in image_config is not a valid string")
			}
			rq.ImageConfig.WorkingDir = workingDir
		}
	}
	return nil
}

func updateAwsLambdaFunctionCode(tenantID, name string, d *schema.ResourceData, c *duplosdk.Client) error {
	log.Printf("[TRACE] updateAwsLambdaFunctionCode(%s): start", name)
	rq := duplosdk.DuploLambdaUpdateRequest{
		FunctionName: d.Get("fullname").(string),
	}

	// Handle the package type
	packageType := getPackageType(d)
	if packageType == "Zip" {
		rq.S3Bucket = d.Get("s3_bucket").(string)
		rq.S3Key = d.Get("s3_key").(string)
	} else if packageType == "Image" {
		rq.ImageURI = d.Get("image_uri").(string)
	}

	if v, ok := getAsStringArray(d, "architectures"); ok && v != nil {
		rq.Architectures = v
	}

	err := c.LambdaFunctionUpdate(tenantID, &rq)

	// TODO: Wait for the changes to be applied.
	log.Printf("[TRACE] updateAwsLambdaFunctionCode(%s): end", name)
	return err
}

func getPackageType(d *schema.ResourceData) (packageType string) {
	packageType = "Zip"
	if v, ok := d.GetOk("package_type"); ok && v != nil && v.(string) != "" {
		packageType = v.(string)
	}
	return
}

func needsAwsLambdaFunctionCodeUpdate(d *schema.ResourceData) bool {
	return d.HasChange("s3_bucket") ||
		d.HasChange("s3_key") ||
		d.HasChange("image_uri") ||
		d.HasChange("architectures")
}

func needsAwsLambdaFunctionConfigUpdate(d *schema.ResourceData) bool {
	return d.HasChange("handler") ||
		d.HasChange("runtime") ||
		d.HasChange("description") ||
		d.HasChange("timeout") ||
		d.HasChange("memory_size") ||
		d.HasChange("environment") ||
		d.HasChange("tags") ||
		d.HasChange("layers") ||
		d.HasChange("tracing_config") ||
		d.HasChange("ephemeral_storage") ||
		d.HasChange("image_config") ||
		d.HasChange("dead_letter_config")
}
