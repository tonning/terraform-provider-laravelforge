package laravelforge

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strconv"
	lf "tonning/terraform-provider-laravelforge/client"
)

func resourceKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeyCreate,
		ReadContext:   resourceKeyRead,
		DeleteContext: resourceKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"overwrite": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	var diags diag.Diagnostics

	log.Printf("[DEBUG] Key creation")
	opts := &lf.KeyCreateRequest{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("username"); ok {
		opts.Username = v.(string)
	}

	if v, ok := d.GetOk("public_key"); ok {
		opts.Key = v.(string)
	}

	if v, ok := d.GetOk("overwrite"); ok {
		opts.Overwrite = v.(bool)
	}

	log.Printf("[DEBUG] Key create configuration: %#v", opts)

	serverId := d.Get("server_id").(string)

	key, err := client.CreateKey(serverId, opts, true)

	if err != nil {
		return err
	}

	log.Printf("[INFO] [LARAVELFORGE] Key response: %#v", key)
	d.SetId(strconv.Itoa(key.Id))
	log.Printf("[INFO] [LARAVELFORGE] Key ID: %s", strconv.Itoa(key.Id))

	resourceKeyRead(ctx, d, m)

	return diags
}

func resourceKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[INFO] [LARAVELFORGE:resourceKeyRead] Start")
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Get("server_id").(string)
	keyId := d.Id()

	key, err := c.GetKey(serverId, keyId)
	log.Printf("[INFO] [LARAVELFORGE:resourceKeyRead] ID: %s Key: %#v", keyId, key)
	if err != nil {
		d.SetId("")
		return diags
	}

	d.SetId(strconv.Itoa(key.Id))

	d.Set("name", key.Name)
	d.Set("username", key.Username)
	d.Set("status", key.Status)

	log.Printf("[INFO] [LARAVELFORGE:resourceKeyRead] End")

	return diags
}

func resourceKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	keyId := d.Id()

	err := c.DeleteKey(d.Get("server_id").(string), keyId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
