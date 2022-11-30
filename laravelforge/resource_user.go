package laravelforge

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"tonning/terraform-provider-laravelforge/client"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"provider": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ubuntu_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"php_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		Read: resourceUserRead,
	}
}

func resourceUserRead(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	userId := d.Id()
	user, err := apiClient.GetUser(userId)

	if err != nil {
		return err
	}

	if err := d.Set("user", user); err != nil {
		return err
	}

	return nil
}
