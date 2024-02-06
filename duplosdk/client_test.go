package duplosdk

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Should call doAPIWithRequestBody with "POST", apiName, apiPath, rq, rp and return its result
func TestPostAPI_CallDoAPIWithRequestBodyAndReturnResult(t *testing.T) {
	srv, c, err := setupClientOneshot(200, "{\"foo\":\"bar\"}")
	defer teardownClient(srv, c)
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
	srv, c, err := setupClientOneshot(200, "not JSON")
	defer teardownClient(srv, c)
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
	srv, c, err := setupClientOneshot(204, "")
	defer teardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.postAPI("testAPI", "/test", &rq, nil)

	assert.Nil(t, result)
}

// Should support eliding a response for blank response bodies
func TestPostAPI_ResponseElided_Null(t *testing.T) {
	srv, c, err := setupClientOneshot(200, "null")
	defer teardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.postAPI("testAPI", "/test", &rq, nil)

	assert.Nil(t, result)
}

// Should support eliding a response for blank response bodies
func TestPostAPI_ResponseElided_Blank(t *testing.T) {
	srv, c, err := setupClientOneshot(200, "")
	defer teardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.postAPI("testAPI", "/test", &rq, nil)

	assert.Nil(t, result)
}

// Should raise an error on invalid response JSON.
func TestPostAPI_ResponseElided_InvalidResponseJson(t *testing.T) {
	srv, c, err := setupClientOneshot(200, "not JSON")
	defer teardownClient(srv, c)
	assert.Nil(t, err, err)

	rq := struct{}{}
	result := c.postAPI("testAPI", "/test", &rq, nil)

	invalidJsonMsg := "postAPI testAPI: received unexpected response: not JSON"
	assert.NotNil(t, result)
	assert.True(t, strings.HasPrefix(result.Error(), invalidJsonMsg))
}

func setupClientOneshot(status int, body string) (srv *httptest.Server, c *Client, err error) {
	srv = setupHttptestOneshot(status, body)
	c, err = NewClient(srv.URL, "none")
	return
}

func teardownClient(srv *httptest.Server, c *Client) {
	srv.Close()
}

/*
setupHttptestOneshot is a function that sets up a one-shot httptest server with the given status code and response body.

Parameters:
- status (int): The HTTP status code to be returned by the server.
- body (string): The response body to be returned by the server.

Returns:
- *httptest.Server: The one-shot httptest server.

Example:

	server := setupHttptestOneshot(200, "Hello, World!")
	defer teardownHttptest(server)
*/
func setupHttptestOneshot(status int, body string) *httptest.Server {
	return setupHttptest(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		res.Write([]byte(body)) // nolint
	}))
}

/*
setupHttptest is a function that sets up a one-shot httptest server with the given handler.

Parameters:
- handler (http.Handler): The handler to be used by the server.

Returns:
- *httptest.Server: The one-shot httptest server.

Example:

	server := setupHttptest(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// handler logic
	}))
	defer teardownHttptest(server)
*/
func setupHttptest(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}

/*
teardownHttptest is a function that closes the given httptest server.

Parameters:
- server (*httptest.Server): The httptest server to be closed.

Example:

	server := setupHttptest(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// handler logic
	}))
	teardownHttptest(server)
*/
func teardownHttptest(server *httptest.Server) {
	server.Close()
}
