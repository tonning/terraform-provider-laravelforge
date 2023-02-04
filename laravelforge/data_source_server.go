package laravelforge

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
	"strconv"
	"time"
	lf "tonning/terraform-provider-laravelforge/client"
)

func dataSourceServer() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServerRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"credential_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloud_provider": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloud_provider_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ubuntu_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"redis_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"php_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"php_cli_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"private_ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"blackfire_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"papertrail_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revoked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_ready": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"network": {
				Type:        schema.TypeList,
				Description: "An array of server IDs that the server should be able to connect to.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := strconv.Itoa(d.Get("id").(int))

	server, err, response := c.GetServer(serverId)
	log.Printf("[INFO] [LARAVELFORGE:dataSourceServerRead] Server: %#v", server)

	if err != nil {
		return diag.FromErr(err)
	}

	if response.StatusCode == http.StatusTooManyRequests {
		time.Sleep(time.Second * 30)

		return dataSourceServerRead(ctx, d, m)
	}

	log.Printf("[INFO] [LARAVELFORGE:dataSourceServerRead] 2 Server: %#v", server)
	d.SetId(strconv.Itoa(server.Id))
	d.Set("credential_id", server.CredentialId)
	d.Set("name", server.Name)
	d.Set("type", server.Type)
	d.Set("cloud_provider", server.Provider)
	d.Set("cloud_provider_id", server.ProviderId)
	d.Set("region", server.Region)
	d.Set("ubuntu_version", server.UbuntuVersion)
	d.Set("db_status", server.DbStatus)
	d.Set("redis_status", server.RedisStatus)
	d.Set("php_version", server.PhpVersion)
	d.Set("php_cli_version", server.PhpCliVersion)
	d.Set("database_type", server.DatabaseType)
	d.Set("ip_address", server.IpAddress)
	d.Set("ssh_port", server.SshPort)
	d.Set("private_ip_address", server.PrivateIpAddress)
	d.Set("local_public_key", server.LocalPublicKey)
	d.Set("blackfire_status", server.BlackfireStatus)
	d.Set("papertrail_status", server.PapertrailStatus)
	d.Set("revoked", server.Revoked)
	d.Set("created_at", server.CreatedAt)
	d.Set("is_ready", server.IsReady)

	return diags
}
