package duplosdk

import (
	"encoding/base64"
	"math"
	"math/rand"
	"net/url"
	"reflect"
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
	tenant, err := c.GetTenantForUser(tenantID)
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
	var lastError ClientError
	var attempt int

	// Calculate the deadline for the retries
	deadline := time.Now().Add(config.Timeout)

	for time.Now().Before(deadline) {
		var result interface{}
		result, lastError = apiCall()
		if lastError == nil {
			return result, nil
		}

		if !config.IsRetryable(lastError) {
			return nil, lastError
		}

		attempt++
		sleepDuration := calculateBackoff(attempt, config)

		time.Sleep(sleepDuration)

		// Check if we've reached the deadline after sleeping
		if time.Now().After(deadline) {
			break
		}
	}

	return nil, lastError
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
