# Cluster Autoscaler for CoreWeave

The Cluster Autoscaler for CoreWeave automatically adjusts the size of your Kubernetes cluster by adding or removing nodes in CoreWeave NodePools based on pending workloads and resource utilization.

## Configuration

### Cloud config

The CoreWeave provider does not require a separate cloud config file. All configuration is managed via Kubernetes CoreWeave NodePool resource and standard Cluster Autoscaler flags.

## Behavior

- The autoscaler monitors unschedulable pods and scales NodePools up or down as needed.
- Minimum and maximum NodePools sizes are configured via the CoreWeave NodePool custom resources.

## Development

To build and test the CoreWeave provider

1. **Build the `cluster-autoscaler` binary:**
    ```sh
    make build-in-docker
    ```

To build and test the CoreWeave provider in k8s:

1. **Build the `cluster-autoscaler` docker image:**
    ```sh
    REGISTRY=gcr.io/k8s-staging-autoscaling TAG=dev make make-image
    ```

2. **Push the Docker image to your registry:**
    ```sh
    REGISTRY=gcr.io/k8s-staging-autoscaling TAG=dev make push-image
    ```

## Usage

To enable the CoreWeave provider, set the following flag when running the autoscaler:

```sh
./cluster-autoscaler --cloud-provider=coreweave
```
## Usage with Helm Charts

When deploying the Cluster Autoscaler for CoreWeave using the provided Helm chart, you can customize its behavior using the `extraArgs` section in your `values.yaml` file.  
These arguments are passed directly to the Cluster Autoscaler container.

## Helm Chart Deployment

You can deploy the Cluster Autoscaler for CoreWeave using the official Helm chart.  
Below are the basic steps:

1. **Add the Helm repository (if not already added):**
    ```sh
    helm repo add autoscaler https://kubernetes.github.io/autoscaler
    helm repo update
    ```

2. **Customize your `values.yaml`:** (Replace image.repository and image.tag below with your registry and the tag you chose)
    - Set `cloudProvider: coreweave`
    - Set `autoDiscovery.clusterName: cluster.local`
    - Set `image.tag: dev`
    - Set `image.repository: gcr.io/k8s-staging-autoscaling/cluster-autoscaler-arm64`
    - Set any desired `extraArgs` as shown [parameters](../../FAQ.md#what-are-the-parameters-to-ca)
    - Optionally adjust resources, tolerations, nodeSelector, etc.

3. **Install or upgrade the chart:**
    ```sh
    helm upgrade --install cluster-autoscaler autoscaler/cluster-autoscaler \
      --namespace kube-system \
      -f values.yaml
    ```

4. **Verify deployment:**
    ```sh
    kubectl -n kube-system get pods -l app.kubernetes.io/name=coreweave-cluster-autoscaler
    ```

For more advanced configuration, refer to the [values.yaml](../../../charts/cluster-autoscaler/values.yaml) in this repository.


## Contributing

Contributions are welcome! Please open issues or pull requests for bug fixes, improvements, or new features.

## License

This project is licensed under the Apache 2.0 License.
