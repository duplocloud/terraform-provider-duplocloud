package duplosdktest

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/julienschmidt/httprouter"
)

type EmuResponder func(in interface{}) (id string, out interface{})
type EmuFactory func() interface{}
type EmuType struct {
	Factory   EmuFactory
	Responder EmuResponder
}
type EmuConfig struct {
	Types map[string]EmuType
}

var (
	emuCreated = []interface{}{}
)

func ResetEmulator() {
	ResetFixtures()
	emuCreated = []interface{}{}
}

func EmuCreated() []interface{} {
	return emuCreated
}

func emuList(location string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		buff := fixtureList(location)
		w.WriteHeader(200)
		w.Write(buff) // nolint
	}
}

func emuGet(location string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")
		buff := fixtureGet(location + "/" + id)
		w.WriteHeader(200)
		w.Write(buff) // nolint
	}
}

func emuCreate(location string, config EmuConfig, empty bool) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		in := config.Types[location].Factory()
		unmarshallRequestBody(r, in)
		id, out := config.Types[location].Responder(in)
		location += "/" + id
		PostFixture(location, out)
		emuCreated = append(emuCreated, out)
		w.WriteHeader(200)
		if !empty {
			w.Write(fc[location]) // nolint
		}
	}
}

func emuNotFound(res http.ResponseWriter, req *http.Request) {
	m := req.Method
	path := req.URL.Path
	if m == "" {
		m = "GET"
	}
	if m == "GET" {
		res.WriteHeader(404)
	} else {
		res.WriteHeader(599)
	}
	res.Write([]byte("No test-case handler for " + m + " " + path)) // nolint
}

func NewEmulator(config EmuConfig) *httptest.Server {
	fc = fixtureCache{}

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(emuNotFound)

	// tenant API emulation
	router.GET("/admin/GetTenantsForUser", emuList("tenant"))
	router.GET("/v2/admin/TenantV2", emuList("tenant"))
	router.GET("/v2/admin/TenantV2/:id", emuGet("tenant"))
	router.GET("/v3/admin/tenant", emuList("tenant"))
	router.GET("/v3/admin/tenant/:id", emuGet("tenant"))
	router.POST("/admin/AddTenant", emuCreate("tenant", config, true))

	return SetupHttptest(router)
}

func unmarshallRequestBody(req *http.Request, target interface{}) {
	body, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		log.Panicf("req.ReadBody: %v: %s", req.URL.Path, err)
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		log.Panicf("json.Unmarshal: %s: %s", req.URL.Path, err)
	}
}
