package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// GlueResource is a free-form JSON object (JObject passthrough).
// The DuploCloud backend accepts and returns AWS Glue payloads as-is,
// only enforcing the Name/Role/Type typing on inputs; the provider
// layer carries the rest through this map.
type GlueResource = map[string]interface{}

// gluePath returns a path under /v3/subscriptions/{tenantID}/aws/glue/.
func gluePath(tenantID, suffix string) string {
	return fmt.Sprintf("v3/subscriptions/%s/aws/glue/%s", tenantID, suffix)
}

// glueGet fetches a single Glue resource.
func (c *Client) glueGet(apiName, tenantID, suffix string) (GlueResource, ClientError) {
	rp := GlueResource{}
	err := c.getAPI(apiName, gluePath(tenantID, suffix), &rp)
	if err != nil {
		if err.Status() == 404 {
			return nil, nil
		}
		return nil, err
	}
	return rp, nil
}

// glueList fetches a list of Glue resources.
func (c *Client) glueList(apiName, tenantID, suffix string) ([]GlueResource, ClientError) {
	rp := []GlueResource{}
	err := c.getAPI(apiName, gluePath(tenantID, suffix), &rp)
	if err != nil {
		return nil, err
	}
	return rp, nil
}

// glueCreate POSTs a new Glue resource. Empty response bodies are tolerated
// because several Glue create endpoints (Database, Connection, Table, etc.)
// return no content; the resource layer always re-reads to populate state.
func (c *Client) glueCreate(apiName, tenantID, suffix string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueWriteWithEmptyOK("POST", apiName, gluePath(tenantID, suffix), rq)
}

// glueUpdate PUTs an existing Glue resource. Tolerates empty responses for
// the same reason as glueCreate.
func (c *Client) glueUpdate(apiName, tenantID, suffix string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueWriteWithEmptyOK("PUT", apiName, gluePath(tenantID, suffix), rq)
}

// glueWriteWithEmptyOK runs a POST or PUT against a Glue endpoint and parses
// the response body as a JObject, treating an empty/null body as success with
// an empty map.
func (c *Client) glueWriteWithEmptyOK(verb, apiName, apiPath string, rq GlueResource) (GlueResource, ClientError) {
	url := fmt.Sprintf("%s/%s", c.HostURL, apiPath)
	rqBody, err := json.Marshal(rq)
	if err != nil {
		message := fmt.Sprintf("%sAPI %s: cannot marshal request to JSON: %s", strings.ToLower(verb), apiName, err.Error())
		return nil, requestHttpError(url, message)
	}
	log.Printf("[TRACE] %sAPI %s: prepared request: %s", strings.ToLower(verb), apiName, url)
	req, herr := http.NewRequest(verb, url, strings.NewReader(string(rqBody)))
	if herr != nil {
		log.Printf("[TRACE] %sAPI %s: cannot build request: %s", strings.ToLower(verb), apiName, herr.Error())
		return nil, nil
	}
	body, httpErr := c.doRequest(req)
	if httpErr != nil {
		return nil, httpErr
	}
	bodyString := strings.TrimSpace(string(body))
	log.Printf("[TRACE] %sAPI %s: received response: %s", strings.ToLower(verb), apiName, bodyString)
	if bodyString == "" || bodyString == "null" || bodyString == "\"\"" {
		return GlueResource{}, nil
	}
	out := GlueResource{}
	if jerr := json.Unmarshal(body, &out); jerr != nil {
		message := fmt.Sprintf("%sAPI %s: cannot unmarshal response from JSON: %s", strings.ToLower(verb), apiName, jerr.Error())
		return nil, appHttpError(req, message)
	}
	return out, nil
}

// glueDelete deletes a Glue resource. Returns nil on 404. The response body
// is ignored so empty (and varying) success responses don't break parsing.
func (c *Client) glueDelete(apiName, tenantID, suffix string) ClientError {
	url := fmt.Sprintf("%s/%s", c.HostURL, gluePath(tenantID, suffix))
	log.Printf("[TRACE] deleteAPI %s: prepared request: %s", apiName, url)
	req, herr := http.NewRequest("DELETE", url, nil)
	if herr != nil {
		log.Printf("[TRACE] deleteAPI %s: cannot build request: %s", apiName, herr.Error())
		return nil
	}
	body, httpErr := c.doRequest(req)
	if httpErr != nil {
		if httpErr.Status() == 404 {
			return nil
		}
		return httpErr
	}
	log.Printf("[TRACE] deleteAPI %s: received response: %s", apiName, strings.TrimSpace(string(body)))
	return nil
}

