package duplosdktest

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fixtureCache map[string][]byte

var (
	fc   = fixtureCache{}
	fdir string
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fdir = path.Join(dir, "../internal/duplosdktest")
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

func SetupHttptestWithFixtures() *httptest.Server {
	fc = fixtureCache{}

	return SetupHttptest(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		m := req.Method
		if m == "" {
			m = "GET"
		}

		// TODO: Handlers for mutation
		if m != "GET" {
			res.WriteHeader(599)
			res.Write([]byte("")) // nolint
			return
		}

		// Lookup
		path := req.URL.Path
		if path == "/admin/GetTenantsForUser" {
			buff := fixureList("fixtures/tenant")
			res.WriteHeader(200)
			res.Write(buff) // nolint
		}

		res.WriteHeader(404)
		res.Write([]byte("")) // nolint
	}))
}

func fixureList(subdir string) []byte {
	if buff, ok := fc[subdir]; ok {
		return buff
	}

	origDir, err := os.Getwd()
	if err != nil {
		log.Panicf("getwd: %s", err)
	}

	dir := path.Join(fdir, subdir)
	err = os.Chdir(dir)
	if err != nil {
		log.Panicf("chdir: %s: %s", dir, err)
	}
	defer func() {
		err = os.Chdir(origDir)
		if err != nil {
			log.Panicf("chdir: %s: %s", origDir, err)
		}
	}()

	// List files
	files, err := os.ReadDir(".")
	if err != nil {
		log.Panicf("readdir: %s: %s", subdir, err)
	}

	// size the buffer
	size := len(files) + 1 // 2 brackets, and commas are (length - 1)
	for i := range files {
		info, err := os.Stat(files[i].Name())
		if err != nil {
			log.Panicf("stat: %s: %s", subdir, err)
		}
		size += int(info.Size())
	}

	// fill the buffer
	buff := make([]byte, 0, size)
	buff = append(buff, []byte("[")...)
	for i := range files {
		if i > 0 {
			buff = append(buff, []byte(",")...)
		}
		content, err := os.ReadFile(files[i].Name())
		if err != nil {
			log.Panicf("readfile: %s: %s", subdir, err)
		}
		buff = append(buff, content...)
	}
	buff = append(buff, []byte("]")...)

	// cache the result and return it
	fc[subdir] = buff
	return buff
}

// nolint
func fixtureGet(file string) []byte {
	if buff, ok := fc[file]; ok {
		return buff
	}

	path := path.Join(fdir, file) + ".json"
	buff, err := os.ReadFile(path)
	if err != nil {
		log.Panicf("readfile: %s: %s", path, err)
	}

	// cache the result and return it
	fc[file] = buff
	return buff
}
