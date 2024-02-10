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
	emuDeleted = []string{}
)

func ResetEmulator() {
	ResetFixtures()
	emuCreated = []interface{}{}
	emuDeleted = []string{}
}

func EmuDeleted() []string {
	return emuDeleted
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
		l := location + "/" + id
		log.Printf("[TRACE] emuGet(%s)", l)
		buff := fixtureGet(l)
		w.WriteHeader(200)
		w.Write(buff) // nolint
	}
}

func emuPost(location string, config EmuConfig, empty bool) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Printf("[TRACE] emuPost(%s)", location)
		in := config.Types[location].Factory()
		unmarshallRequestBody(r, in)
		id, out := config.Types[location].Responder(in)
		l := location + "/" + id
		PostFixture(l, out)
		emuCreated = append(emuCreated, out)
		w.WriteHeader(200)
		if !empty {
			w.Write(fc[l]) // nolint
		}
	}
}

func emuDelete(location string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")
		l := location + "/" + id
		log.Printf("[TRACE] emuDelete(%s)", l)
		if fixtureDelete(l) {
			w.WriteHeader(204)
		} else {
			w.WriteHeader(404)
		}
		emuDeleted = append(emuDeleted, l)
	}
}

func emuNotFound(res http.ResponseWriter, req *http.Request) {
	m := req.Method
	path := req.URL.Path
	if m == "" {
		m = "GET"
	}
	log.Printf("[TRACE] emuNotFound(%s %s)", m, path)
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
	router.POST("/admin/AddTenant", emuPost("tenant", config, true))
	router.POST("/admin/DeleteTenant/:id", emuDelete("tenant"))

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
