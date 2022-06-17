# Cluster Autoscaler for Scaleway

The Scaleway Cloud Provider implementation scales nodes on different pools
attached to a Kapsule cluster. It can be configured from Scaleway Kapsule API.
The cluster pools need to have the option `Autoscaling` set to true to be managed by the autoscaler.

## Configuration

Cluster Autoscaler can be configured with 2 options
### Config file
a config file can be passed with the `--cloud-config` flag.  
here is the corresponding JSON schema:
* `cluster_id`: Kapsule Cluster Id
* `secret_key`: Secret Key used to manage associated Kapsule resources
* `region`: Region where the control-plane is runnning
* `api_url`: URL to contact Scaleway, defaults to `api.scaleway.com`

### Env variables

The values expected by the autoscaler are the same as above

- `CLUSTER_ID`
- `SCW_SECRET_KEY`
- `SCW_REGION`
- `SCW_API_URL`

## Notes

k8s nodes are identified through `node.Spec.ProviderId`, the scaleway node name or id MUST NOT be used.
