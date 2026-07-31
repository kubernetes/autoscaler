# cluster-autoscaler

This repo hosts the provider-agnostic core logic for Kubernetes Cluster Autoscaler, including the cluster snapshot, scale-up/down decision logic, and external/test provider implementations.

Cluster Autoscaler originally started out in [the kubernetes/autoscaler repo](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler), with provider-agnostic core logic
combined with provider-specific logic. That original repo keeps hosting the provider-specific logic, and imports the core CA logic from here to form a complete Cluster Autoscaler component.

The core Cluster Autoscaler logic in this repo is essentially a framework - it determines the overall code flow, and
exposes code hooks that allow injecting custom logic in certain parts of the flow (in particular provider-specific logic).

The eventual goal is for the core CA logic in this repo to form a proper library, exposing only the "building blocks" - without dictating how the code flows
between them. This change is expected to take significant time, and can be tracked in https://github.com/kubernetes/autoscaler/issues/9264.

---
**WARNING: parts of this repo are still WIP** 

This repo is still in the last steps of being split away from [the kubernetes/autoscaler repo](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler).

Most of the files/directories have already been split and are in the correct repo, but some are temporarily duplicated between the 2 repos. If you want to modify any of the following paths, you must do so in the other repo - the files here are just mirrors:
* `pkg/apis`
* `pkg/cloudprovider/azure`
* `pkg/cloudprovider/gce`
* `pkg/proposals` and other files containing documentation

The structure of the files in the repo is still being worked on as well - in particular `pkg/` still contains a bunch of non-Go files. These will be moved to other places soon,
the intention is for `pkg/` to only contain Go code.

This state should be very temporary, the migration can be tracked in https://github.com/kubernetes/autoscaler/issues/9832.

The repo is otherwise fully functional, further changes to core Cluster Autoscaler logic should happen via PRs here.

---

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack channel](https://kubernetes.slack.com/messages/sig-autoscaling)
- [Mailing List](https://groups.google.com/a/kubernetes.io/g/sig-autoscaling)
- Weekly SIG Autoscaling meetings on Thursdays 17:00 CET. Joining the mailing list above should get you the invitation.

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
