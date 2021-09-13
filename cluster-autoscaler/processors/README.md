# Processor interfaces

Processors are golang interfaces that allow modifying key Cluster Autoscaler
data structures at various points in CA logic. This makes them a powerful tool
for building advanced features and customizing CA behavior.

  * Processors are intended to simplify management of customized CA fork by
    allowing most (and hopefully all) custom logic to be encapsulated behind a
    few interfaces.
  * Processors implementing generic features that may be useful to a large
    number of CA users may be merged to kubernetes/autoscaler repository.

Processors are the *preferred way* for introducing new features to core CA
logic. If a feature can be implemented via an existing processor interface, the
reviewers may ask for a PR to be implemented this way.

## Changes and deprecations

Adding new processors, adding new methods to existing processors or adding
additional parameters to existing methods may all happen in minor releases of
Cluster Autoscaler.

A processor interface (or parts of it) may be removed, for example as a result
of Cluster Autoscaler refactor or if it's superseded by a new processor.
Removing a processor interface or any of the methods or removing a method from
an existing processor will follow requirements stated in Rule #5b of
[Kubernetes deprecation
policy](https://kubernetes.io/docs/reference/using-api/deprecation-policy/),
meaning that the interface will function for 1 minor release or 6 months after
the deprecation is announced.
The announcement will be included in Cluster Autoscaler release notes and
discussed in sig-autoscaling weekly meeting.
