package laravelforge

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strconv"
	"strings"
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

	var diags diag.Diagnostics

	log.Printf("[DEBUG] Redirect Rule creation")
	opts := &lf.CreateRedirectRuleRequest{
		From: d.Get("from").(string),
		To:   d.Get("to").(string),
		Type: d.Get("type").(string),
	}

	log.Printf("[DEBUG] Redirect Rule configuration: %#v", opts)

	serverId := d.Get("server_id").(string)
	siteId := d.Get("site_id").(string)

	redirectRule, err := client.CreateRedirectRule(serverId, siteId, opts)
	//redirectRuleId := redirectRule.Id

	//attempts := 0
	//
	//// Wait for status to be other than "installing".
	//for shouldCheck := true; shouldCheck; shouldCheck = redirectRule.Status == "installing" {
	//	redirectRule, err = client.GetRedirectRule(serverId, strconv.Itoa(redirectRuleId))
	//	log.Printf("[INFO] [LARAVELFORGE] Redirect Rule waiting: %#v", redirectRule)
	//
	//	if err != nil {
	//		d.SetId("")
	//		return err
	//	}
	//
	//	if redirectRule.Status == "installed" {
	//		break
	//	}
	//
	//	if attempts > 60 {
	//		return diag.Errorf("Unable to add scheduled redirectRule. Timeout.")
	//	}
	//
	//	time.Sleep(time.Second * 10)
	//	attempts++
	//}

	if err != nil {
		return err
	}

	log.Printf("[INFO] [LARAVELFORGE] Redirect Rule response: %#v", redirectRule)
	d.SetId(strconv.Itoa(redirectRule.Id))
	log.Printf("[INFO] [LARAVELFORGE] Redirect Rule ID: %s", strconv.Itoa(redirectRule.Id))

	resourceRedirectRuleRead(ctx, d, m)

	return diags
}

func resourceRedirectRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[INFO] [LARAVELFORGE:resourceRedirectRuleRead] Start")
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Get("server_id").(string)
	siteId := d.Get("site_id").(string)
	redirectRuleId := d.Id()

	redirectRule, err := c.GetRedirectRule(serverId, siteId, redirectRuleId)
	log.Printf("[INFO] [LARAVELFORGE:resourceRedirectRuleRead] ID: %s Redirect Rule: %#v", redirectRuleId, redirectRule)
	if err != nil {
		d.SetId("")
		return diags
	}

	d.SetId(strconv.Itoa(redirectRule.Id))

	d.Set("from", redirectRule.From)
	d.Set("to", redirectRule.To)
	d.Set("type", redirectRule.Type)
	d.Set("created_at", redirectRule.CreatedAt)

	log.Printf("[INFO] [LARAVELFORGE:resourceRedirectRuleRead] End")

	return diags
}

func resourceRedirectRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*lf.Client)

	var diags diag.Diagnostics

	serverId := d.Get("server_id").(string)
	siteId := d.Get("site_id").(string)
	redirectRuleId := d.Id()

	err := c.DeleteRedirectRule(serverId, siteId, redirectRuleId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
