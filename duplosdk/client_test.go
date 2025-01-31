package duplosdk

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/duplocloud/terraform-provider-duplocloud/internal/duplosdktest"
)

// Should collect a response body and deserialize it from JSON
func TestGetAPI_ResponseExpected(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "GET", 200, "{\"foo\":\"bar\"}")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rp := struct {
		Foo string `json:"foo"`
	}{}
	result := c.getAPI("testAPI", "/test", &rp)

	assert.Nil(t, result)
	assert.Equal(t, "bar", rp.Foo)
}

// Should raise an error on invalid response JSON.
func TestGetAPI_ResponseExpected_InvalidResponseJson(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "GET", 200, "not JSON")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rp := struct{}{}
	result := c.getAPI("testAPI", "/test", &rp)

	invalidJsonMsg := "getAPI testAPI: cannot unmarshal response from JSON:"
	assert.NotNil(t, result)
	assert.True(t, strings.HasPrefix(result.Error(), invalidJsonMsg))
}

// Should support eliding a response for blank response bodies
func TestGetAPI_ResponseElided_NoContent(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "GET", 204, "")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	result := c.getAPI("testAPI", "/test", nil)

	assert.Nil(t, result)
}

// Should support eliding a response for blank response bodies
func TestGetAPI_ResponseElided_Null(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "GET", 200, "null")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	result := c.getAPI("testAPI", "/test", nil)

	assert.Nil(t, result)
}

// Should support eliding a response for blank response bodies
func TestGetAPI_ResponseElided_Blank(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "GET", 200, "")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	result := c.getAPI("testAPI", "/test", nil)

	assert.Nil(t, result)
}

// Should raise an error on invalid response JSON.
func TestGetAPI_ResponseElided_InvalidResponseJson(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "GET", 200, "not JSON")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	result := c.getAPI("testAPI", "/test", nil)

	invalidJsonMsg := "getAPI testAPI: received unexpected response: not JSON"
	assert.NotNil(t, result)
	assert.True(t, strings.HasPrefix(result.Error(), invalidJsonMsg))
}

// Should collect a response body and deserialize it from JSON
func TestPostAPI_ResponseExpected(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "POST", 200, "{\"foo\":\"bar\"}")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	rp := struct {
		Foo string `json:"foo"`
	}{}
	result := c.postAPI("testAPI", "/test", &rq, &rp)

	assert.Nil(t, result)
	assert.Equal(t, "bar", rp.Foo)
}

// Should raise an error on invalid response JSON.
func TestPostAPI_ResponseExpected_InvalidResponseJson(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "POST", 200, "not JSON")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	rp := struct{}{}
	result := c.postAPI("testAPI", "/test", &rq, &rp)

	invalidJsonMsg := "postAPI testAPI: cannot unmarshal response from JSON:"
	assert.NotNil(t, result)
	assert.True(t, strings.HasPrefix(result.Error(), invalidJsonMsg))
}

// Should support eliding a response for blank response bodies
func TestPostAPI_ResponseElided_NoContent(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "POST", 204, "")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.postAPI("testAPI", "/test", &rq, nil)

	assert.Nil(t, result)
}

// Should support eliding a response for blank response bodies
func TestPostAPI_ResponseElided_Null(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "POST", 200, "null")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.postAPI("testAPI", "/test", &rq, nil)

	assert.Nil(t, result)
}

// Should support eliding a response for blank response bodies
func TestPostAPI_ResponseElided_Blank(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "POST", 200, "")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.postAPI("testAPI", "/test", &rq, nil)

	assert.Nil(t, result)
}

// Should raise an error on invalid response JSON.
func TestPostAPI_ResponseElided_InvalidResponseJson(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "POST", 200, "not JSON")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.postAPI("testAPI", "/test", &rq, nil)

	invalidJsonMsg := "postAPI testAPI: received unexpected response: not JSON"
	assert.NotNil(t, result)
	assert.True(t, strings.HasPrefix(result.Error(), invalidJsonMsg))
}

// Should collect a response body and deserialize it from JSON
func TestPutAPI_ResponseExpected(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "PUT", 200, "{\"foo\":\"bar\"}")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	rp := struct {
		Foo string `json:"foo"`
	}{}
	result := c.putAPI("testAPI", "/test", &rq, &rp)

	assert.Nil(t, result)
	assert.Equal(t, "bar", rp.Foo)
}