// ---- Connections ----

func (c *Client) AwsGlueConnectionList(tenantID string) ([]GlueResource, ClientError) {
	return c.glueList(fmt.Sprintf("AwsGlueConnectionList(%s)", tenantID), tenantID, "connections")
}

func (c *Client) AwsGlueConnectionGet(tenantID, name string) (GlueResource, ClientError) {
	return c.glueGet(fmt.Sprintf("AwsGlueConnectionGet(%s, %s)", tenantID, name), tenantID, "connections/"+name)
}

func (c *Client) AwsGlueConnectionCreate(tenantID string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueCreate(fmt.Sprintf("AwsGlueConnectionCreate(%s)", tenantID), tenantID, "connections", rq)
}

func (c *Client) AwsGlueConnectionUpdate(tenantID, name string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueUpdate(fmt.Sprintf("AwsGlueConnectionUpdate(%s, %s)", tenantID, name), tenantID, "connections/"+name, rq)
}

func (c *Client) AwsGlueConnectionDelete(tenantID, name string) ClientError {
	return c.glueDelete(fmt.Sprintf("AwsGlueConnectionDelete(%s, %s)", tenantID, name), tenantID, "connections/"+name)
}

// ---- Crawlers ----

func (c *Client) AwsGlueCrawlerList(tenantID string) ([]GlueResource, ClientError) {
	return c.glueList(fmt.Sprintf("AwsGlueCrawlerList(%s)", tenantID), tenantID, "crawlers")
}

func (c *Client) AwsGlueCrawlerGet(tenantID, name string) (GlueResource, ClientError) {
	return c.glueGet(fmt.Sprintf("AwsGlueCrawlerGet(%s, %s)", tenantID, name), tenantID, "crawlers/"+name)
}

func (c *Client) AwsGlueCrawlerCreate(tenantID string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueCreate(fmt.Sprintf("AwsGlueCrawlerCreate(%s)", tenantID), tenantID, "crawlers", rq)
}

func (c *Client) AwsGlueCrawlerUpdate(tenantID, name string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueUpdate(fmt.Sprintf("AwsGlueCrawlerUpdate(%s, %s)", tenantID, name), tenantID, "crawlers/"+name, rq)
}

func (c *Client) AwsGlueCrawlerDelete(tenantID, name string) ClientError {
	return c.glueDelete(fmt.Sprintf("AwsGlueCrawlerDelete(%s, %s)", tenantID, name), tenantID, "crawlers/"+name)
}

// ---- Jobs ----

func (c *Client) AwsGlueJobList(tenantID string) ([]GlueResource, ClientError) {
	return c.glueList(fmt.Sprintf("AwsGlueJobList(%s)", tenantID), tenantID, "jobs")
}

func (c *Client) AwsGlueJobGet(tenantID, name string) (GlueResource, ClientError) {
	return c.glueGet(fmt.Sprintf("AwsGlueJobGet(%s, %s)", tenantID, name), tenantID, "jobs/"+name)
}

func (c *Client) AwsGlueJobCreate(tenantID string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueCreate(fmt.Sprintf("AwsGlueJobCreate(%s)", tenantID), tenantID, "jobs", rq)
}

func (c *Client) AwsGlueJobUpdate(tenantID, name string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueUpdate(fmt.Sprintf("AwsGlueJobUpdate(%s, %s)", tenantID, name), tenantID, "jobs/"+name, rq)
}

func (c *Client) AwsGlueJobDelete(tenantID, name string) ClientError {
	return c.glueDelete(fmt.Sprintf("AwsGlueJobDelete(%s, %s)", tenantID, name), tenantID, "jobs/"+name)
}

// ---- Triggers ----

func (c *Client) AwsGlueTriggerList(tenantID string) ([]GlueResource, ClientError) {
	return c.glueList(fmt.Sprintf("AwsGlueTriggerList(%s)", tenantID), tenantID, "triggers")
}

func (c *Client) AwsGlueTriggerGet(tenantID, name string) (GlueResource, ClientError) {
	return c.glueGet(fmt.Sprintf("AwsGlueTriggerGet(%s, %s)", tenantID, name), tenantID, "triggers/"+name)
}

func (c *Client) AwsGlueTriggerCreate(tenantID string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueCreate(fmt.Sprintf("AwsGlueTriggerCreate(%s)", tenantID), tenantID, "triggers", rq)
}

