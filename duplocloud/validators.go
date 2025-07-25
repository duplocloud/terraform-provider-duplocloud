package duplocloud

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/apimachinery/pkg/api/resource"
	apiValidation "k8s.io/apimachinery/pkg/api/validation"
	utilValidation "k8s.io/apimachinery/pkg/util/validation"
)

func validateAnnotations(value interface{}, key string) (ws []string, es []error) {
	m := value.(map[string]interface{})
	for k := range m {
		errors := utilValidation.IsQualifiedName(strings.ToLower(k))
		if len(errors) > 0 {
			for _, e := range errors {
				es = append(es, fmt.Errorf("%s (%q) %s", key, k, e))
			}
		}
	}
	return
}

func validateName(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	errors := apiValidation.NameIsDNSSubdomain(v, false)
	if len(errors) > 0 {
		for _, err := range errors {
			es = append(es, fmt.Errorf("%s %s", key, err))
		}
	}
	return
}

func validateGenerateName(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)

	errors := apiValidation.NameIsDNSLabel(v, true)
	if len(errors) > 0 {
		for _, err := range errors {
			es = append(es, fmt.Errorf("%s %s", key, err))
		}
	}
	return
}

func validateLabels(value interface{}, key string) (ws []string, es []error) {
	m := value.(map[string]interface{})
	for k, v := range m {
		for _, msg := range utilValidation.IsQualifiedName(k) {
			es = append(es, fmt.Errorf("%s (%q) %s", key, k, msg))
		}
		val, isString := v.(string)
		if !isString {
			es = append(es, fmt.Errorf("%s.%s (%#v): Expected value to be string", key, k, v))
			return
		}
		for _, msg := range utilValidation.IsValidLabelValue(val) {
			es = append(es, fmt.Errorf("%s (%q) %s", key, val, msg))
		}
	}
	return
}

func validatePortNum(value interface{}, key string) (ws []string, es []error) {
	errors := utilValidation.IsValidPortNum(value.(int))
	if len(errors) > 0 {
		for _, err := range errors {
			es = append(es, fmt.Errorf("%s %s", key, err))
		}
	}
	return
}

func validateResourceQuantity(value interface{}, key string) (ws []string, es []error) {
	if v, ok := value.(string); ok {
		_, err := resource.ParseQuantity(v)
		if err != nil {
			es = append(es, fmt.Errorf("%s.%s : %s", key, v, err))
		}
	}
	return
}

func validateNonNegativeInteger(value interface{}, key string) (ws []string, es []error) {
	v := value.(int)
	if v < 0 {
		es = append(es, fmt.Errorf("%s must be greater than or equal to 0", key))
	}
	return
}

func validatePositiveInteger(value interface{}, key string) (ws []string, es []error) {
	v := value.(int)
	if v <= 0 {
		es = append(es, fmt.Errorf("%s must be greater than 0", key))
	}
	return
}

func validateTerminationGracePeriodSeconds(value interface{}, key string) (ws []string, es []error) {
	v := value.(int)
	if v < 0 {
		es = append(es, fmt.Errorf("%s must be greater than or equal to 0", key))
	}
	return
}

func validateIntGreaterThan(minValue int) func(value interface{}, key string) (ws []string, es []error) {
	return func(value interface{}, key string) (ws []string, es []error) {
		v := value.(int)
		if v < minValue {
			es = append(es, fmt.Errorf("%s must be greater than or equal to %d", key, minValue))
		}
		return
	}
}

// validateTypeStringNullableInt provides custom error messaging for TypeString ints
// Some arguments require an int value or unspecified, empty field.
func validateTypeStringNullableInt(v interface{}, k string) (ws []string, es []error) {
	value, ok := v.(string)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := strconv.ParseInt(value, 10, 64); err != nil {
		es = append(es, fmt.Errorf("%s: cannot parse '%s' as int: %s", k, value, err))
	}

	return
}

func validateModeBits(value interface{}, key string) (ws []string, es []error) {
	if !strings.HasPrefix(value.(string), "0") {
		es = append(es, fmt.Errorf("%s: value %s should start with '0' (octal numeral)", key, value.(string)))
	}
	v, err := strconv.ParseInt(value.(string), 8, 32)
	if err != nil {
		es = append(es, fmt.Errorf("%s :Cannot parse octal numeral (%#v): %s", key, value, err))
	}
	if v < 0 || v > 0777 {
		es = append(es, fmt.Errorf("%s (%#o) expects octal notation (a value between 0 and 0777)", key, v))
	}
	return
}

//nolint:staticcheck
func validateAttributeValueDoesNotContain(searchString string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		input := v.(string)
		if strings.Contains(input, searchString) {
			errors = append(errors, fmt.Errorf(
				"%q must not contain %q",
				k, searchString))
		}
		return
	}
}

//nolint:staticcheck
func validateAttributeValueIsIn(validValues []string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		input := v.(string)
		isValid := false
		for _, s := range validValues {
			if s == input {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, fmt.Errorf(
				"%q must contain a value from %#v, got %q",
				k, validValues, input))
		}
		return

	}
}

//func validateTypeStringNullableIntOrPercent(v interface{}, key string) (ws []string, es []error) {
//	value, ok := v.(string)
//	if !ok {
//		es = append(es, fmt.Errorf("expected type of %s to be string", key))
//		return
//	}
//
//	if value == "" {
//		return
//	}
//
//	if strings.HasSuffix(value, "%") {
//		percent, err := strconv.ParseInt(strings.TrimSuffix(value, "%"), 10, 32)
//		if err != nil {
//			es = append(es, fmt.Errorf("%s: cannot parse '%s' as percent: %s", key, value, err))
//		}
//		if percent < 0 || percent > 100 {
//			es = append(es, fmt.Errorf("%s: '%s' is not between 0%% and 100%%", key, value))
//		}
//	} else if _, err := strconv.ParseInt(value, 10, 32); err != nil {
//		es = append(es, fmt.Errorf("%s: cannot parse '%s' as int or percent: %s", key, value, err))
//	}
//
//	return
//}

func isStringValid(r *regexp.Regexp, message string) bool {
	return r.MatchString(message)

}

func validateDateTimeFormat(v interface{}, p cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	// Assert the value is a string
	input, ok := v.(string)
	if !ok {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid value type",
				Detail:   "Expected a string for datetime validation.",
			}}
	}

	// Define the regex pattern for the datetime format
	pattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2})?$`

	// Compile the regex
	regex := regexp.MustCompile(pattern)

	// Check if the value matches the regex pattern
	if !regex.MatchString(input) {
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid datetime format",
			Detail:   "Invalid datetime format: " + input + ", expected format: YYYY-MM-DDTHH:MM:SS",
		})
	}

	return diagnostics
}

// validateStringLength returns true if string length is less than max length
func validateStringLength(input string, maxLength int) bool {
	return len(input) <= maxLength
}
