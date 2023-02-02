package laravelforge

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strconv"
	lf "tonning/terraform-provider-laravelforge/client"
)

func resourceSite() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSiteCreate,
		ReadContext:   resourceSiteRead,
		UpdateContext: resourceSiteUpdate,
		DeleteContext: resourceSiteDelete,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"directory": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"php_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aliases": {
				Type:        schema.TypeList,
				Description: "A list of domain aliases.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"wildcards": {
				Type:        schema.TypeBool,
				Description: "Whether to use wildcard sub-domains for the site.",
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceSiteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	var diags diag.Diagnostics

	log.Printf("[DEBUG] Site creation")
	opts := &lf.SiteCreateRequest{
		Domain: d.Get("domain").(string),
	}

	if v, ok := d.GetOk("username"); ok {
		opts.Username = v.(string)
	}

	if v, ok := d.GetOk("directory"); ok {
		opts.Directory = v.(string)
	}

	if v, ok := d.GetOk("project_type"); ok {
		opts.ProjectType = v.(string)
	}

	if v, ok := d.GetOk("php_version"); ok {
		opts.PhpVersion = v.(string)
	}

	log.Printf("[DEBUG] Site create configuration: %#v", opts)

	serverId := d.Get("server_id").(string)

	site, err := client.CreateSite(serverId, opts)

	if err != nil {
		return err
	}

	log.Printf("[INFO] [LARAVELFORGE] Site response: %#v", site)
	d.SetId(strconv.Itoa(site.ID))
	log.Printf("[INFO] [LARAVELFORGE] Site ID: %s", strconv.Itoa(site.ID))

	resourceSiteRead(ctx, d, m)

	return diags
}

func resourceSiteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[INFO] [LARAVELFORGE:resourceSiteRead] Start")
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Get("server_id").(string)
	siteId := d.Id()

	site, err := c.GetSite(serverId, siteId)
	log.Printf("[INFO] [LARAVELFORGE:resourceSiteRead] ID: %s Site: %#v", siteId, site)
	if err != nil {
		d.SetId("")
		return diags
	}

	d.SetId(strconv.Itoa(site.ID))

	d.Set("name", site.Name)
	d.Set("username", site.Username)
	d.Set("directory", site.Directory)
	d.Set("status", site.Status)
	d.Set("wildcards", site.Wildcards)

	log.Printf("[INFO] [LARAVELFORGE:resourceSiteRead] End")

	return diags
}

func resourceSiteUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)
	siteID := d.Id()
	serverID := d.Get("server_id").(string)

	if d.HasChanges("domain", "directory", "aliases", "wildcards") {
		siteUpdates := lf.SiteUpdateRequest{
			Name:      d.Get("domain").(string),
			Directory: d.Get("directory").(string),
			Aliases:   d.Get("aliases").([]interface{}),
			Wildcards: d.Get("wildcards").(bool),
		}

		_, err := client.UpdateSite(serverID, siteID, siteUpdates)
		if err != nil {
			return err
		}
	}

	if d.HasChange("php_version") {
		versionUpdate := lf.SiteUpdatePhpVersion{
			Version: d.Get("php_version").(string),
		}
		_, err := client.UpdateSitePhpVersion(serverID, siteID, versionUpdate)
		if err != nil {
			return err
		}
	}

	return resourceSiteRead(ctx, d, m)
}

func resourceSiteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	siteId := d.Id()

	err := c.DeleteSite(d.Get("server_id").(string), siteId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
