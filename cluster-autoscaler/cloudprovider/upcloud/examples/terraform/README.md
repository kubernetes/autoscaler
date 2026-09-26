# Cluster Autoscaler for UpCloud

## Deploy autoscaler using Terraform

### Authentication environment variables
- `UPCLOUD_TOKEN` - UpCloud API token
- `UPCLOUD_USERNAME` - UpCloud's API username
- `UPCLOUD_PASSWORD` - UpCloud's API user's password

### Apply Terraform plan
Init Terraform if needed
```shell
$ terraform init
```

This example supports either `autoscaler_token` or the `autoscaler_username` and `autoscaler_password` input variables. If both are provided, `UPCLOUD_TOKEN` takes precedence in the autoscaler container.

For demonstration purposes, we can use the same account that we use with the Terraform provider:
```shell
$ TF_VAR_autoscaler_username=$UPCLOUD_USERNAME TF_VAR_autoscaler_password=$UPCLOUD_PASSWORD terraform apply
```
or with token auth:
```shell
$ TF_VAR_autoscaler_token=$UPCLOUD_TOKEN terraform apply
```