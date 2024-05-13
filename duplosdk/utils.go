package duplosdk

import (
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// handle path parameter encoding when it might contain slashes
func EncodePathParam(param string) string {
	return url.PathEscape(url.PathEscape(param))
}

//nolint:deadcode,unused // utility function
func isInterfaceNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

// GetDuploServicesNameWithAws builds a duplo resource name, given a tenant ID. The name includes the AWS account ID suffix.
func (c *Client) GetDuploServicesNameWithAws(tenantID, name string) (string, ClientError) {
	return c.GetResourceName("duploservices", tenantID, name, true)
}

func (c *Client) GetDuploServicesNameWithAwsDynamoDbV2(tenantID, name string) (string, ClientError) {
	return c.GetResourceName("duploservices", tenantID, name, false)
}

// GetDuploServicesName builds a duplo resource name, given a tenant ID.
func (c *Client) GetDuploServicesName(tenantID, name string) (string, ClientError) {
	return c.GetResourceName("duploservices", tenantID, name, false)
}

// GetResourceName builds a duplo resource name, given a tenant ID.  It can optionally include the AWS account ID suffix.
func (c *Client) GetResourceName(prefix, tenantID, name string, withAccountSuffix bool) (string, ClientError) {
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return "", err
	}
	if withAccountSuffix {
		accountID, err := c.TenantGetAwsAccountID(tenantID)
		if err != nil {
			return "", err
		}
		return strings.Join([]string{prefix, tenant.AccountName, name, accountID}, "-"), nil
	}
	return strings.Join([]string{prefix, tenant.AccountName, name}, "-"), nil
}

// GetDuploServicesNameWithGcp builds a duplo resource name, given a tenant ID. The name includes the Gcp project ID suffix.
func (c *Client) GetDuploServicesNameWithGcp(tenantID, name string) (string, ClientError) {
	return c.GetResourceNameWithGcp("duploservices", tenantID, name)
}

// GetDuploServicesNameWithGcp builds a duplo resource name, given a tenant ID. The name includes the Gcp project ID suffix.
func (c *Client) GetResourceNameWithGcp(prefix, tenantID, name string) (string, ClientError) {
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return "", err
	}
	projectID, err := c.TenantGetGcpProjectID(tenantID)
	if err != nil {
		return "", err
	}

	return strings.Join([]string{prefix, tenant.AccountName, name, projectID}, "-"), nil
}

// GetDuploServicesPrefix builds a duplo resource name, given a tenant ID.
func (c *Client) GetDuploServicesPrefix(tenantID string) (string, ClientError) {
	return c.GetResourcePrefix("duploservices", tenantID)
}

// GetResourcePrefix builds a duplo resource prefix, given a tenant ID.
func (c *Client) GetResourcePrefix(prefix, tenantID string) (string, ClientError) {
	tenant, err := c.TenantGetV3(tenantID)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{prefix, tenant.AccountName}, "-"), nil
}

// UnprefixName removes a duplo resource prefix from a name.
func UnprefixName(prefix, name string) (string, bool) {
	if strings.HasPrefix(name, prefix) {
		return name[len(prefix)+1:], true
	}

	return name, false
}

// UnwrapName removes a duplo resource prefix and AWS account ID suffix from a name.
func UnwrapName(prefix, accountID, name string, optionalAccountID bool) (string, bool) {
	suffix := "-" + accountID

	var part string
	if !strings.HasSuffix(name, suffix) {
		if !optionalAccountID {
			return name, false
		} else {
			part = name
		}
	} else {
		part = name[0 : len(name)-len(suffix)]
	}

	if !strings.HasPrefix(part, prefix) {
		return name, false
	}

	return part[len(prefix)+1:], true
}

func urlSafeBase64Encode(data string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(data))
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// method that receives the error and returns true if error is retriable, false otherwise
type IsRetryableFunc func(error ClientError) bool

// signature for method that performs the API call which needs the retry logic
type RetryableFunc func() (interface{}, ClientError)

// Configuration to customize behavior of retry wrapper function
type RetryConfig struct {
	MinDelay    time.Duration   // Minimum delay between retries
	MaxDelay    time.Duration   // Maximum delay between retries
	MaxJitter   int             // Maximum jitter in milliseconds to add to the delay
	Timeout     time.Duration   // Total timeout for all retries
	IsRetryable IsRetryableFunc // Function to check if an error is retryable
}

// RetryWithExponentialBackoff tries to execute the provided RetryableFunc according to the RetryConfig.
// Returns the result of the API call or nil and the last error if all retries fail.
func RetryWithExponentialBackoff(apiCall RetryableFunc, config RetryConfig) (interface{}, ClientError) {
	var attempt int
	var retryableMethodName string = GetFunctionName(apiCall)

	// Create a channel to signal the timeout
	timeout := time.After(config.Timeout)

	for {
		fmt.Printf("Calling %s, attempt #%d\n", retryableMethodName, attempt)
		result, lastError := apiCall()
		if lastError == nil {
			return result, nil
		}

		if !config.IsRetryable(lastError) {
			fmt.Printf("Method call %s attempt #%d failed with unrecoverable error\n", retryableMethodName, attempt)
			return nil, lastError
		}

		fmt.Printf("Method call %s attempt #%d failed with retryable error, retrying soon\n", retryableMethodName, attempt)
		attempt++
		sleepDuration := calculateBackoff(attempt, config)

		select {
		case <-timeout:
			// If the timeout channel receives a message, break out of the loop
			fmt.Printf("Method %s failed to succeed before retry timeout\n", retryableMethodName)
			return nil, lastError
		default:
			time.Sleep(sleepDuration)
		}
	}
}

// calculateBackoff calculates the time to wait before the next retry attempt.
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	expBackoff := config.MinDelay * time.Duration(1<<attempt)
	if expBackoff > config.MaxDelay {
		expBackoff = config.MaxDelay
	}

	jitter := time.Duration(
		rand.Intn(
			int(math.Max(float64(0), float64(config.MaxJitter))),
		),
	) * time.Millisecond
	return expBackoff + jitter
}

func DecodeSlashInIdPart(name string) string {
	// for-now we only identified / as problematic e.g. aurora5.7/query_cache_size
	return ReplaceReservedWordsInId("_SLASH_", "/", name)
}

func EncodeSlashInIdPart(name string) string {
	return ReplaceReservedWordsInId("/", "_SLASH_", name)
}

func ReplaceReservedWordsInId(find, replace, name string) string {
	if name != "" && strings.Contains(name, find) {
		log.Printf("[TRACE] ReplaceReservedWordsInId %s %s %s %s ", find, replace, name, strings.Replace(name, find, replace, -1))
		return strings.Replace(name, find, replace, -1)
	}
	return name
}
