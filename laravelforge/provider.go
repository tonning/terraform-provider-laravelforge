package laravelforge

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"tonning/terraform-provider-laravelforge/client"
)

func Provider() *schema.Provider {
	log.Printf("[INFO] [LARAVELFORGE] STARTING")
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LARAVELFORGE_TOKEN", nil),
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LARAVELFORGE_ORGANIZATION", nil),
				Description: "The slug of the Laravel Forge organization. Defaults to the only organization the token has access to.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"laravelforge_server":         resourceServer(),
			"laravelforge_site":           resourceSite(),
			"laravelforge_key":            resourceKey(),
			"laravelforge_sslcertificate": resourceSslCertificate(),
			"laravelforge_scheduledjob":   resourceScheduledJob(),
			"laravelforge_daemon":         resourceDaemon(),
			"laravelforge_redirectrule":   resourceRedirectRule(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"laravelforge_site":   dataSourceSite(),
			"laravelforge_server": dataSourceServer(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)
	organization := d.Get("organization").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := client.NewClient(nil, &token, &organization)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Laravel Forge client",
			Detail:   err.Error(),
		})
		return nil, diags
	}

	return c, diags
}
