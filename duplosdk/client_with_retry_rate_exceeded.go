package duplosdk

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// retry Rate exceeded
// retry - intial setting for dynamodb, need more insights to make it generic.
const (
	RateExceededMaxRetries = 9
	RateExceededMsg        = "Rate exceeded"
	MinStartingDelay       = 1
	MaxStartingDelay       = 7
	MinDelay               = 3
	MaxDelay               = 15
	MinJitterDelay         = 6
)

func calculateBackoffInterval(apiCaller string, attempt int) time.Duration {
	var minDelay, maxDelay int
	switch {
	case attempt == 1:
		minDelay, maxDelay = MinStartingDelay, MaxStartingDelay
	case attempt <= 3:
		minDelay, maxDelay = MinDelay, MaxDelay
	default:
		minDelay, maxDelay = MinJitterDelay, MinJitterDelay+attempt*MinStartingDelay
	}
	interval := rand.Intn(maxDelay-minDelay+1) + minDelay
	if attempt > 1 {
		log.Printf("[TRACE] calculateBackoffInterval minDelay: %d, maxDelay: %d, interval: %d %s", minDelay, maxDelay, interval, apiCaller)
	}
	return time.Duration(interval) * time.Second
}

func retryApiCall(apiCaller string, operationApiCall func() ClientError) ClientError {
	var err ClientError
	var attempt int
	var sleepDuration time.Duration
	for attempt = 1; attempt <= RateExceededMaxRetries; attempt++ {
		delay := calculateBackoffInterval(apiCaller, attempt)
		sleepDuration += delay
		log.Printf("[TRACE] retryApiCall START sleep start (loop_sleep, retry_attempts, api) (%d,%d,%s)", int(delay.Seconds()), attempt, apiCaller)
		time.Sleep(delay)
		if attempt > 1 {
			log.Printf("[WARN] retryApiCall sleep done (loop_sleep, retry_attempts, api) (%d,%d,%s)", int(delay.Seconds()), attempt, apiCaller)
		}
		err = operationApiCall()
		if err == nil {
			return nil
		}

		if !isRateExceededError(err) {
			return err
		}

		log.Printf("[WARN] FAILED_WITH_RATE_EXCEEDED: retryApiCall (total_sleep, retry_attempts, api) (%d,%d,%s)", int(sleepDuration.Seconds()), attempt, apiCaller)
	}
	return newClientError(fmt.Sprintf("FAILED_WITH_RATE_EXCEEDED: Max retry attempts exceeded. (total_sleep, retry_attempts, api) (%d,%d,%s)", int(sleepDuration.Seconds()), attempt, apiCaller))
}

func isRateExceededError(err ClientError) bool {
	if err != nil {
		if value, exists := err.Response()["Message"]; exists {
			if strError, ok := value.(string); ok {
				if RateExceededMsg == strError {
					log.Printf("[ERROR] FAILED_WITH_RATE_EXCEEDED isRateExceededError: detected? %s", strError)
					return true
				}
			}
		}
	}
	return false
}

func (c *Client) getAPIWithRetry(apiName, apiPath string, rp interface{}) ClientError {
	apiCaller := fmt.Sprintf("GET (%s, %s)", apiName, apiPath)
	operation := func() ClientError {
		return c.doAPI("GET", apiName, apiPath, rp)
	}
	return retryApiCall(apiCaller, operation)
}

func (c *Client) deleteAPIWithRetry(apiName, apiPath string, rp interface{}) ClientError {
	apiCaller := fmt.Sprintf("DELETE (%s, %s)", apiName, apiPath)
	operation := func() ClientError {
		return c.doAPI("DELETE", apiName, apiPath, rp)
	}
	return retryApiCall(apiCaller, operation)
}

func (c *Client) postAPIWithRetry(apiName, apiPath string, rq, rp interface{}) ClientError {
	apiCaller := fmt.Sprintf("POST (%s, %s)", apiName, apiPath)
	operation := func() ClientError {
		return c.doAPIWithRequestBody("POST", apiName, apiPath, rq, rp)
	}
	return retryApiCall(apiCaller, operation)
}

func (c *Client) putAPIWithRetry(apiName, apiPath string, rq, rp interface{}) ClientError {
	apiCaller := fmt.Sprintf("PUT (%s, %s)", apiName, apiPath)
	operation := func() ClientError {
		return c.doAPIWithRequestBody("PUT", apiName, apiPath, rq, rp)
	}
	return retryApiCall(apiCaller, operation)
}
