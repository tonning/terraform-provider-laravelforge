package laravelforge

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	lf "tonning/terraform-provider-laravelforge/client"
)

func resourceSslCertificate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSslCertificateCreate,
		ReadContext:   resourceSslCertificateRead,
		UpdateContext: resourceSslCertificateUpdate,
		DeleteContext: resourceSslCertificateDelete,
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"site_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"letsencrypt",
					"clone",
					"existing",
				}, true),
			},
			"certificate_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"domains": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The domain to issue the certificate for. Forge issues a certificate per domain, so only the first entry is used. Defaults to the primary domain of the site.",
			},
			"verification_method": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "How a Let's Encrypt certificate is verified: `http-01` or `dns-01`. Defaults to `dns-01` for domains with wildcard sub-domains, otherwise `http-01`.",
				ValidateFunc: validation.StringInSlice([]string{
					"http-01",
					"dns-01",
				}, false),
			},
			"token": {
				Type:         schema.TypeString,
				RequiredWith: []string{"dns_provider"},
				Optional:     true,
				Sensitive:    true,
				Deprecated:   "Ignored. The current Laravel Forge API does not accept DNS provider credentials.",
			},
			"dns_provider": {
				Type:         schema.TypeString,
				RequiredWith: []string{"token"},
				ValidateFunc: validation.StringInSlice([]string{
					"digitalocean",
				}, true),
				Optional:   true,
				Deprecated: "Ignored. The current Laravel Forge API does not accept DNS provider credentials.",
			},
			"activate": {
				Description: "Should activate the new SSL certificate finished installing.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"keep_existing_on_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"existing": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"request_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSslCertificateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Get("server_id").(string)
	siteId := d.Get("site_id").(string)
	certificateType := strings.ToLower(d.Get("type").(string))
	activate := d.Get("activate").(bool)

	domains, err := client.ListDomains(serverId, siteId)
	if err != nil {
		return diag.FromErr(err)
	}

	var domain *lf.Domain
	requested := d.Get("domains").([]interface{})
	for i := range domains {
		if len(requested) > 0 {
			if name, _ := requested[0].(string); domains[i].Name == name {
				domain = &domains[i]
			}
		} else if domains[i].Type == "primary" {
			domain = &domains[i]
		}
	}
	if domain == nil {
		return diag.Errorf("Unable to find the domain to issue the certificate for on site %s", siteId)
	}
	if len(requested) > 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Only one domain per certificate",
			Detail:   fmt.Sprintf("Laravel Forge issues a certificate per domain. Only %s is used.", domain.Name),
		})
	}

	opts := &lf.CertificateCreateRequest{
		Type:   certificateType,
		Enable: activate,
	}

	switch certificateType {
	case "letsencrypt":
		opts.VerificationMethod = d.Get("verification_method").(string)
		if opts.VerificationMethod == "" {
			opts.VerificationMethod = "http-01"
			if domain.AllowWildcardSubdomains {
				opts.VerificationMethod = "dns-01"
			}
		}
	case "clone":
		opts.CloneCertificateId = d.Get("certificate_id").(int)
	case "existing":
		opts.ExistingCert = d.Get("certificate").(string)
		opts.ExistingKey = d.Get("key").(string)
	}

	certificate, err := client.CreateCertificate(serverId, siteId, domain.Id, opts)
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}
	if certificate.Id == "" {
		return append(diags, diag.Errorf("Certificate not created")...)
	}

	d.SetId(certificate.Id)
	d.Set("domain_id", domain.Id)

	activationRequested := false
	err = lf.WaitUntil(30*time.Minute, func() (bool, error) {
		current, err := client.GetCertificate(serverId, siteId, domain.Id, certificate.Id)
		if err != nil {
			return false, err
		}

		if strings.HasPrefix(current.Status, "failed") {
			return false, fmt.Errorf("certificate installation failed with status %q", current.Status)
		}
		if current.Status != "installed" {
			return false, nil
		}
		if !activate || current.Active {
			return true, nil
		}

		if !activationRequested {
			activationRequested = true
			if err := client.EnableCertificate(serverId, siteId, domain.Id, certificate.Id); err != nil {
				return false, err
			}
		}

		return false, nil
	})
	if err != nil {
		return append(diags, diag.Errorf("Unable to install SSL certificate: %s", err)...)
	}

	return append(diags, resourceSslCertificateRead(ctx, d, m)...)
}

func resourceSslCertificateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	serverId := d.Get("server_id").(string)
	siteId := d.Get("site_id").(string)
	id := d.Id()

	isImporting := false
	var importedType string
	var importedCertificateId int

	// Imports use the format server.site.id.type.certificate_id.
	if strings.Contains(id, ".") {
		parts := strings.Split(id, ".")
		if len(parts) != 5 {
			return diag.Errorf("Invalid import ID %q, expected server_id.site_id.id.type.certificate_id", id)
		}
		isImporting = true
		serverId, siteId, id, importedType = parts[0], parts[1], parts[2], parts[3]
		importedCertificateId, _ = strconv.Atoi(parts[4])
	}

	var certificate *lf.Certificate
	var err error
	if domainId := d.Get("domain_id").(string); domainId != "" && !isImporting {
		certificate, err = c.GetCertificate(serverId, siteId, domainId, id)
	} else {
		certificate, err = c.FindCertificate(serverId, siteId, id)
	}
	if err != nil {
		if lf.IsNotFound(err) {
			d.SetId("")

			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(certificate.Id)
	d.Set("domain_id", certificate.DomainId)
	d.Set("request_status", certificate.RequestStatus)
	d.Set("created_at", certificate.CreatedAt)
	d.Set("existing", certificate.Type == "existing")
	d.Set("active", certificate.Active)

	if len(d.Get("domains").([]interface{})) == 0 {
		if domains, err := c.ListDomains(serverId, siteId); err == nil {
			for _, domain := range domains {
				if domain.Id == certificate.DomainId {
					d.Set("domains", []string{domain.Name})
				}
			}
		}
	}

	if isImporting {
		d.Set("server_id", serverId)
		d.Set("site_id", siteId)
		d.Set("type", importedType)
		d.Set("certificate_id", importedCertificateId)
	}

	return nil
}

func resourceSslCertificateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceSslCertificateRead(ctx, d, m)
}

func resourceSslCertificateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.Get("keep_existing_on_delete").(bool) {
		d.SetId("")

		return nil
	}

	c := m.(*lf.Client)

	serverId := d.Get("server_id").(string)
	siteId := d.Get("site_id").(string)
	domainId := d.Get("domain_id").(string)

	if domainId == "" {
		certificate, err := c.FindCertificate(serverId, siteId, d.Id())
		if err != nil {
			if lf.IsNotFound(err) {
				d.SetId("")

				return nil
			}
			return diag.FromErr(err)
		}
		domainId = certificate.DomainId
	}

	if err := c.DeleteCertificate(serverId, siteId, domainId, d.Id()); err != nil && !lf.IsNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
