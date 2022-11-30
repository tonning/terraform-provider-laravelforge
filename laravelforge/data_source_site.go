package laravelforge

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	lf "tonning/terraform-provider-laravelforge/client"
)

func dataSourceSite() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSiteRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"directory": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSiteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	serverID := d.Get("server_id").(string)
	siteID := strconv.Itoa(d.Get("id").(int))

	site, err := c.GetSite(serverID, siteID)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("Name", site.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("Username", site.Username); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("Directory", site.Directory); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("Status", site.Status); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(siteID)

	return diags
}
