package duplosdk

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type clientError struct {
	message  string
	status   int
	url      string
	response map[string]interface{}
}

func (e clientError) Error() string {
	return e.message
}

func (e clientError) Status() int {
	return e.status
}

func (e clientError) PossibleMissingAPI() bool {
	return e.status == 500 || e.status == 404
}

func (e clientError) URL() string {
	return e.url
}

func (e clientError) Response() map[string]interface{} {
	return e.response
}

type CustomError struct {
	clientError
	message  string
	status   int
	url      string
	response map[string]interface{}
}

func (e CustomError) Status() int {
	return e.status
}

func NewCustomError(message string, status int) CustomError {
	return CustomError{
		message:  message,
		status:   status,
		url:      "",
		response: make(map[string]interface{}),
	}
}

type ClientError interface {
	Error() string
	Status() int
	PossibleMissingAPI() bool
	URL() string
	Response() map[string]interface{}
}

func newHttpError(req *http.Request, status int, message string) ClientError {
	response := map[string]interface{}{"Message": message}
	return clientError{status: status, url: req.URL.String(), message: message, response: response}
}

// An error encountered before we could build the request.
func requestHttpError(url string, message string) ClientError {
	response := map[string]interface{}{"Message": message}
	return clientError{status: -1, url: url, message: message, response: response}
}

// An error encountered before we could parse the response.
func ioHttpError(req *http.Request, err error) ClientError {
	return newHttpError(req, -1, err.Error())
}

// An application logic error encountered in spite of a semantically correct response.
func appHttpError(req *http.Request, message string) ClientError {
	return newHttpError(req, -1, message)
}

// An application logic error encountered in spite of a semantically correct response.
func newClientError(message string) ClientError {
	response := map[string]interface{}{"Message": message}
	return clientError{status: -1, url: "", message: message, response: response}
}

// An error encountered in the HTTP response.
func responseHttpError(req *http.Request, res *http.Response) ClientError {
	status := res.StatusCode
	url := req.URL.String()
	response := map[string]interface{}{}

	// Read the body, but tolerate a failure.
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	message := "(read of body failed)"
	if err == nil {
		message = string(bytes)
	}

	// Older APIs do not always return helpful errors to API clients.
	if !strings.HasPrefix(req.URL.Path, "/v3/") && (status == 400 || status == 404) {
		message = fmt.Sprintf("%s. Please verify object exists in duplocloud.", message)
	}

	// Handle APIs that return proper JSON
	mime := strings.SplitN(res.Header.Get("content-type"), ";", 2)[0]
	if mime == "application/json" {
		err = json.Unmarshal(bytes, &response)
		if err != nil {
			log.Printf("[TRACE] duplo-responseHttpError: failed to parse error response JSON: %s, %s", err, string(bytes))
		}
	}

	// Build the final error message.
	message = fmt.Sprintf("url: %s, status: %d, message: %s", url, status, message)
	log.Printf("[TRACE] duplo-responseHttpError: %s", message)

	// Handle responses that are missing a message - or a JSON parse failure
	if _, ok := response["Message"]; !ok {
		response["Message"] = message
	}

	return clientError{status: res.StatusCode, url: url, message: message, response: response}
}

// Client is a Duplo API client
type Client struct {
	HTTPClient  *http.Client
	HostURL     string
	Token       string
	UserAccount string
}

// NewClient creates a new Duplo API client
func NewClient(host, token string) (*Client, error) {
	if host != "" && token != "" {
		tokenBearer := fmt.Sprintf("Bearer %s", token)
		c := Client{
			HTTPClient: &http.Client{Timeout: 30 * time.Second},
			HostURL:    host,
			Token:      tokenBearer,
		}
		return &c, nil
	}
	return nil, fmt.Errorf("missing provider config for 'duplo_token' 'duplo_host'. Not defined in environment var / main.tf")
}

