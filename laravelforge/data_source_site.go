package laravelforge

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lf "tonning/terraform-provider-laravelforge/client"
)

func dataSourceSite() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSiteRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"directory": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSiteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	siteID := strconv.Itoa(d.Get("id").(int))

	site, err := c.GetSite(siteID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(siteID)
	d.Set("name", site.Name)
	d.Set("username", site.User)
	d.Set("directory", site.WebDirectory)
	d.Set("status", site.Status)

	return nil
}
