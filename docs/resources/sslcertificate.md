---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "laravelforge_sslcertificate Resource - terraform-provider-laravelforge"
subcategory: ""
description: |-
  
---

# laravelforge_sslcertificate (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `server_id` (String)
- `site_id` (String)
- `type` (String)

### Optional

- `activate` (Boolean) Should activate the new SSL certificate finished installing.
- `certificate` (String)
- `certificate_id` (Number)
- `dns_provider` (String)
- `domains` (List of String)
- `keep_existing_on_delete` (Boolean)
- `key` (String)
- `token` (String, Sensitive)

### Read-Only

- `active` (Boolean)
- `created_at` (String)
- `existing` (Boolean)
- `id` (String) The ID of this resource.
- `request_status` (String)


