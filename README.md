# [WIP] cluster-autoscaler

This repo is intended to host the provider-agnostic core logic for Kubernetes Cluster Autoscaler, including the cluster snapshot, scale-up/down decision logic, and external/test provider interfaces.

Cluster Autoscaler originally started out in [the kubernetes/autoscaler repo](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler), with provider-agnostic core logic
combined with provider-specific logic. This original repo is intended to keep hosting the provider-specific logic and import the core CA logic from here to form a complete Cluster Autoscaler component.

---
**WARNING:** The migration of core Cluster Autoscaler logic from [kubernetes/autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) to this repo is still in progress, and can be tracked
in the following issues:
* Full migration - including refactoring the core Cluster Autoscaler logic into a proper library: https://github.com/kubernetes/autoscaler/issues/9264
* First steps - just a mechanical split of existing Cluster Autoscaler logic between the two repos: https://github.com/kubernetes/autoscaler/issues/9832

Until the first mechanical steps of the migration are completed, some files might be duplicated between the two repositories. This should be very temporary.

Until the full migration is completed, the core Cluster Autoscaler logic hosted here doesn't really form a proper library - it's more of a framework pattern, with
years of tech debt from the old repo. The full migration is expected to take considerable time.

---

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack channel](https://kubernetes.slack.com/messages/sig-autoscaling)
- [Mailing List](https://groups.google.com/a/kubernetes.io/g/sig-autoscaling)
- Weekly SIG Autoscaling meetings on Thursdays 17:00 CET. Joining the mailing list above should get you the invitation.

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
