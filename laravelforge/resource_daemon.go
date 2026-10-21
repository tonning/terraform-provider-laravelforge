package laravelforge

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lf "tonning/terraform-provider-laravelforge/client"
)

func resourceDaemon() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDaemonCreate,
		ReadContext:   resourceDaemonRead,
		DeleteContext: resourceDaemonDelete,
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
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"directory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"processes": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ForceNew: true,
			},
			"start_secs": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The total number of seconds the program must stay running in order to consider the start successful.",
				ForceNew:    true,
			},
			"stop_wait_secs": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				Description: "The number of seconds Supervisor will allow for the daemon to gracefully stop before forced termination.",
				ForceNew:    true,
			},
			"stop_signal": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "SIGTERM",
				Description: "The signal used to kill the program when a stop is requested.",
				ForceNew:    true,
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

func resourceDaemonCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	serverId := d.Get("server_id").(string)

	opts := &lf.CreateDaemonRequest{
		Name:         daemonName(d),
		Command:      d.Get("command").(string),
		Directory:    d.Get("directory").(string),
		User:         d.Get("user").(string),
		Processes:    d.Get("processes").(int),
		Startsecs:    d.Get("start_secs").(int),
		Stopwaitsecs: d.Get("stop_wait_secs").(int),
		Stopsignal:   d.Get("stop_signal").(string),
	}

	daemon, err := client.CreateDaemon(serverId, opts)
	if err != nil {
		return diag.FromErr(err)
	}
	if daemon.Id == "" {
		return diag.Errorf("Daemon not created")
	}

	d.SetId(daemon.Id)

	err = lf.WaitUntil(5*time.Minute, func() (bool, error) {
		current, err := client.GetDaemon(serverId, daemon.Id)
		if err != nil {
			return false, err
		}

		return current.Status != "installing", nil
	})
	if err != nil {
		return diag.Errorf("Unable to add daemon: %s", err)
	}

	return resourceDaemonRead(ctx, d, m)
}

// daemonName derives a stable name, as the Forge API requires one.
func daemonName(d *schema.ResourceData) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%s|%s|%s", d.Get("command"), d.Get("user"), d.Get("directory"))))

	return "terraform-" + hex.EncodeToString(sum[:])[:8]
}

func resourceDaemonRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	daemon, err := c.GetDaemon(d.Get("server_id").(string), d.Id())
	if err != nil {
		if lf.IsNotFound(err) {
			d.SetId("")

			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(daemon.Id)

	d.Set("command", daemon.Command)
	d.Set("user", daemon.User)
	d.Set("processes", daemon.Processes)
	d.Set("status", daemon.Status)
	d.Set("created_at", daemon.CreatedAt)

	// Forge may fill in a default directory when none was given.
	if d.Get("directory").(string) != "" {
		d.Set("directory", daemon.Directory)
	}

	return nil
}

func resourceDaemonDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	if err := c.DeleteDaemon(d.Get("server_id").(string), d.Id()); err != nil && !lf.IsNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
