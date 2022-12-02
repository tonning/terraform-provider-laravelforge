package laravelforge

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"tonning/terraform-provider-laravelforge/client"
)

func Provider() *schema.Provider {
	log.Printf("[INFO] [LARAVELFORGE] STARTING")
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LARAVELFORGE_TOKEN", "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhdWQiOiIxIiwianRpIjoiOWIxZDdjYjczOGFhYmM4ZGQzYTEzYWJmNTZiMzkwYjQ4MDBhNGE5NDM4MzRkN2I5YjVkYTZmMGEwZjg3MjdkNzY1NjdiZjc1NGRkZDVhN2UiLCJpYXQiOjE2Njk2MzExMjEuNTIyMDg5LCJuYmYiOjE2Njk2MzExMjEuNTIyMDkxLCJleHAiOjE5ODUyNTAzMjEuNTE0NTEzLCJzdWIiOiIzNTczMCIsInNjb3BlcyI6W119.f6SZ583ubtkYhb9AmvdoBRvdknAA7ckybhqY8slsHbaWU8edhvD_VWUFRaUc_R74HSEAGmPiOukqr6XS3tD_epY9fFz13wnMNe8OlH31kGspnAu6M69LKnmkTk7amGwunPfo25C111f1stl47LHZmfpAwIYpa_6YzBiacI00wPxOcqBDyV5pd55qcrWwpp5KFnBE_ZSwR9Yhq7oLJN9PuhPzOE0faPOqPrRBwuo6Ry66ymSwha3ldc5XmwVq4lJQVMFPwjwjybNZkenc_hd5feD3d1uM7viKEPXSolIetTy1PtjM2tZD78YmT_7ADGIaY3vBl8aki-zZ8YAdQ-nwY_sQ560LxcKvR2nj8fKo8-LmNazOjF-xRXX0niI_AkVC7Uh8z9cka0DtCLaISnZta7NmV-ythMVtO6WH7ECMLqh6sQeQVwF3Df9GKxGtXOXydLeppRPH7CCyeQTVy4RLxG0Td5DFW1HDLBfmiREw743CTfmTBgnWdD7PB5a7SbI6UYUGq66ZyfIH4zyp2GC4Rm3QnQ20z9faNhgB1SII8-lz2U5MQvi6HVGPCPjiG5PhmwmFidTuy0mx_T94D6-P64qH_hvWreb5EB0xSq5pkxAh0mYmhWNGYLm09-c2yRO82VssyHy1k39QLFEVtzXM-ixWOivNu1RSVpy3MJqhXbg"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"laravelforge_server": resourceServer(),
			"laravelforge_site":   resourceSite(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"laravelforge_site": dataSourceSite(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("api_token").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := client.NewClient(nil, &token)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Laravel Forge client",
			Detail:   "Unable to auth user for authenticated Laravel Forge client",
		})
		return nil, diags
	}

	return c, diags
}
