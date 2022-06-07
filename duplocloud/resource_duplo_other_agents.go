package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func duploOtherAgentsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description: "Resource name to create other agents in duplo.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},

		"agent": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"agent_name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"agent_linux_package_path": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"agent_windows_package_path": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"linux_agent_install_status_cmd": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"linux_agent_service_name": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"linux_agent_uninstall_status_cmd": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"linux_install_cmd": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"windows_agent_service_name": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"user_request_reset_is_pending": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"execution_count": {
						Type:     schema.TypeInt,
						Computed: true,
					},
				},
			},
		},
	}
}

func resourceOtherAgents() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_other_agents` manages an other agents in Duplo.",

		ReadContext:   resourceOtherAgentsRead,
		CreateContext: resourceOtherAgentsCreate,
		UpdateContext: resourceOtherAgentsUpdate,
		DeleteContext: resourceOtherAgentsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploOtherAgentsSchema(),
	}
}

func resourceOtherAgentsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	name := id
	log.Printf("[TRACE] resourceOtherAgentsRead  (%s): start", name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DuploOtherAgentGet()
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve other agents %s : %s", name, clientErr)
	}

	// TODO Set ccomputed attributes from duplo object to tf state.
	d.Set("name", name)
	d.Set("agent", flattenOtherAgents(duplo))

	log.Printf("[TRACE] resourceOtherAgentsRead(%s): end", name)
	return nil
}

func resourceOtherAgentsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceOtherAgentsCreate(%s): start", name)
	c := m.(*duplosdk.Client)

	rq := expandOtherAgents(d.Get("agent").([]interface{}))
	err = c.DuploOtherAgentCreate(rq)
	if err != nil {
		return diag.Errorf("Error creating other agents '%s': %s", name, err)
	}

	id := name
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "other agents", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploOtherAgentGet()
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceOtherAgentsRead(ctx, d, m)
	log.Printf("[TRACE] resourceOtherAgentsCreate(%s): end", name)
	return diags
}

func resourceOtherAgentsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceOtherAgentsCreate(ctx, d, m)
}

func resourceOtherAgentsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	name := id
	log.Printf("[TRACE] resourceOtherAgentsDelete(%s): start", name)

	c := m.(*duplosdk.Client)
	clientErr := c.DuploOtherAgentDelete()
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete other agents '%s': %s", name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "other agents", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.DuploOtherAgentExists(); rp || err != nil {
			return rp, err
		}
		return nil, nil
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceOtherAgentsDelete(%s): end", name)
	return nil
}

func expandOtherAgent(m map[string]interface{}) duplosdk.DuploDuploOtherAgentReq {
	req := duplosdk.DuploDuploOtherAgentReq{
		AgentName:                    m["agent_name"].(string),
		AgentLinuxPackagePath:        m["agent_linux_package_path"].(string),
		AgentWindowsPackagePath:      m["agent_windows_package_path"].(string),
		LinuxAgentInstallStatusCmd:   m["linux_agent_install_status_cmd"].(string),
		LinuxAgentServiceName:        m["linux_agent_service_name"].(string),
		LinuxAgentUninstallStatusCmd: m["linux_agent_uninstall_status_cmd"].(string),
		LinuxInstallCmd:              m["linux_install_cmd"].(string),
		WindowsAgentServiceName:      m["windows_agent_service_name"].(string),
	}
	return req
}

func expandOtherAgents(lst []interface{}) *[]duplosdk.DuploDuploOtherAgentReq {
	var qty int
	items := make([]duplosdk.DuploDuploOtherAgentReq, 0, len(lst))
	//var items []duplosdk.DuploAwsCloudfrontCacheBehavior
	for _, v := range lst {
		items = append(items, expandOtherAgent(v.(map[string]interface{})))
		qty++
	}
	return &items
}
func flattenOtherAgents(duplo *[]duplosdk.DuploDuploOtherAgent) []interface{} {

	lst := []interface{}{}
	for _, v := range *duplo {
		lst = append(lst, flattenOtherAgent(v))
	}
	return lst
}
func flattenOtherAgent(duplo duplosdk.DuploDuploOtherAgent) map[string]interface{} {
	m := map[string]interface{}{
		"agent_name":                       duplo.AgentName,
		"agent_linux_package_path":         duplo.AgentLinuxPackagePath,
		"agent_windows_package_path":       duplo.AgentWindowsPackagePath,
		"linux_agent_install_status_cmd":   duplo.LinuxAgentInstallStatusCmd,
		"linux_agent_service_name":         duplo.LinuxAgentServiceName,
		"linux_agent_uninstall_status_cmd": duplo.LinuxAgentUninstallStatusCmd,
		"linux_install_cmd":                duplo.LinuxInstallCmd,
		"windows_agent_service_name":       duplo.WindowsAgentServiceName,
		"user_request_reset_is_pending":    duplo.UserRequestResetIsPending,
		"execution_count":                  duplo.ExecutionCount,
	}
	return m
}
