package laravelforge

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lf "tonning/terraform-provider-laravelforge/client"
)

func capitalize(value string) string {
	r := []rune(value)
	if len(r) == 0 {
		return value
	}
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}

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
					return capitalize(val.(string))
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

	opts := &lf.CreateScheduledJob{
		Command:   d.Get("command").(string),
		Frequency: strings.ToLower(d.Get("frequency").(string)),
		User:      d.Get("user").(string),
	}

	if opts.Frequency == "custom" {
		field := func(name string) string {
			if v := d.Get(name).(string); v != "" {
				return v
			}

			return "*"
		}
		opts.Cron = strings.Join([]string{field("minute"), field("hour"), field("day"), field("month"), field("weekday")}, " ")
	}

	serverId := d.Get("server_id").(string)

	job, err := client.CreateScheduledJob(serverId, opts)
	if err != nil {
		return diag.FromErr(err)
	}
	if job.Id == "" {
		return diag.Errorf("Scheduled job not created")
	}

	d.SetId(job.Id)

	err = lf.WaitUntil(10*time.Minute, func() (bool, error) {
		current, err := client.GetScheduledJob(serverId, job.Id)
		if err != nil {
			return false, err
		}

		return current.Status != "installing", nil
	})
	if err != nil {
		return diag.Errorf("Unable to add scheduled job: %s", err)
	}

	return resourceScheduledJobRead(ctx, d, m)
}

func resourceScheduledJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	job, err := c.GetScheduledJob(d.Get("server_id").(string), d.Id())
	if err != nil {
		if lf.IsNotFound(err) {
			d.SetId("")

			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(job.Id)

	d.Set("command", job.Command)
	// Keep the notation produced by the StateFunc.
	d.Set("frequency", capitalize(job.Frequency))
	d.Set("user", job.User)
	d.Set("cron", job.Cron)
	d.Set("status", job.Status)
	d.Set("created_at", job.CreatedAt)

	return nil
}

func resourceScheduledJobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	if err := c.DeleteScheduledJob(d.Get("server_id").(string), d.Id()); err != nil && !lf.IsNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
