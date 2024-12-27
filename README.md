## Developing
To build locally for development purposes, run `make dev`. Make sure the path to your Terraform project is correct in the "./GNUmakefile" `LOCALPATH`

### Debugging
Run `export TF_LOG=DEBUG`. Now when you run `terraform apply` it will be verbose.

## Releasing
Commit your changes, push up, and tag a new version with `v` prefix, i.e. `v1.2.3`. This will kick off the release process via a Github action. (https://github.com/tonning/terraform-provider-laravelforge/actions/workflows/release.yml).
This new version should automatically be picked up by the [Terraform registry](https://registry.terraform.io/providers/tonning/laravelforge/latest).
If not, you might have to [re-sync](https://registry.terraform.io/providers/tonning/laravelforge/1.0.2/settings/resync) permissions between Terraform and Github.

### Pulling in release after developing
* In your Terraform application delete the directory in `.terraform/` related to the version you just release, i.e. for version `v1.2.3` delete `.terraform/providers/registry.terraform.io/tonning/laravelforge/1.2.3`.
* Delete the `laravelforge` directory (not just the version) from `/Users/tonning/.terraform.d/plugins/registry.terraform.io/tonning/laravelforge`. Otherwise Terraform won't be able to pull from the registry.
* In your Terraform application run `terraform init -upgrade`.