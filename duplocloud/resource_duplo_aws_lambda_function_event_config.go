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

func awsLambdaFunctionEventInvokeConfigSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the lambda asynchronous invocation configuration will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"function_name": {
			Description: "Name of Lambda function this configuration should apply to",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]*$`), "Invalid AWS lambda function name"),
			),
		},
		"max_retry_attempts": {
			Description:  "Maximum number of attempts a Lambda function may retry in case of error",
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(1, 2),
		},
		"max_event_age_in_seconds": {
			Description:  "The maximum age of a request that Lambda sends to a function for processing",
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(60, 21600),
		},
		"qualifier": {
			Description: "The qualifier for the lambda event invoke configuration",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"destination_config": {
			Description: "A configuration block to specify event destinations",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"on_success": {
						Description: "Configured destination for successful asynchronous invocations",
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"destination": {
									Description: "The AWS ARN of the destination resource",
									Type:        schema.TypeString,
									Required:    true,
								},
							},
						},
					},
					"on_failure": {
						Description: "Configured destination for failed asynchronous invocations",
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"destination": {
									Description: "The AWS ARN of the destination resource",
									Type:        schema.TypeString,
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Resource for managing an AWS lambda function
func resourceAwsLambdaEventInvokeConfigFunction() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_lambda_function_event_config` manages an AWS lambda function in Duplo.",

		ReadContext:   resourceAwsLambdaFunctionEventInvokeConfigRead,
		CreateContext: resourceAwsLambdaFunctionEventInvokeConfigCreate,
		UpdateContext: resourceAwsLambdaFunctionEventInvokeConfigUpdate,
		DeleteContext: resourceAwsLambdaFunctionEventInvokeConfigDelete,
		// Importer: &schema.ResourceImporter{
		// 	StateContext: schema.ImportStatePassthroughContext,
		// },
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsLambdaFunctionEventInvokeConfigSchema(),
	}
}

// READ resource
func resourceAwsLambdaFunctionEventInvokeConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id, _ := strings.CutSuffix(d.Id(), "/eventInvokeConfig")
	tenantId, functionName, err := parseAwsLambdaFunctionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLambdaFunctionEventInvokeConfigRead(%s, %s): start", tenantId, functionName)

	duploClient := m.(*duplosdk.Client)
	resource, clientErr := duploClient.LambdaEventInvokeConfigGet(tenantId, functionName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("") // object missing
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s lambda function '%s' event invoke config: %s",
			tenantId, functionName, clientErr)
	}

	d.Set("tenant_id", tenantId)
	d.Set("function_name", functionName)
	applyLambdaEventInvokeConfigToTfResource(d, resource)

	log.Printf("[TRACE] resourceAwsLambdaFunctionEventInvokeConfigRead(%s, %s): end", tenantId, functionName)
	return nil
}

// CREATE resource
func resourceAwsLambdaFunctionEventInvokeConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createOrUpdateLambdaEventInvokeConfiguration(ctx, d, m)
}

// UPDATE resource
func resourceAwsLambdaFunctionEventInvokeConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createOrUpdateLambdaEventInvokeConfiguration(ctx, d, m)
}

func createOrUpdateLambdaEventInvokeConfiguration(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	duploClient := m.(*duplosdk.Client)

	functionName := d.Get("function_name").(string)
	tenantId := d.Get("tenant_id").(string)

	request := buildPutLambdaEventInvokeRequest(d)

	err := duploClient.LambdaEventInvokeConfigCreateOrUpdate(tenantId, functionName, request)

	if err != nil {
		return diag.Errorf("Could not successfully create event invoke config for lambda %s", functionName)
	}
	id := fmt.Sprintf("%s/%s/eventInvokeConfig", tenantId, functionName)
	d.SetId(id)

	diags := resourceAwsLambdaFunctionEventInvokeConfigRead(ctx, d, m)
	log.Printf("[TRACE] createOrUpdateLambdaEventInvokeConfiguration(%s, %s): end", tenantId, functionName)
	return diags
}

// DELETE resource
func resourceAwsLambdaFunctionEventInvokeConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id, _ := strings.CutSuffix(d.Id(), "/eventInvokeConfig")
	tenantId, functionName, err := parseAwsLambdaFunctionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}

	duploClient := m.(*duplosdk.Client)
	clientErr := duploClient.LambdaEventInvokeConfigDelete(tenantId, functionName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf(
			"Unexpected error caught while deleting tenant %s's function %s event invoke configuration: %s",
			tenantId, functionName, clientErr)
	}

	log.Printf("[TRACE] resourceAwsLambdaFunctionEventInvokeConfigDelete(%s, %s): end", tenantId, functionName)

	return nil
}

func buildPutLambdaEventInvokeRequest(d *schema.ResourceData) duplosdk.PutLambdaFunctionEventInvokeConfiguration {
	request := duplosdk.PutLambdaFunctionEventInvokeConfiguration{
		LambdaFunctionEventInvokeConfiguration: duplosdk.LambdaFunctionEventInvokeConfiguration{
			FunctionName:      d.Get("function_name").(string),
			DestinationConfig: &duplosdk.DestinationConfiguration{},
		},
	}
	addIfDefined(&request, "MaximumEventAgeInSeconds", d.Get("max_event_age_in_seconds").(int))
	addIfDefined(&request, "MaximumRetryAttempts", d.Get("max_retry_attempts").(int))
	addIfDefined(&request, "Qualifier", d.Get("qualifier").(string))
	if value, ok := d.GetOk("destination_config"); ok && len(value.([]interface{})) > 0 {
		destinationConfig := value.([]interface{})[0].(map[string]interface{})

		if onSuccess, exists := destinationConfig["on_success"]; exists && len(onSuccess.([]interface{})) > 0 {
			onSuccessDestination := onSuccess.([]interface{})[0].(map[string]interface{})["destination"].(string)
			request.DestinationConfig.OnSuccess = &duplosdk.DestinationTarget{
				Destination: onSuccessDestination,
			}
		}

		if onFailure, exists := destinationConfig["on_failure"]; exists && len(onFailure.([]interface{})) > 0 {
			onFailureDestination := onFailure.([]interface{})[0].(map[string]interface{})["destination"].(string)
			request.DestinationConfig.OnFailure = &duplosdk.DestinationTarget{
				Destination: onFailureDestination,
			}
		}

		if request.DestinationConfig.OnFailure == nil && request.DestinationConfig.OnSuccess == nil {
			request.DestinationConfig = nil
		}
	}
	return request
}

func applyLambdaEventInvokeConfigToTfResource(d *schema.ResourceData, resource *duplosdk.LambdaFunctionEventInvokeConfiguration) {
	d.Set("max_retry_attempts", resource.MaximumRetryAttempts)
	d.Set("max_event_age_in_seconds", resource.MaximumEventAgeInSeconds)

	if resource.DestinationConfig != nil {
		tfDestinationConfig := map[string][]map[string]string{}

		if resource.DestinationConfig.OnFailure != nil {
			onFailure := map[string]string{
				"destination": resource.DestinationConfig.OnFailure.Destination,
			}
			tfDestinationConfig["on_failure"] = []map[string]string{onFailure}
		}
		if resource.DestinationConfig.OnSuccess != nil {
			onSuccess := map[string]string{
				"destination": resource.DestinationConfig.OnSuccess.Destination,
			}
			tfDestinationConfig["on_success"] = []map[string]string{onSuccess}
		}

		if resource.DestinationConfig.OnSuccess != nil || resource.DestinationConfig.OnFailure != nil {
			d.Set("destination_config", []interface{}{tfDestinationConfig})
		}
	}
}