func (c *Client) AwsGlueTriggerUpdate(tenantID, name string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueUpdate(fmt.Sprintf("AwsGlueTriggerUpdate(%s, %s)", tenantID, name), tenantID, "triggers/"+name, rq)
}

func (c *Client) AwsGlueTriggerDelete(tenantID, name string) ClientError {
	return c.glueDelete(fmt.Sprintf("AwsGlueTriggerDelete(%s, %s)", tenantID, name), tenantID, "triggers/"+name)
}

// ---- Workflows ----

func (c *Client) AwsGlueWorkflowList(tenantID string) ([]GlueResource, ClientError) {
	return c.glueList(fmt.Sprintf("AwsGlueWorkflowList(%s)", tenantID), tenantID, "workflows")
}

func (c *Client) AwsGlueWorkflowGet(tenantID, name string) (GlueResource, ClientError) {
	return c.glueGet(fmt.Sprintf("AwsGlueWorkflowGet(%s, %s)", tenantID, name), tenantID, "workflows/"+name)
}

func (c *Client) AwsGlueWorkflowCreate(tenantID string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueCreate(fmt.Sprintf("AwsGlueWorkflowCreate(%s)", tenantID), tenantID, "workflows", rq)
}

func (c *Client) AwsGlueWorkflowUpdate(tenantID, name string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueUpdate(fmt.Sprintf("AwsGlueWorkflowUpdate(%s, %s)", tenantID, name), tenantID, "workflows/"+name, rq)
}

func (c *Client) AwsGlueWorkflowDelete(tenantID, name string) ClientError {
	return c.glueDelete(fmt.Sprintf("AwsGlueWorkflowDelete(%s, %s)", tenantID, name), tenantID, "workflows/"+name)
}

// ---- Databases ----

func (c *Client) AwsGlueDatabaseList(tenantID string) ([]GlueResource, ClientError) {
	return c.glueList(fmt.Sprintf("AwsGlueDatabaseList(%s)", tenantID), tenantID, "databases")
}

func (c *Client) AwsGlueDatabaseGet(tenantID, name string) (GlueResource, ClientError) {
	return c.glueGet(fmt.Sprintf("AwsGlueDatabaseGet(%s, %s)", tenantID, name), tenantID, "databases/"+name)
}

func (c *Client) AwsGlueDatabaseCreate(tenantID string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueCreate(fmt.Sprintf("AwsGlueDatabaseCreate(%s)", tenantID), tenantID, "databases", rq)
}

func (c *Client) AwsGlueDatabaseUpdate(tenantID, name string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueUpdate(fmt.Sprintf("AwsGlueDatabaseUpdate(%s, %s)", tenantID, name), tenantID, "databases/"+name, rq)
}

func (c *Client) AwsGlueDatabaseDelete(tenantID, name string) ClientError {
	return c.glueDelete(fmt.Sprintf("AwsGlueDatabaseDelete(%s, %s)", tenantID, name), tenantID, "databases/"+name)
}

// ---- Tables (nested under Databases) ----

func (c *Client) AwsGlueTableList(tenantID, databaseName string) ([]GlueResource, ClientError) {
	suffix := fmt.Sprintf("databases/%s/tables", databaseName)
	return c.glueList(fmt.Sprintf("AwsGlueTableList(%s, %s)", tenantID, databaseName), tenantID, suffix)
}

func (c *Client) AwsGlueTableGet(tenantID, databaseName, name string) (GlueResource, ClientError) {
	suffix := fmt.Sprintf("databases/%s/tables/%s", databaseName, name)
	return c.glueGet(fmt.Sprintf("AwsGlueTableGet(%s, %s, %s)", tenantID, databaseName, name), tenantID, suffix)
}

func (c *Client) AwsGlueTableCreate(tenantID, databaseName string, rq GlueResource) (GlueResource, ClientError) {
	suffix := fmt.Sprintf("databases/%s/tables", databaseName)
	return c.glueCreate(fmt.Sprintf("AwsGlueTableCreate(%s, %s)", tenantID, databaseName), tenantID, suffix, rq)
}

func (c *Client) AwsGlueTableUpdate(tenantID, databaseName, name string, rq GlueResource) (GlueResource, ClientError) {
	suffix := fmt.Sprintf("databases/%s/tables/%s", databaseName, name)
	return c.glueUpdate(fmt.Sprintf("AwsGlueTableUpdate(%s, %s, %s)", tenantID, databaseName, name), tenantID, suffix, rq)
}