func (c *Client) doRequestWithStatus(req *http.Request, expectedStatus int) ([]byte, ClientError) {
	req.Header.Set("Authorization", c.Token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if c.UserAccount != "" {
		req.Header.Set("DuploUser", c.UserAccount)
	}
	res, err := c.HTTPClient.Do(req)

	// Handle I/O errors
	if err != nil {
		return nil, ioHttpError(req, err)
	}

	// Pass through HTTP errors, unexpected redirects, or unexpected status codes.
	if res.StatusCode > 300 || (expectedStatus > 0 && expectedStatus != res.StatusCode) {
		return nil, responseHttpError(req, res)
	}

	// Othterwise, we have a response that needs reading.
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("[TRACE] duplo-doRequest: %s", err)
		return nil, ioHttpError(req, err)
	}

	return body, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, ClientError) {
	return c.doRequestWithStatus(req, 0)
}

// Utility method to call an API with a GET request, handling logging, etc.
func (c *Client) getAPI(apiName string, apiPath string, rp interface{}) ClientError {
	return c.doAPI("GET", apiName, apiPath, rp)
}

// Utility method to call an API with a DELETE request, handling logging, etc.
func (c *Client) deleteAPI(apiName string, apiPath string, rp interface{}) ClientError {
	return c.doAPI("DELETE", apiName, apiPath, rp)
}

// Utility method to call an API without a request body, handling logging, etc.
func (c *Client) doAPI(verb string, apiName string, apiPath string, rp interface{}) ClientError {
	apiName = fmt.Sprintf("%sAPI %s", strings.ToLower(verb), apiName)

	// Build the request
	url := fmt.Sprintf("%s/%s", c.HostURL, apiPath)
	log.Printf("[TRACE] %s: prepared request: %s", apiName, url)
	req, err := http.NewRequest(verb, url, nil)
	if err != nil {
		log.Printf("[TRACE] %s: cannot build request: %s", apiName, err.Error())
		return nil
	}

	// Call the API and get the response.
	body, httpErr := c.doRequest(req)
	if httpErr != nil {
		log.Printf("[TRACE] %s: failed: %s", apiName, httpErr.Error())
		return httpErr
	}
	bodyString := string(body)
	if !strings.Contains(apiName, "K8SecretGetList") && !strings.Contains(apiName, "SsmParameterGet") && !strings.Contains(apiName, "SsmParameterList") {
		log.Printf("[TRACE] %s: received response: %s", apiName, bodyString)
	}
	// Check for an expected "null" response.
	if rp == nil {
		log.Printf("[TRACE] %s: expected null response", apiName)
		if bodyString == "null" || bodyString == "" || bodyString == "\"\"" {
			return nil
		}
		message := fmt.Sprintf("%s: received unexpected response: %s", apiName, bodyString)
		log.Printf("[TRACE] %s", message)
		return appHttpError(req, message)
	}

	// Otherwise, interpret it as an object.
	err = json.Unmarshal(body, rp)
	if err != nil {
		message := fmt.Sprintf("%s: cannot unmarshal response from JSON: %s", apiName, err.Error())
		log.Printf("[TRACE] %s", message)
		return newHttpError(req, -1, message)
	}
	return nil
}

// Utility method to call an API with a request, handling logging, etc.
func (c *Client) doAPIWithRequestBody(verb string, apiName string, apiPath string, rq interface{}, rp interface{}) ClientError {
	apiName = fmt.Sprintf("%sAPI %s", strings.ToLower(verb), apiName)
	url := fmt.Sprintf("%s/%s", c.HostURL, apiPath)
	// Build the request
	rqBody, err := json.Marshal(rq)
	if err != nil {
		message := fmt.Sprintf("%s: cannot marshal request to JSON: %s", apiName, err.Error())
		log.Printf("[TRACE] %s", message)
		return requestHttpError(url, message)
	}
	log.Printf("[TRACE] %s: prepared request: %s <= (%s)", apiName, url, rqBody)
	req, err := http.NewRequest(verb, url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] %s: cannot build request: %s", apiName, err.Error())
		return nil
	}

	// Call the API and get the response
	body, httpErr := c.doRequest(req)
	if httpErr != nil {
		log.Printf("[TRACE] %s: failed: %s", apiName, httpErr.Error())
		return httpErr
	}
	bodyString := string(body)
	log.Printf("[TRACE] %s: received response: %s", apiName, bodyString)

	// Check for an expected "null" response.
	if rp == nil {
		log.Printf("[TRACE] %s: expected null response", apiName)
		if bodyString == "null" || bodyString == "" || bodyString == "\"\"" {
			return nil
		}
		message := fmt.Sprintf("%s: received unexpected response: %s", apiName, bodyString)
		log.Printf("[TRACE] %s", message)
		return appHttpError(req, message)
	}

	// Otherwise, interpret it as an object.
	err = json.Unmarshal(body, rp)
	if err != nil {
		message := fmt.Sprintf("%s: cannot unmarshal response from JSON: %s", apiName, err.Error())
		log.Printf("[TRACE] %s", message)
		return appHttpError(req, message)
	}
	return nil
}

