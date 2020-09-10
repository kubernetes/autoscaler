# Fork
This is a fork of [kubernetes/autoscaler](https://github.com/kubernetes/autoscaler) developed for the [gardener/machine-controller-manager](https://github.com/gardener/machine-controller-manager) and [gardener/gardener](https://github.com/gardener/gardener) project.

# Rationale behind the fork

[gardener](https://github.com/gardener/gardener) project uses [gardener/machine-controller-manager](https://github.com/gardener/machine-controller-manager) to manage machines for clusters deployed using gardener. The scaling of machines is then to be supported by an autoscaler that talks to machine-controller-manager to scale the machines attached to a cluster. For this we have forked the [kubernetes/autoscaler](https://github.com/kubernetes/autoscaler) and have implemented a new cloud provider interface to support the machine-controller-manager.

- [kubernetes/autoscaler](https://github.com/kubernetes/autoscaler) has a long term plan for supporting machine APIs (https://github.com/kubernetes/kube-deploy/blob/master/cluster-api/pkg/apis/cluster/v1alpha1/machine_types.go) to support scaling of machines in a cluster.

- [gardener/machine-controller-manager](https://github.com/gardener/machine-controller-manager) is meant to align with these machine APIs in future.

- Keeping the above in mind, the short term solution is to have a customized fork of autoscaler to support the machine-controller-manager in order to support the current requirements for Gardener and mitigate to the original cluster autoscaler as soon as both autoscaler and machine-controller-manager have aligned with machine APIs.
