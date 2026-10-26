package laravelforge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The primary domain of the site. The Forge API does not support renaming it.",
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSiteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	serverId := d.Get("server_id").(string)

	opts := &lf.SiteCreateRequest{
		Domain:       d.Get("domain").(string),
		Username:     d.Get("username").(string),
		WebDirectory: d.Get("directory").(string),
		ProjectType:  d.Get("project_type").(string),
		PhpVersion:   d.Get("php_version").(string),
		Wildcards:    d.Get("wildcards").(bool),
	}

	site, err := client.CreateSite(serverId, opts)
	if err != nil {
		return diag.FromErr(err)
	}
	if site.Id == "" {
		return diag.Errorf("Site not created")
	}

	d.SetId(site.Id)

	err = lf.WaitUntil(15*time.Minute, func() (bool, error) {
		s, err := client.GetSite(site.Id)
		if err != nil {
			return false, err
		}
		if s.Status == "failed" {
			return false, fmt.Errorf("the site failed to install")
		}

		return s.Status != "creating" && s.Status != "installing", nil
	})
	if err != nil {
		return diag.Errorf("Unable to create site: %s", err)
	}

	if diags := syncDomains(client, d); diags.HasError() {
		return diags
	}

	return resourceSiteRead(ctx, d, m)
}

func resourceSiteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	site, err := c.GetSite(d.Id())
	if err != nil {
		if lf.IsNotFound(err) {
			d.SetId("")

			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(site.Id)

	d.Set("username", site.User)
	d.Set("status", site.Status)

	// Forge reports the absolute path, so treat a matching suffix as unchanged.
	configured := strings.Trim(d.Get("directory").(string), "/")
	if configured == "" || !strings.HasSuffix(strings.Trim(site.WebDirectory, "/"), configured) {
		d.Set("directory", site.WebDirectory)
	}

	if site.Wildcards != nil {
		d.Set("wildcards", *site.Wildcards)
	}

	return nil
}

func resourceSiteUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)
	siteID := d.Id()
	serverID := d.Get("server_id").(string)

	if d.HasChanges("directory", "php_version") {
		siteUpdates := lf.SiteUpdateRequest{
			Directory:  d.Get("directory").(string),
			PhpVersion: d.Get("php_version").(string),
		}

		if err := client.UpdateSite(serverID, siteID, siteUpdates); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("aliases", "wildcards") {
		if diags := syncDomains(client, d); diags.HasError() {
			return diags
		}
	}

	return resourceSiteRead(ctx, d, m)
}

func resourceSiteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	if err := c.DeleteSite(d.Get("server_id").(string), d.Id()); err != nil && !lf.IsNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

// syncDomains makes the aliases and the wildcard setting of the site's domains
// match the configuration.
func syncDomains(client *lf.Client, d *schema.ResourceData) diag.Diagnostics {
	serverId := d.Get("server_id").(string)
	siteId := d.Id()
	primary := d.Get("domain").(string)
	wildcards := d.Get("wildcards").(bool)

	desired := map[string]bool{}
	var aliases []string
	for _, alias := range d.Get("aliases").([]interface{}) {
		name, ok := alias.(string)
		if !ok || name == "" || name == primary || desired[name] {
			continue
		}
		desired[name] = true
		aliases = append(aliases, name)
	}

	existing, err := client.ListDomains(serverId, siteId)
	if err != nil {
		return diag.FromErr(err)
	}

	present := map[string]bool{}
	for _, domain := range existing {
		present[domain.Name] = true

		if domain.Type == "alias" && !desired[domain.Name] {
			if err := client.DeleteDomain(serverId, siteId, domain.Id); err != nil && !lf.IsNotFound(err) {
				return diag.FromErr(err)
			}
			continue
		}

		if domain.AllowWildcardSubdomains != wildcards {
			if err := client.UpdateDomain(serverId, siteId, domain.Id, domain.WwwRedirectType, wildcards); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	for _, alias := range aliases {
		if present[alias] {
			continue
		}
		if err := client.CreateDomain(serverId, siteId, alias, wildcards); err != nil {
			return diag.Errorf("Unable to add alias %s: %s", alias, err)
		}
	}

	return nil
}
