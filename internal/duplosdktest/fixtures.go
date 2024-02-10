package duplosdktest

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
)

type fixtureCache map[string][]byte
type handlerCache map[string]http.HandlerFunc

type FixtureAdder func() string
type FixtureWriter func()

var (
	fc   = fixtureCache{}
	hc   = handlerCache{}
	fdir string
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fdir = path.Join(dir, "../internal/duplosdktest/fixtures")
}

func SetupHttptestWithFixtures() *httptest.Server {
	fc = fixtureCache{}

	return SetupHttptest(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		m := req.Method
		if m == "" {
			m = "GET"
		}
		path := req.URL.Path

		if m != "GET" {
			if h, ok := hc[path]; ok {
				h(res, req)
			} else {
				res.WriteHeader(599)
				res.Write([]byte("No test-case handler for " + m + " " + path)) // nolint
			}
		} else if path == "/admin/GetTenantsForUser" {
			buff := fixtureList("tenant")
			res.WriteHeader(200)
			res.Write(buff) // nolint
		} else if strings.HasPrefix(path, "/v2/admin/TenantV2/") {
			id := strings.TrimPrefix(path, "/v2/admin/TenantV2/")
			buff := fixtureGet("tenant/" + id)
			res.WriteHeader(200)
			res.Write(buff) // nolint
		} else {
			res.WriteHeader(404)
			res.Write([]byte("")) // nolint
		}
	}))
}

func OnPostFixture(reqPath string, fixtureLocation string, source interface{}, target interface{}, adder FixtureAdder) {
	hc[reqPath] = func(res http.ResponseWriter, req *http.Request) {
		unmarshallRequestBody(req, source)
		fixtureLocation += "/" + adder()
		if target == nil {
			PostFixture(fixtureLocation, source)
		} else {
			PostFixture(fixtureLocation, target)
		}
		res.WriteHeader(200)
		if target != nil {
			res.Write(fc[fixtureLocation]) // nolint
		}
	}
}

func unmarshallRequestBody(req *http.Request, target interface{}) {
	body, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		log.Panicf("req.ReadBody: %s: %s", req.URL.Path, err)
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		log.Panicf("json.Unmarshal: %s: %s", req.URL.Path, err)
	}
}

func ResetFixtures() {
	fc = fixtureCache{}
	hc = handlerCache{}
}

func PatchFixture(location string, target interface{}, writer FixtureWriter) {
	body := fixtureGet(location)
	err := json.Unmarshal(body, target)
	if err != nil {
		log.Panicf("json.Unmarshal: %s: %s", location, err)
	}

	writer()

	PostFixture(location, target)
}

func PostFixture(location string, source interface{}) {
	body, err := json.Marshal(source)
	if err != nil {
		log.Panicf("json.Marshal: %s: %s", location, err)
	}

	log.Printf("[TRACE] %s: (over)wrote new fixture", location)
	fc[location] = body

	// Invalidate any cached list
	delete(fc, path.Dir(location))
}

func fixtureList(location string) []byte {
	// Return the data if it is cached
	if buff, ok := fc[location]; ok {
		return buff
	}

	dir := path.Join(fdir, location)

	// List files
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Panicf("readdir: %s: %s", location, err)
	}

	// Cache all non-present elements.
	for i := range files {
		name := files[i].Name()
		if strings.HasSuffix(name, ".json") {
			fixtureGet(path.Join(location, strings.TrimSuffix(name, ".json")))
		}
	}

	// Collect all elements, and calculate our final size.
	prefix := location + "/"
	size := 2 // 2 brackets
	locations := make([]string, 0, len(files)*2)
	for key, buff := range fc {

		// Only matching keys.
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		name := strings.TrimPrefix(key, prefix)

		// Only direct descendants.
		if strings.Index(name, "/") != -1 {
			continue
		}

		if size != 2 {
			size += 1 // leading comma
		}
		size += len(buff)
		locations = append(locations, key)
	}

	// Fill the buffer
	buff := make([]byte, 0, size)
	buff = append(buff, []byte("[")...)
	for i := range locations {
		if i > 0 {
			buff = append(buff, []byte(",")...)
		}
		buff = append(buff, fc[locations[i]]...)
	}
	buff = append(buff, []byte("]")...)

	// cache the result and return it
	fc[location] = buff
	return buff
}

// nolint
func fixtureGet(location string) []byte {
	if buff, ok := fc[location]; ok {
		return buff
	}

	path := path.Join(fdir, location) + ".json"
	buff, err := os.ReadFile(path)
	if err != nil {
		log.Panicf("readfile: %s: %s", path, err)
	}

	// cache the result and return it
	fc[location] = buff
	return buff
}
