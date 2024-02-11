package duplosdktest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
func SetupHttptestOneshot(t *testing.T, expectedMethod string, status int, body string) *httptest.Server {
	return SetupHttptest(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		m := req.Method
		if m == "" {
			m = "GET"
		}
		assert.Equal(t, expectedMethod, m)
		assert.Equal(t, "Bearer FAKE", req.Header.Get("Authorization"))
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
func SetupHttptest(handler http.Handler) *httptest.Server {
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
func TeardownHttptest(server *httptest.Server) {
	server.Close()
}