// Should raise an error on invalid response JSON.
func TestPutAPI_ResponseExpected_InvalidResponseJson(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "PUT", 200, "not JSON")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	rp := struct{}{}
	result := c.putAPI("testAPI", "/test", &rq, &rp)

	invalidJsonMsg := "putAPI testAPI: cannot unmarshal response from JSON:"
	assert.NotNil(t, result)
	assert.True(t, strings.HasPrefix(result.Error(), invalidJsonMsg))
}

// Should support eliding a response for blank response bodies
func TestPutAPI_ResponseElided_NoContent(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "PUT", 204, "")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.putAPI("testAPI", "/test", &rq, nil)

	assert.Nil(t, result)
}

// Should support eliding a response for blank response bodies
func TestPutAPI_ResponseElided_Null(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "PUT", 200, "null")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.putAPI("testAPI", "/test", &rq, nil)

	assert.Nil(t, result)
}

// Should support eliding a response for blank response bodies
func TestPutAPI_ResponseElided_Blank(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "PUT", 200, "")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.putAPI("testAPI", "/test", &rq, nil)

	assert.Nil(t, result)
}

// Should raise an error on invalid response JSON.
func TestPutAPI_ResponseElided_InvalidResponseJson(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "PUT", 200, "not JSON")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.putAPI("testAPI", "/test", &rq, nil)

	invalidJsonMsg := "putAPI testAPI: received unexpected response: not JSON"
	assert.NotNil(t, result)
	assert.True(t, strings.HasPrefix(result.Error(), invalidJsonMsg))
}

// Should collect a response body and deserialize it from JSON
func TestDeleteAPI_ResponseExpected(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "DELETE", 200, "{\"foo\":\"bar\"}")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rp := struct {
		Foo string `json:"foo"`
	}{}
	result := c.deleteAPI("testAPI", "/test", &rp)

	assert.Nil(t, result)
	assert.Equal(t, "bar", rp.Foo)
}

// Should raise an error on invalid response JSON.
func TestDeleteAPI_ResponseExpected_InvalidResponseJson(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "DELETE", 200, "not JSON")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	rp := struct{}{}
	result := c.deleteAPI("testAPI", "/test", &rp)

	invalidJsonMsg := "deleteAPI testAPI: cannot unmarshal response from JSON:"
	assert.NotNil(t, result)
	assert.True(t, strings.HasPrefix(result.Error(), invalidJsonMsg))
}

// Should support eliding a response for blank response bodies
func TestDeleteAPI_ResponseElided_NoContent(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "DELETE", 204, "")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	result := c.deleteAPI("testAPI", "/test", nil)

	assert.Nil(t, result)
}

// Should support eliding a response for blank response bodies
func TestDeleteAPI_ResponseElided_Null(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "DELETE", 200, "null")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	result := c.deleteAPI("testAPI", "/test", nil)

	assert.Nil(t, result)
}

// Should support eliding a response for blank response bodies
func TestDeleteAPI_ResponseElided_Blank(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "DELETE", 200, "")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	result := c.deleteAPI("testAPI", "/test", nil)

	assert.Nil(t, result)
}

// Should raise an error on invalid response JSON.
func TestDeleteAPI_ResponseElided_InvalidResponseJson(t *testing.T) {
	srv, c, err := SetupClientOneshot(t, "DELETE", 200, "not JSON")
	defer TeardownClient(srv, c)
	assert.Nil(t, err, err)

	result := c.deleteAPI("testAPI", "/test", nil)

	invalidJsonMsg := "deleteAPI testAPI: received unexpected response: not JSON"
	assert.NotNil(t, result)
	assert.True(t, strings.HasPrefix(result.Error(), invalidJsonMsg))
}

func SetupClientOneshot(t *testing.T, expectedMethod string, status int, body string) (srv *httptest.Server, c *Client, err error) {
	srv = SetupHttptestOneshot(t, expectedMethod, status, body)
	c, err = NewClient(srv.URL, "FAKE")
	return
}

func TeardownClient(srv *httptest.Server, c *Client) {
	TeardownHttptest(srv)
}
