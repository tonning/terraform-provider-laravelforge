package laravelforge

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strconv"
	"strings"
	"time"
	lf "tonning/terraform-provider-laravelforge/client"
	"unicode"
)

func resourceScheduledJob() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScheduledJobCreate,
		ReadContext:   resourceScheduledJobRead,
		DeleteContext: resourceScheduledJobDelete,
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"command": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"frequency": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateDiagFunc: func(v any, p cty.Path) diag.Diagnostics {
					value := v.(string)
					expected := []string{
						"minutely",
						"hourly",
						"nightly",
						"weekly",
						"monthly",
						"reboot",
						"custom",
					}
					var diags diag.Diagnostics
					for _, acceptedValue := range expected {
						if acceptedValue == value {
							return diags
						}
					}
					diagnostic := diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Value not accepted",
						Detail:   fmt.Sprintf("%q is not in list of accepted values. Please use one of: %s", value, strings.Join(expected, ", ")),
					}
					diags = append(diags, diagnostic)

					return diags
				},
				StateFunc: func(val any) string {
					r := []rune(val.(string))
					r[0] = unicode.ToUpper(r[0])

					return string(r)
				},
			},
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"minute": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"hour": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"day": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"month": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"weekday": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"cron": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceScheduledJobCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	var diags diag.Diagnostics

	log.Printf("[DEBUG] Scheduled Job creation")
	opts := &lf.CreateScheduledJob{
		Command:   d.Get("command").(string),
		Frequency: d.Get("frequency").(string),
		User:      d.Get("user").(string),
		Minute:    d.Get("minute").(string),
		Hour:      d.Get("hour").(string),
		Day:       d.Get("day").(string),
		Month:     d.Get("month").(string),
		Weekday:   d.Get("weekday").(string),
	}

	log.Printf("[DEBUG] Scheduled Job configuration: %#v", opts)

	serverId := d.Get("server_id").(string)

	job, err := client.CreateScheduledJob(serverId, opts)
	jobId := job.Id

	attempts := 0

	// Wait for status to be other than "installing".
	for shouldCheck := true; shouldCheck; shouldCheck = job.Status == "installing" {
		job, err := client.GetScheduledJob(serverId, strconv.Itoa(jobId))
		log.Printf("[INFO] [LARAVELFORGE] Scheduled Job waiting: %#v", job)

		if err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}

		if job.Status == "installed" {
			break
		}

		if attempts > 60 {
			return diag.Errorf("Unable to add scheduled job. Timeout.")
		}

		time.Sleep(time.Second * 10)
		attempts++
	}

	if err != nil {
		return err
	}

	log.Printf("[INFO] [LARAVELFORGE] Scheduled Job response: %#v", job)
	d.SetId(strconv.Itoa(job.Id))
	log.Printf("[INFO] [LARAVELFORGE] Scheduled Job ID: %s", strconv.Itoa(job.Id))

	resourceScheduledJobRead(ctx, d, m)

	return diags
}

func resourceScheduledJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[INFO] [LARAVELFORGE:resourceScheduledJobRead] Start")
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Get("server_id").(string)
	jobId := d.Id()

	job, err := c.GetScheduledJob(serverId, jobId)
	log.Printf("[INFO] [LARAVELFORGE:resourceScheduledJobRead] ID: %s Job: %#v", jobId, job)
	if err != nil {
		d.SetId("")
		return diags
	}

	d.SetId(strconv.Itoa(job.Id))

	d.Set("command", job.Command)
	d.Set("frequency", job.Frequency)
	d.Set("user", job.User)
	d.Set("cron", job.Cron)
	d.Set("status", job.Status)
	d.Set("created_at", job.CreatedAt)

	log.Printf("[INFO] [LARAVELFORGE:resourceScheduledJobRead] End")

	return diags
}

func resourceScheduledJobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	jobId := d.Id()

	err := c.DeleteScheduledJob(d.Get("server_id").(string), jobId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
