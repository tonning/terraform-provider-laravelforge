package laravelforge

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lf "tonning/terraform-provider-laravelforge/client"
)

func resourceRedirectRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedirectRuleCreate,
		ReadContext:   resourceRedirectRuleRead,
		DeleteContext: resourceRedirectRuleDelete,
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
			"from": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"to": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateDiagFunc: func(v any, p cty.Path) diag.Diagnostics {
					value := v.(string)
					expected := []string{
						"redirect",
						"permanent",
					}
					var diags diag.Diagnostics
					for _, acceptedValue := range expected {
						if acceptedValue == value {
							return diags
						}
					}
					diagnostic := diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Value not accepted",
						Detail:   fmt.Sprintf("%q is not in list of accepted values. Please use one of: %s", value, strings.Join(expected, ", ")),
					}
					diags = append(diags, diagnostic)

					return diags
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRedirectRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lf.Client)

	opts := &lf.CreateRedirectRuleRequest{
		From: d.Get("from").(string),
		To:   d.Get("to").(string),
		Type: d.Get("type").(string),
	}

	redirectRule, err := client.CreateRedirectRule(d.Get("server_id").(string), d.Get("site_id").(string), opts)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(redirectRule.Id)

	return resourceRedirectRuleRead(ctx, d, m)
}

func resourceRedirectRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	redirectRule, err := c.GetRedirectRule(d.Get("server_id").(string), d.Get("site_id").(string), d.Id())
	if err != nil {
		if lf.IsNotFound(err) {
			d.SetId("")

			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(redirectRule.Id)

	d.Set("from", redirectRule.From)
	d.Set("to", redirectRule.To)
	d.Set("type", redirectRule.Type)
	d.Set("created_at", redirectRule.CreatedAt)

	return nil
}

func resourceRedirectRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	err := c.DeleteRedirectRule(d.Get("server_id").(string), d.Get("site_id").(string), d.Id())
	if err != nil && !lf.IsNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
