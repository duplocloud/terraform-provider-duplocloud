package duplosdktest

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"

	"github.com/julienschmidt/httprouter"
)

type EmuResponder func(verb string, in interface{}) (id string, out interface{})
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
	emuParams  = regexp.MustCompile(":[^/]+")
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

func EmuLastCreated() interface{} {
	return emuCreated[len(emuCreated)-1]
}

func emuLocation(location string, ps httprouter.Params) string {
	return emuParams.ReplaceAllStringFunc(location, func(match string) string {
		return ps.ByName(match[1:])
	})
}

func emuList(location string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		l := emuLocation(location, ps)
		log.Printf("[TRACE] emuList(%s)", l)
		buff := ListFixtures(l)
		w.WriteHeader(200)
		w.Write(buff) // nolint
	}
}

func emuGet(location, idKey string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName(idKey)
		l := emuLocation(location, ps) + "/" + id
		log.Printf("[TRACE] emuGet(%s)", l)
		buff := GetFixture(l)
		w.WriteHeader(200)
		w.Write(buff) // nolint
	}
}

func emuWrite(logPrefix, location string, config EmuConfig, hasId, empty bool) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		in := config.Types[location].Factory()
		l := emuLocation(location, ps)
		log.Printf("[TRACE] emu%s(%s)", logPrefix, l)
		unmarshallRequestBody(r, in)
		id, out := config.Types[location].Responder(r.Method, in)
		l += "/" + id
		PostFixture(l, out)
		emuCreated = append(emuCreated, out)
		w.WriteHeader(200)
		if !empty {
			w.Write(fc[l]) // nolint
		}
	}
}

func emuPut(location string, config EmuConfig, empty bool) httprouter.Handle {
	return emuWrite("Put", location, config, true, empty)
}

func emuPost(location string, config EmuConfig, empty bool) httprouter.Handle {
	return emuWrite("Post", location, config, false, empty)
}

func emuDelete(location, idKey string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName(idKey)
		l := emuLocation(location, ps) + "/" + id
		log.Printf("[TRACE] emuDelete(%s)", l)
		if DeleteFixture(l) {
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

	// tenant APIs
	router.GET("/admin/GetTenantsForUser", emuList("tenant"))
	router.GET("/v2/admin/TenantV2", emuList("tenant"))
	router.GET("/v2/admin/TenantV2/:tenantId", emuGet("tenant", "tenantId"))
	router.GET("/v3/admin/tenant", emuList("tenant"))
	router.GET("/v3/admin/tenant/:tenantId", emuGet("tenant", "tenantId"))
	router.POST("/admin/AddTenant", emuPost("tenant", config, true))
	router.POST("/admin/DeleteTenant/:tenantId", emuDelete("tenant", "tenantId"))
	router.GET("/v3/admin/tenant/:tenantId/metadata", emuList("tenant/:tenantId/metadata"))
	router.GET("/v3/admin/tenant/:tenantId/metadata/:key", emuGet("tenant/:tenantId/metadata", "key"))
	router.POST("/v3/admin/tenant/:tenantId/metadata", emuPost("tenant/:tenantId/metadata", config, false))
	router.PUT("/v3/admin/tenant/:tenantId/metadata/:key", emuPut("tenant/:tenantId/metadata", config, false))
	router.DELETE("/v3/admin/tenant/:tenantId/metadata/:key", emuDelete("tenant/:tenantId/metadata", "key"))

	// non-admin tenant APIs
	router.GET("/subscriptions/:tenantId/GetExternalSubnets", emuList("tenant/:tenantId/external_subnets"))
	router.GET("/subscriptions/:tenantId/GetInternalSubnets", emuList("tenant/:tenantId/internal_subnets"))

	// AWS host APIs
	router.GET("/v2/subscriptions/:tenantId/NativeHostV2", emuList("tenant/:tenantId/aws_host"))
	router.GET("/v2/subscriptions/:tenantId/NativeHostV2/:id", emuGet("tenant/:tenantId/aws_host", "id"))
	router.POST("/v2/subscriptions/:tenantId/NativeHostV2", emuPost("tenant/:tenantId/aws_host", config, false))
	router.DELETE("/v2/subscriptions/:tenantId/NativeHostV2/:id", emuDelete("tenant/:tenantId/aws_host", "id"))

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
