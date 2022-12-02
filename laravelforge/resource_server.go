package laravelforge

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strconv"
	lf "tonning/terraform-provider-laravelforge/client"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cloud_provider": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ubuntu_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"php_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_ready": {
				Type:     schema.TypeBool,
				Required: false,
				Computed: true,
			},
			"provision_command": {
				Type:     schema.TypeString,
				Required: false,
				Computed: true,
			},
			"sudo_password": {
				Type:     schema.TypeString,
				Required: false,
				Computed: true,
			},
		},
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,
	}
}

func resourceServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	var diags diag.Diagnostics

	log.Printf("[DEBUG] Server creation")

	opts := &lf.ServerCreateRequest{
		Name:             d.Get("name").(string),
		Provider:         d.Get("cloud_provider").(string),
		Type:             d.Get("type").(string),
		UbuntuVersion:    d.Get("ubuntu_version").(string),
		PhpVersion:       d.Get("php_version").(string),
		IpAddress:        d.Get("ip_address").(string),
		PrivateIpAddress: d.Get("private_ip_address").(string),
	}

	server, err := client.CreateServer(opts)

	if err != nil {
		d.SetId("")
		return diags
	}

	d.SetId(strconv.Itoa(server.Server.ID))
	d.Set("is_ready", server.Server.IsReady)
	d.Set("provision_command", server.ProvisionCommand)
	d.Set("sudo_password", server.SudoPassword)

	resourceServerRead(ctx, d, m)

	return diags
}

func resourceServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Id()

	server, err := client.GetServer(serverId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(server.ID))

	d.Set("name", server.Name)
	d.Set("type", server.Type)
	d.Set("php_version", server.PhpVersion)
	d.Set("ip_address", server.IpAddress)
	d.Set("private_ip_address", server.PrivateIpAddress)
	d.Set("is_ready", server.IsReady)

	log.Printf("[INFO] [LARAVELFORGE:resourceSiteRead] End")

	return diags
}
func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)
	serverId := d.Id()

	serverUpdates := lf.ServerUpdateRequest{
		Name:             d.Get("domain").(string),
		IpAddress:        d.Get("ip_address").(string),
		PrivateIpAddress: d.Get("private_ip_address").(string),
	}

	_, err := client.UpdateServer(serverId, serverUpdates)
	if err != nil {
		return err
	}

	return resourceSiteRead(ctx, d, m)
}

func resourceServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Id()

	err := c.DeleteServer(serverId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
