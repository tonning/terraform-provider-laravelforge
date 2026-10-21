package laravelforge

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lf "tonning/terraform-provider-laravelforge/client"
)

func dataSourceServer() *schema.Resource {
	computedString := func() *schema.Schema {
		return &schema.Schema{Type: schema.TypeString, Computed: true}
	}

	return &schema.Resource{
		ReadContext: dataSourceServerRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"credential_id":      computedString(),
			"name":               computedString(),
			"type":               computedString(),
			"cloud_provider":     computedString(),
			"cloud_provider_id":  computedString(),
			"region":             computedString(),
			"size":               computedString(),
			"ubuntu_version":     computedString(),
			"db_status":          computedString(),
			"redis_status":       computedString(),
			"php_version":        computedString(),
			"php_cli_version":    computedString(),
			"database_type":      computedString(),
			"ip_address":         computedString(),
			"private_ip_address": computedString(),
			"local_public_key":   computedString(),
			"ssh_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			// Kept for compatibility. Forge no longer reports these.
			"blackfire_status":  computedString(),
			"papertrail_status": computedString(),
			"revoked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": computedString(),
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

	server, err := c.GetServer(strconv.Itoa(d.Get("id").(int)))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(server.Id)
	d.Set("credential_id", string(server.CredentialId))
	d.Set("name", server.Name)
	d.Set("type", server.Type)
	d.Set("cloud_provider", server.Provider)
	d.Set("cloud_provider_id", server.Identifier)
	d.Set("region", server.Region)
	d.Set("size", server.Size)
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
	d.Set("revoked", server.Revoked)
	d.Set("created_at", server.CreatedAt)
	d.Set("is_ready", server.IsReady)

	return nil
}
