package duplosdktest

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
)

type fixtureCache map[string][]byte

type FixturePatcher func()

var (
	fc   = fixtureCache{}
	fdir string
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fdir = path.Join(dir, "../internal/duplosdktest/fixtures")
}

func ResetFixtures() {
	fc = fixtureCache{}
}

func PatchFixture(location string, target interface{}, patcher FixturePatcher) {
	body := fixtureGet(location)
	err := json.Unmarshal(body, target)
	if err != nil {
		log.Panicf("json.Unmarshal: %s: %s", location, err)
	}

	patcher()

	body, err = json.Marshal(target)
	if err != nil {
		log.Panicf("json.Marshal: %s: %s", location, err)
	}

	fc[location] = body

	// Invalidate any cached list
	delete(fc, path.Dir(location))
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
			buff := fixtureList("tenant")
			res.WriteHeader(200)
			res.Write(buff) // nolint
		}

		res.WriteHeader(404)
		res.Write([]byte("")) // nolint
	}))
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

	// Cache all constituent elements, and calculate our final size.
	size := len(files) + 1 // 2 brackets, and commas are (length - 1)
	buffs := make([][]byte, len(files))
	for i := range files {
		name := files[i].Name()

		if strings.HasSuffix(name, ".json") {
			buffs[i] = fixtureGet(path.Join(location, strings.TrimSuffix(name, ".json")))
		}
		size += len(buffs[i])
	}

	// fill the buffer
	buff := make([]byte, 0, size)
	buff = append(buff, []byte("[")...)
	for i := range files {
		if i > 0 {
			buff = append(buff, []byte(",")...)
		}
		buff = append(buff, buffs[i]...)
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