func (c *Client) AwsGlueTableDelete(tenantID, databaseName, name string) ClientError {
	suffix := fmt.Sprintf("databases/%s/tables/%s", databaseName, name)
	return c.glueDelete(fmt.Sprintf("AwsGlueTableDelete(%s, %s, %s)", tenantID, databaseName, name), tenantID, suffix)
}

// ---- Registries ----

func (c *Client) AwsGlueRegistryList(tenantID string) ([]GlueResource, ClientError) {
	return c.glueList(fmt.Sprintf("AwsGlueRegistryList(%s)", tenantID), tenantID, "registries")
}

func (c *Client) AwsGlueRegistryGet(tenantID, name string) (GlueResource, ClientError) {
	return c.glueGet(fmt.Sprintf("AwsGlueRegistryGet(%s, %s)", tenantID, name), tenantID, "registries/"+name)
}

func (c *Client) AwsGlueRegistryCreate(tenantID string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueCreate(fmt.Sprintf("AwsGlueRegistryCreate(%s)", tenantID), tenantID, "registries", rq)
}

func (c *Client) AwsGlueRegistryUpdate(tenantID, name string, rq GlueResource) (GlueResource, ClientError) {
	return c.glueUpdate(fmt.Sprintf("AwsGlueRegistryUpdate(%s, %s)", tenantID, name), tenantID, "registries/"+name, rq)
}

func (c *Client) AwsGlueRegistryDelete(tenantID, name string) ClientError {
	return c.glueDelete(fmt.Sprintf("AwsGlueRegistryDelete(%s, %s)", tenantID, name), tenantID, "registries/"+name)
}

// ---- Schemas (nested under Registries) ----

func (c *Client) AwsGlueSchemaList(tenantID, registryName string) ([]GlueResource, ClientError) {
	suffix := fmt.Sprintf("registries/%s/schemas", registryName)
	return c.glueList(fmt.Sprintf("AwsGlueSchemaList(%s, %s)", tenantID, registryName), tenantID, suffix)
}

func (c *Client) AwsGlueSchemaGet(tenantID, registryName, name string) (GlueResource, ClientError) {
	suffix := fmt.Sprintf("registries/%s/schemas/%s", registryName, name)
	return c.glueGet(fmt.Sprintf("AwsGlueSchemaGet(%s, %s, %s)", tenantID, registryName, name), tenantID, suffix)
}

func (c *Client) AwsGlueSchemaCreate(tenantID, registryName string, rq GlueResource) (GlueResource, ClientError) {
	suffix := fmt.Sprintf("registries/%s/schemas", registryName)
	return c.glueCreate(fmt.Sprintf("AwsGlueSchemaCreate(%s, %s)", tenantID, registryName), tenantID, suffix, rq)
}

func (c *Client) AwsGlueSchemaUpdate(tenantID, registryName, name string, rq GlueResource) (GlueResource, ClientError) {
	suffix := fmt.Sprintf("registries/%s/schemas/%s", registryName, name)
	return c.glueUpdate(fmt.Sprintf("AwsGlueSchemaUpdate(%s, %s, %s)", tenantID, registryName, name), tenantID, suffix, rq)
}

func (c *Client) AwsGlueSchemaDelete(tenantID, registryName, name string) ClientError {
	suffix := fmt.Sprintf("registries/%s/schemas/%s", registryName, name)
	return c.glueDelete(fmt.Sprintf("AwsGlueSchemaDelete(%s, %s, %s)", tenantID, registryName, name), tenantID, suffix)
}

// AwsGlueSchemaVersionList returns the list of versions for a schema.
func (c *Client) AwsGlueSchemaVersionList(tenantID, registryName, schemaName string) ([]GlueResource, ClientError) {
	suffix := fmt.Sprintf("registries/%s/schemas/%s/versions", registryName, schemaName)
	return c.glueList(fmt.Sprintf("AwsGlueSchemaVersionList(%s, %s, %s)", tenantID, registryName, schemaName), tenantID, suffix)
}

// AwsGlueSchemaVersionGet fetches a single schema version (including SchemaDefinition).
func (c *Client) AwsGlueSchemaVersionGet(tenantID, registryName, schemaName, versionID string) (GlueResource, ClientError) {
	suffix := fmt.Sprintf("registries/%s/schemas/%s/versions/%s", registryName, schemaName, versionID)
	return c.glueGet(fmt.Sprintf("AwsGlueSchemaVersionGet(%s, %s, %s, %s)", tenantID, registryName, schemaName, versionID), tenantID, suffix)
}
