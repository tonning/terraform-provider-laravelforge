---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "laravelforge_daemon Resource - terraform-provider-laravelforge"
subcategory: ""
description: |-
  
---

# laravelforge_daemon (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `command` (String)
- `server_id` (String)
- `user` (String)

### Optional

- `directory` (String)
- `processes` (Number)
- `start_secs` (Number) The total number of seconds the program must stay running in order to consider the start successful.
- `stop_signal` (String) The signal used to kill the program when a stop is requested.
- `stop_wait_secs` (Number) The number of seconds Supervisor will allow for the daemon to gracefully stop before forced termination.

### Read-Only

- `created_at` (String)
- `id` (String) The ID of this resource.
- `status` (String)


