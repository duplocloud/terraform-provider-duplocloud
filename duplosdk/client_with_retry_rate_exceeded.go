package duplosdk

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
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

type RetryConf struct {
	RateExceededMaxRetries int
	MinStartingDelay       int
	MaxStartingDelay       int
	MinDelay               int
	MaxDelay               int
	MinJitterDelay         int
}

func NewRetryConf() RetryConf {
	return RetryConf{
		RateExceededMaxRetries: RateExceededMaxRetries,
		MinStartingDelay:       MinStartingDelay,
		MaxStartingDelay:       MaxStartingDelay,
		MinDelay:               MinDelay,
		MaxDelay:               MaxDelay,
		MinJitterDelay:         MinJitterDelay,
	}
}
func calculateBackoffInterval(apiCaller string, attempt, minStartingDelay, maxStartingDelay, minDelay, maxDelay, minJitterDelay int) time.Duration {
	var miDelay, mxDelay int
	switch {
	case attempt == 1:
		miDelay, mxDelay = minStartingDelay, maxStartingDelay
	case attempt <= 3:
		miDelay, mxDelay = minDelay, maxDelay
	default:
		miDelay, mxDelay = minJitterDelay, minJitterDelay+attempt*minStartingDelay
	}
	interval := rand.Intn(mxDelay-miDelay+1) + miDelay
	if attempt > 1 {
		log.Printf("[TRACE] calculateBackoffInterval minDelay: %d, maxDelay: %d, interval: %d %s", miDelay, mxDelay, interval, apiCaller)
	}
	return time.Duration(interval) * time.Second
}

func retryApiCall(apiCaller string, operationApiCall func() ClientError, conf *RetryConf) ClientError {
	var err ClientError
	var attempt int
	var sleepDuration time.Duration
	if conf == nil {
		return newClientError("RetryConf is nil")
	}
	for attempt = 1; attempt <= conf.RateExceededMaxRetries; attempt++ {
		delay := calculateBackoffInterval(apiCaller, attempt, conf.MaxStartingDelay, conf.MaxStartingDelay, conf.MinDelay, conf.MaxDelay, conf.MinJitterDelay)
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
		logString := ""
		if !isRateExceededError(err) && !is400OrTimeoutsError(err) {
			return err
		}
		if isRateExceededError(err) {
			logString = "FAILED_WITH_RATE_EXCEEDED"
		} else {
			logString = "FAILED_WITH_400_OR_TIMEOUT"
		}
		log.Printf("[WARN] %s: retryApiCall (total_sleep, retry_attempts, api) (%d,%d,%s)", logString, int(sleepDuration.Seconds()), attempt, apiCaller)
	}
	return newClientError(fmt.Sprintf("API_RETRIES: Max retry attempts exceeded. (total_sleep, retry_attempts, api) (%d,%d,%s)", int(sleepDuration.Seconds()), attempt, apiCaller))
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

func is400OrTimeoutsError(err ClientError) bool {
	if err != nil {
		if value, exists := err.Response()["Message"]; exists {
			if strError, ok := value.(string); ok {
				if strings.Contains(strError, "HRESULT") {
					log.Printf("[ERROR] FAILED_WITH_TIMEOUT_PERIOD_EXPIRED is400OrTimeoutsError: detected? %s", strError)
					return true
				}
			}
		} else if err.Status() == 400 {
			log.Printf("[ERROR] FAILED_WITH_400 is400OrTimeoutsError: detected? %s", err.Error())
			return true
		}
	}
	return false
}

func (c *Client) getAPIWithRetry(apiName, apiPath string, rp interface{}, conf *RetryConf) ClientError {
	apiCaller := fmt.Sprintf("GET (%s, %s)", apiName, apiPath)
	operation := func() ClientError {
		return c.doAPI("GET", apiName, apiPath, rp)
	}
	return retryApiCall(apiCaller, operation, conf)
}

func (c *Client) deleteAPIWithRetry(apiName, apiPath string, rp interface{}, conf *RetryConf) ClientError {
	apiCaller := fmt.Sprintf("DELETE (%s, %s)", apiName, apiPath)
	operation := func() ClientError {
		return c.doAPI("DELETE", apiName, apiPath, rp)
	}
	return retryApiCall(apiCaller, operation, conf)
}

func (c *Client) postAPIWithRetry(apiName, apiPath string, rq, rp interface{}, conf *RetryConf) ClientError {
	apiCaller := fmt.Sprintf("POST (%s, %s)", apiName, apiPath)
	operation := func() ClientError {
		return c.doAPIWithRequestBody("POST", apiName, apiPath, rq, rp)
	}
	return retryApiCall(apiCaller, operation, conf)
}

func (c *Client) putAPIWithRetry(apiName, apiPath string, rq, rp interface{}, conf *RetryConf) ClientError {
	apiCaller := fmt.Sprintf("PUT (%s, %s)", apiName, apiPath)
	operation := func() ClientError {
		return c.doAPIWithRequestBody("PUT", apiName, apiPath, rq, rp)
	}
	return retryApiCall(apiCaller, operation, conf)
}
