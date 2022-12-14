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
		},
		ResourcesMap: map[string]*schema.Resource{
			"laravelforge_server": resourceServer(),
			"laravelforge_site":   resourceSite(),
			"laravelforge_key":    resourceKey(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"laravelforge_site": dataSourceSite(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := client.NewClient(nil, &token)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Laravel Forge client",
			Detail:   "Unable to auth user for authenticated Laravel Forge client",
		})
		return nil, diags
	}

	return c, diags
}
