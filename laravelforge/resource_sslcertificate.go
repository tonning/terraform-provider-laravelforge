package laravelforge

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strconv"
	"strings"
	"time"
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
				}, true),
			},
			"certificate_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"domains": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"token": {
				Type:         schema.TypeString,
				RequiredWith: []string{"dns_provider"},
				Optional:     true,
				Sensitive:    true,
			},
			"dns_provider": {
				Type:         schema.TypeString,
				RequiredWith: []string{"token"},
				ValidateFunc: validation.StringInSlice([]string{
					"digitalocean",
				}, true),
				Optional: true,
			},
			"keep_existing_on_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
	log.Printf("[DEBUG] SSL Certificate creation")
	client := m.(*lf.Client)

	var diags diag.Diagnostics
	var certificate *lf.Certificate
	var err diag.Diagnostics
	certificateType := d.Get("type").(string)
	serverId := d.Get("server_id").(string)
	siteId := d.Get("site_id").(string)

	if certificateType == "letsencrypt" {
		opts := &lf.SslCertificateCreateRequest{
			Domains: d.Get("domains").([]interface{}),
			DnsProvider: lf.DnsProvider{
				Type:              d.Get("dns_provider").(string),
				DigitaloceanToken: d.Get("token").(string),
			},
		}

		log.Printf("[DEBUG] SSL Certificate create configuration: %#v", opts)

		certificate, err = client.ObtainLetsEncryptSslCertificate(serverId, siteId, opts)
	} else if certificateType == "clone" {
		opts := &lf.SslCertificateCloneRequest{
			Type:          "clone",
			CertificateId: d.Get("certificate_id").(int),
		}

		certificate, err = client.CloneExistingSslCertificate(serverId, siteId, opts)
	}

	certificateId := certificate.Id
	attempts := 0

	// Wait for status to be other than "installing".
	for shouldCheck := true; shouldCheck; shouldCheck = certificate.Status == "installing" {
		certificate, err := client.GetCertificate(serverId, siteId, strconv.Itoa(certificateId))
		log.Printf("[INFO] [LARAVELFORGE] SSL Certificate waiting: %#v", certificate)

		if err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}

		if certificate.Status == "installed" {
			break
		}

		if attempts > 50 {
			return diag.Errorf("Unable to install SSL certificate. Timeout.")
		}

		time.Sleep(time.Second * 10)
		attempts++
	}

	if err != nil {
		return err
	}

	attempts = 0

	// Wait for status to be other than "installing".
	client.ActivateCertificate(serverId, siteId, strconv.Itoa(certificateId))

	for shouldCheck := true; shouldCheck; shouldCheck = certificate.Active == true {
		certificate, err := client.GetCertificate(serverId, siteId, strconv.Itoa(certificateId))
		log.Printf("[INFO] [LARAVELFORGE] SSL Activation waiting: %#v", certificate)

		if err != nil {
			return diag.Errorf("Unable to activate the certificate.")
		}

		if certificate.Active == true {
			break
		}

		if attempts > 50 {
			return diag.Errorf("Unable to activate SSL certificate. Timeout.")
		}

		time.Sleep(time.Second * 10)
		attempts++
	}

	log.Printf("[INFO] [LARAVELFORGE] SSL Certificate response: %#v", certificate)
	d.SetId(strconv.Itoa(certificate.Id))
	log.Printf("[INFO] [LARAVELFORGE] SSL Certificate ID: %s", strconv.Itoa(certificate.Id))

	resourceSslCertificateRead(ctx, d, m)

	return diags
}

func resourceSslCertificateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Start")
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Get("server_id").(string)
	siteId := d.Get("site_id").(string)
	certificateId := d.Id()
	isImporting := false

	if strings.Contains(certificateId, ".") {
		isImporting = true
		parts := strings.Split(certificateId, ".")
		serverId = parts[0]
		siteId = parts[1]
		certificateId = parts[2]
	}

	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] ID: %s Server ID: %s, Site ID: %s", certificateId, serverId, siteId)

	certificate, err := c.GetCertificate(serverId, siteId, certificateId)
	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] ID: %s Certificate: %#v", certificateId, certificate)
	if err != nil {
		d.SetId("")
		return diags
	}

	////domainsDoNotMatch := certificate.Domain != d.Get("domains").(string)
	////log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Domains do not match: %b", domainsDoNotMatch)
	////log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Domains: %s", certificate.Domain)
	//domains := ""
	//for index, domain := range d.Get("domains").([]interface{}) {
	//	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Domain: %s", domain)
	//	if index == 0 {
	//		domains = domain.(string)
	//	} else {
	//		domains = domains + "," + domain.(string)
	//	}
	//}
	//log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Domains: %s", domains)
	//log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Given Domains: %s, Read Domains: %s", domains, certificate.Domain)
	//if domains == certificate.Domain {
	//	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Domains Equal")
	//} else {
	//	d.SetId("")
	//	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Domains Not Equal")
	//
	//	return resourceSslCertificateCreate(ctx, d, m)
	//}

	d.SetId(strconv.Itoa(certificate.Id))

	domains := strings.Split(certificate.Domain, ",")
	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Read Domains: %#v", domains)
	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Read Domains: %s", domains)
	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Current Domains: %s", d.Get("domains"))
	d.Set("domains", domains)
	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] Domains after set: %s", d.Get("domains"))
	d.Set("request_status", certificate.RequestStatus)
	d.Set("created_at", certificate.CreatedAt)
	d.Set("existing", certificate.Existing)
	d.Set("active", certificate.Active)

	if isImporting == true {
		d.Set("server_id", serverId)
		d.Set("site_id", siteId)
	}

	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateRead] End")

	return diags
}

func resourceSslCertificateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//d.SetId("")
	//client := m.(*lf.Client)
	//siteID := d.Id()
	//serverID := d.Get("server_id").(string)
	log.Printf("[INFO] [LARAVELFORGE:resourceSslCertificateUpdate] ID: %s", d.Get("id"))
	//
	//if d.HasChanges("domain", "directory", "aliases", "wildcards") {
	//	siteUpdates := lf.SiteUpdateRequest{
	//		Name:      d.Get("domain").(string),
	//		Directory: d.Get("directory").(string),
	//		Aliases:   d.Get("aliases").([]interface{}),
	//		Wildcards: d.Get("wildcards").(bool),
	//	}
	//
	//	_, err := client.UpdateSite(serverID, siteID, siteUpdates)
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	//if d.HasChange("php_version") {
	//	versionUpdate := lf.SiteUpdatePhpVersion{
	//		Version: d.Get("php_version").(string),
	//	}
	//	_, err := client.UpdateSitePhpVersion(serverID, siteID, versionUpdate)
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	//return resourceSslCertificateRead(ctx, d, m)
	//return resourceSslCertificateCreate(ctx, d, m)
	return diag.Diagnostics{}
}

func resourceSslCertificateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	if d.Get("keep_existing_on_delete").(bool) == true {
		d.SetId("")

		return diags
	}

	c := m.(*lf.Client)

	serverId := d.Get("server_id").(string)
	siteId := d.Get("site_id").(string)
	certificateId := d.Id()

	err := c.DeleteCertificate(serverId, siteId, certificateId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
