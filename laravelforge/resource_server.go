package laravelforge

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The server provider, i.e. `custom`, `ocean2` (DigitalOcean), `hetzner`, `aws`, `vultr`, `akamai` or `laravel`.",
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
			"size": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The size of the server. Not used for custom servers.",
			},
			"ubuntu_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"php_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"database_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
			"ssh_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "The SSH port of a custom server.",
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
		Type:             d.Get("type").(string),
		Region:           d.Get("region").(string),
		Size:             d.Get("size").(string),
		UbuntuVersion:    d.Get("ubuntu_version").(string),
		PhpVersion:       d.Get("php_version").(string),
		DatabaseType:     d.Get("database_type").(string),
		IpAddress:        d.Get("ip_address").(string),
		PrivateIpAddress: d.Get("private_ip_address").(string),
		SshPort:          d.Get("ssh_port").(int),
		VpcUuid:          d.Get("ocean2_vpc_uuid").(string),
	}

	if v := d.Get("credential_id").(string); v != "" {
		credentialId, err := strconv.Atoi(v)
		if err != nil {
			return diag.Errorf("credential_id must be a number: %s", err)
		}
		opts.CredentialId = credentialId
	}

	server, err := client.CreateServer(opts)
	if err != nil {
		return diag.Errorf("Error: %s", err)
	}

	if server == nil || server.Id == "" {
		return diag.Errorf("Server not created")
	}

	d.SetId(server.Id)
	d.Set("provision_command", server.ProvisionCommand)
	d.Set("sudo_password", server.SudoPassword)

	// A custom server only becomes ready once the provision command has been
	// run on it, which happens after the resource is created.
	if opts.Provider != "custom" {
		err = lf.WaitUntil(20*time.Minute, func() (bool, error) {
			s, err := client.GetServer(server.Id)
			if err != nil {
				return false, err
			}

			return s.IsReady, nil
		})
		if err != nil {
			return diag.Errorf("Unable to create server: %s", err)
		}
	}

	if d.Get("opcache").(bool) {
		if err := client.EnableOpcache(server.Id); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Unable to enable OPcache",
				Detail:   err.Error(),
			})
		}
	}

	return append(diags, resourceServerRead(ctx, d, m)...)
}

func resourceServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	var diags diag.Diagnostics

	server, err := client.GetServer(d.Id())
	if err != nil {
		if lf.IsNotFound(err) {
			d.SetId("")

			return diags
		}
		return diag.FromErr(err)
	}

	d.SetId(server.Id)

	d.Set("name", server.Name)
	d.Set("ip_address", server.IpAddress)
	d.Set("private_ip_address", server.PrivateIpAddress)
	d.Set("is_ready", server.IsReady)

	// Forge may report these in a different notation than the configuration.
	// Only fill them in when unknown, i.e. after an import.
	if d.Get("type").(string) == "" {
		d.Set("type", server.Type)
	}
	if d.Get("php_version").(string) == "" {
		d.Set("php_version", server.PhpVersion)
	}
	if d.Get("cloud_provider").(string) == "" {
		d.Set("cloud_provider", server.Provider)
	}

	publicKey := server.LocalPublicKey
	if publicKey == "" {
		if key, err := client.GetServerPublicKey(server.Id); err == nil {
			publicKey = key
		}
	}
	d.Set("public_key", publicKey)

	return diags
}

func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)
	serverId := d.Id()

	serverUpdates := lf.ServerUpdateRequest{
		Name:             d.Get("name").(string),
		IpAddress:        d.Get("ip_address").(string),
		PrivateIpAddress: d.Get("private_ip_address").(string),
	}

	if err := client.UpdateServer(serverId, serverUpdates); err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("opcache") {
		var err error
		if d.Get("opcache").(bool) {
			err = client.EnableOpcache(serverId)
		} else {
			err = client.DisableOpcache(serverId)
		}
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceServerRead(ctx, d, m)
}

func resourceServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	if err := c.DeleteServer(d.Id()); err != nil && !lf.IsNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
