package laravelforge

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strconv"
	"time"
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

	var diags diag.Diagnostics

	log.Printf("[DEBUG] Daemon creation")
	opts := &lf.CreateDaemonRequest{
		Command:      d.Get("command").(string),
		Directory:    d.Get("directory").(string),
		User:         d.Get("user").(string),
		Processes:    d.Get("processes").(int),
		Startsecs:    d.Get("start_secs").(int),
		Stopwaitsecs: d.Get("stop_wait_secs").(int),
		Stopsignal:   d.Get("stop_signal").(string),
	}

	log.Printf("[DEBUG] Daemon configuration: %#v", opts)

	serverId := d.Get("server_id").(string)

	daemon, err := client.CreateDaemon(serverId, opts)
	daemonId := daemon.Id

	attempts := 0

	// Wait for status to be other than "installing".
	for shouldCheck := true; shouldCheck; shouldCheck = daemon.Status == "installing" {
		daemon, err := client.GetDaemon(serverId, strconv.Itoa(daemonId))
		log.Printf("[INFO] [LARAVELFORGE] Daemon waiting: %#v", daemon)

		if err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}

		if daemon.Status == "installed" {
			break
		}

		if attempts > 10 {
			return diag.Errorf("Unable to add daemon. Timeout.")
		}

		time.Sleep(time.Second * 5)
		attempts++
	}

	if err != nil {
		return err
	}

	log.Printf("[INFO] [LARAVELFORGE] Daemon response: %#v", daemon)
	d.SetId(strconv.Itoa(daemon.Id))
	log.Printf("[INFO] [LARAVELFORGE] Daemon ID: %s", strconv.Itoa(daemon.Id))

	resourceDaemonRead(ctx, d, m)

	return diags
}

func resourceDaemonRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[INFO] [LARAVELFORGE:resourceDaemonRead] Start")
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Get("server_id").(string)
	daemonId := d.Id()

	daemon, err := c.GetDaemon(serverId, daemonId)
	log.Printf("[INFO] [LARAVELFORGE:resourceDaemonRead] ID: %s Daemon: %#v", daemonId, daemon)
	if err != nil {
		d.SetId("")
		return diags
	}

	d.SetId(strconv.Itoa(daemon.Id))

	d.Set("command", daemon.Command)
	d.Set("directory", daemon.Directory)
	d.Set("user", daemon.User)
	d.Set("processes", daemon.Processes)
	d.Set("start_secs", daemon.Startsecs)
	d.Set("stop_wait_secs", daemon.Stopwaitsecs)
	d.Set("stop_signal", daemon.Stopsignal)
	d.Set("status", daemon.Status)
	d.Set("created_at", daemon.CreatedAt)

	log.Printf("[INFO] [LARAVELFORGE:resourceDaemonRead] End")

	return diags
}

func resourceDaemonDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	daemonId := d.Id()

	err := c.DeleteDaemon(d.Get("server_id").(string), daemonId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