// Utility method to call an API with a PUT request, handling logging, etc.
//
//nolint:deadcode,unused // internal API function
func (c *Client) putAPI(apiName string, apiPath string, rq interface{}, rp interface{}) ClientError {
	return c.doAPIWithRequestBody("PUT", apiName, apiPath, rq, rp)
}

// Utility method to call an API with a POST request, handling logging, etc.
func (c *Client) postAPI(apiName string, apiPath string, rq interface{}, rp interface{}) ClientError {
	return c.doAPIWithRequestBody("POST", apiName, apiPath, rq, rp)
}

// retry
const (
	RateExceededMaxRetries = 5
	RateExceededMsg        = "Rate exceeded"
	MinStartingDelay       = 2
	MaxStartingDelay       = 6
	MinDelay               = 3
	MaxDelay               = 15
	MinJitterDelay         = 6
)

func calculateBackoffInterval(attempt int) time.Duration {
	var minDelay, maxDelay int
	switch {
	case attempt == 1:
		minDelay, maxDelay = MinStartingDelay, MaxStartingDelay
	case attempt <= 3:
		minDelay, maxDelay = MinDelay, MaxDelay
	default:
		minDelay, maxDelay = MinJitterDelay, MinJitterDelay+attempt*MinStartingDelay
	}
	return time.Duration(rand.Intn(maxDelay-minDelay+1)+minDelay) * time.Second
}

func retryOperation(operation func() ClientError) ClientError {
	var err ClientError
	for attempt := 1; attempt <= RateExceededMaxRetries; attempt++ {
		delay := calculateBackoffInterval(attempt)
		time.Sleep(delay)

		err = operation()
		if err == nil {
			return nil
		}
		if !isRateExceededError(err) {
			return err
		}
	}
	return newClientError("Max retry attempts exceeded")
}

func isRateExceededError(err ClientError) bool {
	return strings.Contains(err.Error(), RateExceededMsg)
}

func (c *Client) doAPIWithRetry(verb, apiName, apiPath string, rp interface{}) ClientError {
	operation := func() ClientError {
		return c.doAPI(verb, apiName, apiPath, rp)
	}
	return retryOperation(operation)
}

func (c *Client) getAPIWithRetry(apiName, apiPath string, rp interface{}) ClientError {
	operation := func() ClientError {
		return c.doAPI("GET", apiName, apiPath, rp)
	}
	return retryOperation(operation)
}

func (c *Client) deleteAPIWithRetry(apiName, apiPath string, rp interface{}) ClientError {
	operation := func() ClientError {
		return c.doAPI("DELETE", apiName, apiPath, rp)
	}
	return retryOperation(operation)
}

func (c *Client) postAPIWithRetry(apiName, apiPath string, rq, rp interface{}) ClientError {
	operation := func() ClientError {
		return c.doAPIWithRequestBody("POST", apiName, apiPath, rq, rp)
	}
	return retryOperation(operation)
}

func (c *Client) putAPIWithRetry(apiName, apiPath string, rq, rp interface{}) ClientError {
	operation := func() ClientError {
		return c.doAPIWithRequestBody("PUT", apiName, apiPath, rq, rp)
	}
	return retryOperation(operation)
}
