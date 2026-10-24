# Release instructions

The following lists the steps needed to release the *addon-resizer*.

1. Verify that you have the
   [permissions needed to push new images to the repository](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/RELEASE.md#permissions).

1. Create a GitHub issue to track the release
   ([example](https://github.com/kubernetes/autoscaler/issues/7766)).

1. Bump the patch version in the
   [Makefile](https://github.com/kubernetes/autoscaler/blob/fa98f9b4c12a0cb25d6d7fd79497272137aa3458/addon-resizer/Makefile#L30)
   `TAG` ([example PR](https://github.com/kubernetes/autoscaler/pull/4079)).

1. Push images for all architectures to the
   [staging repository](https://gcr.io/k8s-staging-autoscaling).

   ```shell
   $ cd addon-resizer
   $ make all-push
   ```

1. Perform a [manual test](#manual-testing). Add test details to the GitHub 
   issue.

1. Promote images from the *staging* to the *prod* repository:

   1. Update the image mappings in the
      [`images.yaml` file](https://github.com/kubernetes/k8s.io/blob/main/registry.k8s.io/images/k8s-staging-autoscaling/images.yaml).

      To automatically update the `image.yaml` file, clone the
      https://github.com/kubernetes/k8s.io/ repository and run:

      ```shell
      $ cd registry.k8s.io/images/k8s-staging-autoscaling
      $ ./add-version.sh addon-resizer <version>
      ```

   1. Open a PR with those changes
      ([example](https://github.com/kubernetes/k8s.io/pull/2048)). After it's
      merged, the *promoter* will run automatically and upload the image from
      the *staging* to *prod* GCR repository. The post submit job status can be
      tracked on
      [testgrid](https://testgrid.k8s.io/sig-k8s-infra-k8sio#post-k8sio-image-promo).

1. Run the following to verify that the image is available:

   ```shell
   $ gcloud container images describe registry.k8s.io/autoscaling/addon-resizer:<version>
   $ # Alternatively: docker manifest inspect registry.k8s.io/autoscaling/addon-resizer:<version>
   ```

1. Close the GitHub issue associated with this release.

## Manual testing

1.  Create a cluster with 2 nodes.

    ```shell
    $ gcloud container clusters create --num-nodes=2 addon-resizer-release
    ```

1.  Deploy an
    [example deployment](https://github.com/kubernetes/autoscaler/tree/master/addon-resizer/deploy/example.yaml).
    Update the image path to point to the image you want to test.

1.  Note cpu and memory request of the deployment:

    ```shell
    $ kubectl get pods -o yaml | grep -A 6 -i resources:
    ```

1.  Resize cluster from 2 to 5 nodes:

    ```shell
    $ gcloud container clusters resize --num-nodes=5 -q addon-resizer-release
    ```

1.  Note cpu and memory request of the deployment. They should increase.

1.  Delete the test cluster.

    ```shell
    $ gcloud container clusters delete -q addon-resizer-release
    ```
