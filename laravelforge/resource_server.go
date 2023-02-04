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
			"credential_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
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
				Optional: true,
				Computed: true,
			},
			"private_ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"opcache": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ocean2_vpc_uuid": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"network": {
				Type:        schema.TypeList,
				Description: "An array of server IDs that the server should be able to connect to.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
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
				Type:      schema.TypeString,
				Required:  false,
				Computed:  true,
				Sensitive: false,
			},
			"public_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: false,
			},
		},
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	var diags diag.Diagnostics

	log.Printf("[DEBUG] Server creation")

	opts := &lf.ServerCreateRequest{
		Name:             d.Get("name").(string),
		Provider:         d.Get("cloud_provider").(string),
		CredentialId:     d.Get("credential_id").(string),
		Type:             d.Get("type").(string),
		Region:           d.Get("region").(string),
		UbuntuVersion:    d.Get("ubuntu_version").(string),
		PhpVersion:       d.Get("php_version").(string),
		IpAddress:        d.Get("ip_address").(string),
		PrivateIpAddress: d.Get("private_ip_address").(string),
	}

	server, err, _ := client.CreateServer(opts)
	if err != nil {
		return diag.Errorf("Error: %s", err)
	}

	log.Printf("[INFO] [LARAVELFORGE:resourceSiteCreate] Server: %#v", server)

	if server == nil {
		return diag.Errorf("Server not created")
	}

	serverId := server.Server.Id
	attempts := 0

	// Wait for status to be other than "installing".
	for shouldCheck := true; shouldCheck; shouldCheck = server.Server.IsReady {
		server, err, _ := client.GetServer(strconv.Itoa(serverId))
		log.Printf("[INFO] [LARAVELFORGE:resourceSiteCreate] Waiting - Attempts: %#v Server: %#v", attempts, server)

		if err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}

		if server.IsReady {
			break
		}

		if attempts > 10 {
			return diag.Errorf("Unable to create server. Too many attempts.")
		}

		time.Sleep(time.Second * 30)
		attempts++
	}

	if err != nil {
		d.SetId("")
		return diags
	}

	d.SetId(strconv.Itoa(server.Server.Id))
	d.Set("is_ready", server.Server.IsReady)
	d.Set("provision_command", server.ProvisionCommand)
	d.Set("sudo_password", server.SudoPassword)
	d.Set("public_key", server.Server.LocalPublicKey)

	if d.Get("opcache").(bool) == true {
		err := client.EnableOpcache(strconv.Itoa(serverId))
		if err != nil {
			return err
		}
	}

	resourceServerRead(ctx, d, m)

	return diags
}

func resourceServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Id()

	server, err, response := client.GetServer(serverId)
	if err != nil {
		if response.StatusCode == 404 {
			d.SetId("")

			return diags
		}
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(server.Id))

	d.Set("name", server.Name)
	d.Set("type", server.Type)
	d.Set("php_version", server.PhpVersion)
	d.Set("ip_address", server.IpAddress)
	d.Set("private_ip_address", server.PrivateIpAddress)
	d.Set("is_ready", server.IsReady)
	d.Set("public_key", server.LocalPublicKey)

	log.Printf("[INFO] [LARAVELFORGE:resourceSiteRead] End")

	return diags
}

func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[INFO] [LARAVELFORGE:resourceServerUpdate] Start")
	client := m.(*lf.Client)
	serverId := d.Id()
	log.Printf("[INFO] [LARAVELFORGE:resourceServerUpdate] Start 2")

	serverUpdates := lf.ServerUpdateRequest{
		Name:             d.Get("name").(string),
		IpAddress:        d.Get("ip_address").(string),
		PrivateIpAddress: d.Get("private_ip_address").(string),
	}

	log.Printf("[INFO] [LARAVELFORGE:resourceServerUpdate] server updates: %#v", serverUpdates)

	_, err, _ := client.UpdateServer(serverId, serverUpdates)
	if err != nil {
		return err
	}

	if d.Get("opcache").(bool) == true {
		err := client.EnableOpcache(serverId)
		if err != nil {
			return err
		}
	} else {
		err := client.DisableOpcache(serverId)
		if err != nil {
			return err
		}
	}

	return resourceServerRead(ctx, d, m)
}

func resourceServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Id()

	err, res := c.DeleteServer(serverId)
	if err != nil {
		if res.StatusCode != 404 {
			return diag.FromErr(err)
		}
	}

	d.SetId("")

	return diags
}
