package laravelforge

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	opts := &lf.KeyCreateRequest{
		Name:      d.Get("name").(string),
		Username:  d.Get("username").(string),
		Key:       d.Get("public_key").(string),
		Overwrite: d.Get("overwrite").(bool),
	}

	key, err := client.CreateKey(d.Get("server_id").(string), opts)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(key.Id)

	return resourceKeyRead(ctx, d, m)
}

func resourceKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	key, err := c.GetKey(d.Get("server_id").(string), d.Id())
	if err != nil {
		if lf.IsNotFound(err) {
			d.SetId("")

			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(key.Id)

	d.Set("name", key.Name)
	d.Set("username", key.User)
	d.Set("status", key.Status)

	return nil
}

func resourceKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	if err := c.DeleteKey(d.Get("server_id").(string), d.Id()); err != nil && !lf.IsNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
