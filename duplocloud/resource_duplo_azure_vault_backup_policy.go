package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureVaultBackupPolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"infra_name": {
			Description:  "The name of the infrastructure.  Infrastructure names are globally unique and less than 13 characters.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringLenBetween(2, 12),
		},
		"name": {
			Description: "Specifies the name of the vault backup policy.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"instant_restore_retention_days": {
			Description:  "Specifies the instant restore retention range in days.",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(1, 30),
		},
		"timezone": {
			Description: "Specifies the timezone.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "UTC",
		},
		"policy_type": {
			Description: "Type of the Backup Policy.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Default:     "V1",
			ValidateFunc: validation.StringInSlice([]string{
				"V1",
				"V2",
			}, false),
		},
		"backup": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"frequency": {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							"Hourly",
							"Daily",
							"Weekly",
						}, false),
					},

					"time": { // applies to all backup schedules & retention times (they all must be the same)
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringMatch(
							regexp.MustCompile("^([01][0-9]|[2][0-3]):([03][0])$"), // time must be on the hour or half past
							"Time of day must match the format HH:mm where HH is 00-23 and mm is 00 or 30",
						),
					},

					"weekdays": { // only for weekly
						Type:     schema.TypeSet,
						Optional: true,
						Set:      HashStringIgnoreCase,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							DiffSuppressFunc: CaseDifference,
							ValidateFunc:     validation.IsDayOfTheWeek(true),
						},
					},

					"hour_interval": {
						Type:     schema.TypeInt,
						Optional: true,
						ValidateFunc: validation.IntInSlice([]int{
							4,
							6,
							8,
							12,
						}),
					},

					"hour_duration": {
						Type:         schema.TypeInt,
						Optional:     true,
						ValidateFunc: validation.IntBetween(4, 24),
					},
				},
			},
		},
		"retention_daily": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"count": {
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 9999), // Azure no longer supports less than 7 daily backups. This should be updated in 3.0 provider

					},
				},
			},
		},
		"retention_weekly": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"count": {
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 9999),
					},

					"weekdays": {
						Type:     schema.TypeSet,
						Required: true,
						Set:      HashStringIgnoreCase,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							DiffSuppressFunc: CaseDifference,
							ValidateFunc:     validation.IsDayOfTheWeek(true),
						},
					},
				},
			},
		},

		"retention_monthly": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"count": {
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 9999),
					},

					"weeks": {
						Type:     schema.TypeSet,
						Required: true,
						Set:      HashStringIgnoreCase,
						Elem: &schema.Schema{
							Type: schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{
								"First",
								"Second",
								"Third",
								"Fourth",
								"Last",
							}, false),
						},
					},

					"weekdays": {
						Type:     schema.TypeSet,
						Required: true,
						Set:      HashStringIgnoreCase,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							DiffSuppressFunc: CaseDifference,
							ValidateFunc:     validation.IsDayOfTheWeek(true),
						},
					},
				},
			},
		},

		"retention_yearly": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"count": {
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 9999),
					},

					"months": {
						Type:     schema.TypeSet,
						Required: true,
						Set:      HashStringIgnoreCase,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							DiffSuppressFunc: CaseDifference,
							ValidateFunc:     validation.IsMonth(true),
						},
					},

					"weeks": {
						Type:     schema.TypeSet,
						Required: true,
						Set:      HashStringIgnoreCase,
						Elem: &schema.Schema{
							Type: schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{
								"First",
								"Second",
								"Third",
								"Fourth",
								"Last",
							}, false),
						},
					},

					"weekdays": {
						Type:     schema.TypeSet,
						Required: true,
						Set:      HashStringIgnoreCase,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							DiffSuppressFunc: CaseDifference,
							ValidateFunc:     validation.IsDayOfTheWeek(true),
						},
					},
				},
			},
		},
		"azure_id": {
			Description: "Azure id for vault backup policy.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAzureVaultBackupPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_vault_backup_policy` manages a Vault Backup Policy in Duplo.",

		ReadContext:   resourceAzureVaultBackupPolicyRead,
		CreateContext: resourceAzureVaultBackupPolicyCreate,
		UpdateContext: resourceAzureVaultBackupPolicyUpdate,
		DeleteContext: resourceAzureVaultBackupPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureVaultBackupPolicySchema(),
	}
}

func resourceAzureVaultBackupPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	infraName, name, err := parseAzureVaultBackupPolicyIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureVaultBackupPolicyRead(%s, %s): start", infraName, name)

	c := m.(*duplosdk.Client)
	policy, clientErr := c.VaultBackupPolicyGet(infraName, name)
	if policy == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve infra %s azure vault backup policy %s : %s", infraName, name, clientErr)
	}

	err = flattenAzureVaultBackupPolicy(infraName, d, policy)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceAzureVaultBackupPolicyRead(%s, %s): end", infraName, name)
	return nil
}

func resourceAzureVaultBackupPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	infraName := d.Get("infra_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureVaultBackupPolicyCreate(%s, %s): start", infraName, name)
	c := m.(*duplosdk.Client)

	rq, err := expandAzureVaultBackupPolicy(d)
	if err != nil {
		return diag.Errorf("Error expanding infra %s azure vault backup policy '%s': %s", infraName, name, err)
	}
	err = c.VaultBackupPolicyCreate(infraName, rq)
	if err != nil {
		return diag.Errorf("Error creating infra %s azure vault backup policy '%s': %s", infraName, name, err)
	}

	id := fmt.Sprintf("%s/%s", infraName, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure vault backup policy", id, func() (interface{}, duplosdk.ClientError) {
		return c.VaultBackupPolicyGet(infraName, name)
	})
	if diags != nil {
		return diags
	}

	d.SetId(id)

	diags = resourceAzureVaultBackupPolicyRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureVaultBackupPolicyCreate(%s, %s): end", infraName, name)
	return diags
}

func resourceAzureVaultBackupPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureVaultBackupPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	infraName, name, err := parseAzureVaultBackupPolicyIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureVaultBackupPolicyDelete(%s, %s): start", infraName, name)

	c := m.(*duplosdk.Client)
	rq := &duplosdk.DuploAzureVaultBackupPolicy{
		Name: name,
	}
	clientErr := c.VaultBackupPolicyDelete(infraName, rq)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure vault backup policy '%s': %s", infraName, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure vault backup policy", id, func() (interface{}, duplosdk.ClientError) {
		return c.VaultBackupPolicyGet(infraName, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureVaultBackupPolicyDelete(%s, %s): end", infraName, name)
	return nil
}

func expandAzureVaultBackupPolicy(d *schema.ResourceData) (*duplosdk.DuploAzureVaultBackupPolicy, error) {
	policyName := d.Get("name").(string)
	infraName := d.Get("infra_name").(string)
	timeOfDay := d.Get("backup.0.time").(string)
	dateOfDay, err := time.Parse(time.RFC3339, fmt.Sprintf("2018-07-30T%s:00Z", timeOfDay))
	if err != nil {
		return nil, fmt.Errorf("generating time from %q for policy %q (Infra %q): %+v", timeOfDay, policyName, infraName, err)
	}
	times := append(make([]time.Time, 0), dateOfDay)

	schedulePolicy, err := expandBackupProtectionPolicyVMSchedule(d, times)
	if err != nil {
		return nil, err
	}
	properties := &duplosdk.DuploAzureVaultBackupPolicyProperties{
		SchedulePolicy:       schedulePolicy,
		PolicyType:           d.Get("policy_type").(string),
		TimeZone:             d.Get("timezone").(string),
		BackupManagementType: "AzureIaasVM",
		RetentionPolicy: &duplosdk.DuploAzureVaultBackupRetentionPolicy{ // SimpleRetentionPolicy only has duration property ¯\_(ツ)_/¯
			RetentionPolicyType: "LongTermRetentionPolicy",
			DailySchedule:       expandBackupProtectionPolicyVMRetentionDaily(d, times),
			WeeklySchedule:      expandBackupProtectionPolicyVMRetentionWeekly(d, times),
			MonthlySchedule:     expandBackupProtectionPolicyVMRetentionMonthly(d, times),
			YearlySchedule:      expandBackupProtectionPolicyVMRetentionYearly(d, times),
		},
	}
	return &duplosdk.DuploAzureVaultBackupPolicy{
		Name:       d.Get("name").(string),
		Properties: properties,
	}, nil
}

func parseAzureVaultBackupPolicyIdParts(id string) (infraName, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		infraName, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureVaultBackupPolicy(infraName string, d *schema.ResourceData, duplo *duplosdk.DuploAzureVaultBackupPolicy) error {
	d.Set("name", duplo.Name)
	d.Set("infra_name", infraName)
	d.Set("azure_id", duplo.ID)

	properties := duplo.Properties

	d.Set("timezone", properties.TimeZone)
	d.Set("instant_restore_retention_days", properties.InstantRpRetentionRangeInDays)
	schedule := properties.SchedulePolicy
	if schedule != nil && properties.PolicyType == "V2" {
		if err := d.Set("backup", flattenBackupProtectionPolicyVMScheduleV2(schedule)); err != nil {
			return fmt.Errorf("setting `backup`: %+v", err)
		}
	} else {
		if err := d.Set("backup", flattenBackupProtectionPolicyVMSchedule(schedule)); err != nil {
			return fmt.Errorf("setting `backup`: %+v", err)
		}
	}

	policyType := "V1"
	if properties.PolicyType != "" {
		policyType = string(properties.PolicyType)
	}
	d.Set("policy_type", policyType)

	retention := properties.RetentionPolicy
	if retention != nil && retention.RetentionPolicyType == "LongTermRetentionPolicy" {
		if s := retention.DailySchedule; s != nil {
			if err := d.Set("retention_daily", flattenBackupProtectionPolicyVMRetentionDaily(s)); err != nil {
				return fmt.Errorf("setting `retention_daily`: %+v", err)
			}
		} else {
			d.Set("retention_daily", nil)
		}

		if s := retention.WeeklySchedule; s != nil {
			if err := d.Set("retention_weekly", flattenBackupProtectionPolicyVMRetentionWeekly(s)); err != nil {
				return fmt.Errorf("setting `retention_weekly`: %+v", err)
			}
		} else {
			d.Set("retention_weekly", nil)
		}

		if s := retention.MonthlySchedule; s != nil {
			if err := d.Set("retention_monthly", flattenBackupProtectionPolicyVMRetentionMonthly(s)); err != nil {
				return fmt.Errorf("setting `retention_monthly`: %+v", err)
			}
		} else {
			d.Set("retention_monthly", nil)
		}

		if s := retention.YearlySchedule; s != nil {
			if err := d.Set("retention_yearly", flattenBackupProtectionPolicyVMRetentionYearly(s)); err != nil {
				return fmt.Errorf("setting `retention_yearly`: %+v", err)
			}
		} else {
			d.Set("retention_yearly", nil)
		}
	}
	return nil
}

func flattenBackupProtectionPolicyVMSchedule(schedule *duplosdk.DuploAzureVaultBackupSchedulePolicy) []interface{} {
	block := map[string]interface{}{}

	block["frequency"] = string(schedule.ScheduleRunFrequency)

	if times := schedule.ScheduleRunTimes; times != nil && len(*times) > 0 {
		block["time"] = (*times)[0].Format("15:04")
	}

	if days := schedule.ScheduleRunDays; days != nil && len(*days) > 0 {
		weekdays := make([]interface{}, 0)
		for _, d := range *days {
			weekdays = append(weekdays, d)
		}
		block["weekdays"] = schema.NewSet(schema.HashString, weekdays)
	}

	return []interface{}{block}
}

func flattenBackupProtectionPolicyVMScheduleV2(schedule *duplosdk.DuploAzureVaultBackupSchedulePolicy) []interface{} {
	block := map[string]interface{}{}

	frequency := schedule.ScheduleRunFrequency
	block["frequency"] = frequency

	switch frequency {
	case "Hourly":
		schedule := schedule.HourlySchedule
		if schedule.Interval != 0 {
			block["hour_interval"] = schedule.Interval
		}

		if schedule.ScheduleWindowDuration != 0 {
			block["hour_duration"] = schedule.ScheduleWindowDuration
		}

		if schedule.ScheduleWindowStartTime != nil {
			block["time"] = schedule.ScheduleWindowStartTime.Format("15:04")
		}
	case "Daily":
		schedule := schedule.DailySchedule
		if times := schedule.ScheduleRunTimes; times != nil && len(*times) > 0 {
			block["time"] = (*times)[0].Format("15:04")
		}
	case "Weekly":
		schedule := schedule.WeeklySchedule
		if days := schedule.ScheduleRunDays; days != nil && len(*days) > 0 {
			weekdays := make([]interface{}, 0)
			for _, d := range *days {
				weekdays = append(weekdays, duplosdk.DayOfWeekString(d))
			}
			block["weekdays"] = schema.NewSet(schema.HashString, weekdays)
		}

		if times := schedule.ScheduleRunTimes; times != nil && len(*times) > 0 {
			block["time"] = (*times)[0].Format("15:04")
		}
	default:
	}

	return []interface{}{block}
}

func flattenBackupProtectionPolicyVMRetentionDaily(daily *duplosdk.DuploAzureVaultBackupDailySchedule) []interface{} {
	block := map[string]interface{}{}

	if duration := daily.RetentionDuration; duration != nil {
		if v := duration.Count; v != 0 {
			block["count"] = v
		}
	}

	return []interface{}{block}
}

func flattenBackupProtectionPolicyVMRetentionMonthly(monthly *duplosdk.DuploAzureVaultBackupMonthlySchedule) []interface{} {
	block := map[string]interface{}{}

	if duration := monthly.RetentionDuration; duration != nil {
		if v := duration.Count; v != 0 {
			block["count"] = v
		}
	}

	if weekly := monthly.RetentionScheduleWeekly; weekly != nil {
		block["weekdays"], block["weeks"] = flattenBackupProtectionPolicyVMRetentionWeeklyFormat(weekly)
	}

	return []interface{}{block}
}

func flattenBackupProtectionPolicyVMRetentionWeeklyFormat(retention *duplosdk.DuploAzureVaultBackupRetentionScheduleWeekly) (weekdays, weeks *schema.Set) {
	if days := retention.DaysOfTheWeek; days != nil {
		slice := make([]interface{}, 0)
		for _, d := range *days {
			slice = append(slice, d)
		}
		weekdays = schema.NewSet(schema.HashString, slice)
	}

	if days := retention.WeeksOfTheMonth; days != nil {
		slice := make([]interface{}, 0)
		for _, d := range *days {
			slice = append(slice, d)
		}
		weeks = schema.NewSet(schema.HashString, slice)
	}

	return weekdays, weeks
}

func flattenBackupProtectionPolicyVMRetentionWeekly(weekly *duplosdk.DuploAzureVaultBackupWeeklySchedule) []interface{} {
	block := map[string]interface{}{}

	if duration := weekly.RetentionDuration; duration != nil {
		if v := duration.Count; v != 0 {
			block["count"] = v
		}
	}

	if days := weekly.DaysOfTheWeek; days != nil {
		weekdays := make([]interface{}, 0)
		for _, d := range *days {
			weekdays = append(weekdays, d)
		}
		block["weekdays"] = schema.NewSet(schema.HashString, weekdays)
	}

	return []interface{}{block}
}

func flattenBackupProtectionPolicyVMRetentionYearly(yearly *duplosdk.DuploAzureVaultBackupYearlySchedule) []interface{} {
	block := map[string]interface{}{}

	if duration := yearly.RetentionDuration; duration != nil {
		if v := duration.Count; v != 0 {
			block["count"] = v
		}
	}

	if weekly := yearly.RetentionScheduleWeekly; weekly != nil {
		block["weekdays"], block["weeks"] = flattenBackupProtectionPolicyVMRetentionWeeklyFormat(weekly)
	}

	if months := yearly.MonthsOfYear; months != nil {
		slice := make([]interface{}, 0)
		for _, d := range *months {
			slice = append(slice, d)
		}
		block["months"] = schema.NewSet(schema.HashString, slice)
	}

	return []interface{}{block}
}

func expandBackupProtectionPolicyVMSchedule(d *schema.ResourceData, times []time.Time) (*duplosdk.DuploAzureVaultBackupSchedulePolicy, error) {
	if bb, ok := d.Get("backup").([]interface{}); ok && len(bb) > 0 {
		block := bb[0].(map[string]interface{})

		policyType := d.Get("policy_type").(string)
		if policyType == "V1" {
			schedule := duplosdk.DuploAzureVaultBackupSchedulePolicy{ // LongTermSchedulePolicy has no properties
				ScheduleRunTimes:   &times,
				SchedulePolicyType: "SimpleSchedulePolicy",
			}

			if v, ok := block["frequency"].(string); ok {
				schedule.ScheduleRunFrequency = v
			}

			if v, ok := block["weekdays"].(*schema.Set); ok {
				days := make([]int, 0)
				for _, day := range v.List() {
					days = append(days, duplosdk.DayOfWeekIndex(day.(string)))
				}
				schedule.ScheduleRunDays = &days
			}

			return &schedule, nil
		} else {
			frequency := block["frequency"].(string)
			schedule := duplosdk.DuploAzureVaultBackupSchedulePolicy{
				SchedulePolicyType:   "SimpleSchedulePolicyV2",
				ScheduleRunFrequency: frequency,
			}

			switch frequency {
			case "Hourly":
				interval, ok := block["hour_interval"].(int)
				if !ok {
					return nil, fmt.Errorf("`hour_interval` must be specified when `backup.0.frequency` is `Hourly`")
				}

				duration, ok := block["hour_duration"].(int)
				if !ok {
					return nil, fmt.Errorf("`hour_duration` must be specified when `backup.0.frequency` is `Hourly`")
				}

				if duration%interval != 0 {
					return nil, fmt.Errorf("`hour_duration` must be multiplier of `hour_interval`")
				}

				schedule.HourlySchedule = &duplosdk.DuploAzureVaultBackupHourlySchedule{
					Interval:                interval,
					ScheduleWindowStartTime: &times[0],
					ScheduleWindowDuration:  duration,
				}
			case "Daily":
				schedule.DailySchedule = &duplosdk.DuploAzureVaultBackupDailySchedule{
					ScheduleRunTimes: &times,
				}
			case "Weekly":
				weekDays, ok := block["weekdays"].(*schema.Set)
				if !ok {
					return nil, fmt.Errorf("`weekdays` must be specified when `backup.0.frequency` is `Weekly`")
				}

				days := make([]int, 0)
				for _, day := range weekDays.List() {
					days = append(days, duplosdk.DayOfWeekIndex(day.(string)))
				}

				schedule.WeeklySchedule = &duplosdk.DuploAzureVaultBackupWeeklySchedule{
					ScheduleRunDays:  &days,
					ScheduleRunTimes: &times,
				}
			default:
				return nil, fmt.Errorf("Unrecognized value for backup.0.frequency")
			}

			return &schedule, nil
		}
	}

	return nil, nil
}

func expandBackupProtectionPolicyVMRetentionDaily(d *schema.ResourceData, times []time.Time) *duplosdk.DuploAzureVaultBackupDailySchedule {
	if rb, ok := d.Get("retention_daily").([]interface{}); ok && len(rb) > 0 {
		block := rb[0].(map[string]interface{})

		return &duplosdk.DuploAzureVaultBackupDailySchedule{
			RetentionTimes: &times,
			RetentionDuration: &duplosdk.DuploAzureVaultBackupRetentionDuration{
				Count:        block["count"].(int),
				DurationType: "Days",
			},
		}
	}

	return nil
}

func expandBackupProtectionPolicyVMRetentionWeekly(d *schema.ResourceData, times []time.Time) *duplosdk.DuploAzureVaultBackupWeeklySchedule {
	if rb, ok := d.Get("retention_weekly").([]interface{}); ok && len(rb) > 0 {
		block := rb[0].(map[string]interface{})

		retention := duplosdk.DuploAzureVaultBackupWeeklySchedule{
			RetentionTimes: &times,
			RetentionDuration: &duplosdk.DuploAzureVaultBackupRetentionDuration{
				Count:        block["count"].(int),
				DurationType: "Weeks",
			},
		}

		if v, ok := block["weekdays"].(*schema.Set); ok {
			days := make([]string, 0)
			for _, day := range v.List() {
				days = append(days, day.(string))
			}
			retention.DaysOfTheWeek = &days
		}

		return &retention
	}

	return nil
}

func expandBackupProtectionPolicyVMRetentionMonthly(d *schema.ResourceData, times []time.Time) *duplosdk.DuploAzureVaultBackupMonthlySchedule {
	if rb, ok := d.Get("retention_monthly").([]interface{}); ok && len(rb) > 0 {
		block := rb[0].(map[string]interface{})

		retention := duplosdk.DuploAzureVaultBackupMonthlySchedule{
			RetentionScheduleFormatType: "Weekly", // this is always weekly ¯\_(ツ)_/¯
			RetentionScheduleWeekly:     expandBackupProtectionPolicyVMRetentionWeeklyFormat(block),
			RetentionTimes:              &times,
			RetentionDuration: &duplosdk.DuploAzureVaultBackupRetentionDuration{
				Count:        block["count"].(int),
				DurationType: "Months",
			},
		}

		return &retention
	}

	return nil
}

func expandBackupProtectionPolicyVMRetentionYearly(d *schema.ResourceData, times []time.Time) *duplosdk.DuploAzureVaultBackupYearlySchedule {
	if rb, ok := d.Get("retention_yearly").([]interface{}); ok && len(rb) > 0 {
		block := rb[0].(map[string]interface{})

		retention := duplosdk.DuploAzureVaultBackupYearlySchedule{
			RetentionScheduleFormatType: "Weekly", // this is always weekly ¯\_(ツ)_/¯
			RetentionScheduleWeekly:     expandBackupProtectionPolicyVMRetentionWeeklyFormat(block),
			RetentionTimes:              &times,
			RetentionDuration: &duplosdk.DuploAzureVaultBackupRetentionDuration{
				Count:        block["count"].(int),
				DurationType: "Years",
			},
		}

		if v, ok := block["months"].(*schema.Set); ok {
			months := make([]string, 0)
			for _, month := range v.List() {
				months = append(months, month.(string))
			}
			retention.MonthsOfYear = &months
		}

		return &retention
	}

	return nil
}

func expandBackupProtectionPolicyVMRetentionWeeklyFormat(block map[string]interface{}) *duplosdk.DuploAzureVaultBackupRetentionScheduleWeekly {
	weekly := duplosdk.DuploAzureVaultBackupRetentionScheduleWeekly{}

	if v, ok := block["weekdays"].(*schema.Set); ok {
		days := make([]string, 0)
		for _, day := range v.List() {
			days = append(days, day.(string))
		}
		weekly.DaysOfTheWeek = &days
	}

	if v, ok := block["weeks"].(*schema.Set); ok {
		weeks := make([]string, 0)
		for _, week := range v.List() {
			weeks = append(weeks, week.(string))
		}
		weekly.WeeksOfTheMonth = &weeks
	}

	return &weekly
}
