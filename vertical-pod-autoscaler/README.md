# Vertical Pod Autoscaler

## Contents

<!-- toc -->
- [Intro](#intro)
- [Getting Started](#getting-started)
- [Components and Architecture](#components-and-architecture)
- [Features and Known limitations](#features-and-known-limitations)
- [Development and testing](#development-and-testing)
- [Related links](#related-links)
<!-- /toc -->

## Intro

Vertical Pod Autoscaler (VPA) frees users from the necessity of setting
up-to-date resource requests for the containers in their pods. When
configured, it will set the requests automatically based on usage and thus
allow proper scheduling onto nodes so that appropriate resource overhead is
available for each pod. It will also maintain ratios between requests and
limits that were specified in initial containers configuration.

It can both down-scale pods that are over-requesting resources, and also
up-scale pods that are under-requesting resources based on their usage over
time.

Autoscaling is configured with a
[Custom Resource Definition object](https://kubernetes.io/docs/concepts/api-extension/custom-resources/)
called [VerticalPodAutoscaler](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go).
It allows to specify which pods should be vertically autoscaled as well as if/how the
resource recommendations are applied.

To enable vertical pod autoscaling on your cluster please follow the installation
procedure described below.

## Getting Started

See [Installation](./docs/installation.md) for a guide on installation, followed by a the [Quick start](./docs/quickstart.md) guide.

Also refer to the [FAQ](./docs/faq.md) for more.

## Components and Architecture

The Vertical Pod Autoscaler consists of three parts. The recommender, updater and admission-controller.
Read more about them on the [components](./docs/components.md) page.

## Features and Known limitations

You can also read about the [features](./docs/features.md) and [known limitations](./docs/known-limitations.md) of the VPA.

> [!IMPORTANT]
> At the moment, VPA is not compatible with workloads that define pod-level `resources` stanzas. Work has started to add pod-level resources support for VPA, for more details see [AEP-7571](https://github.com/kubernetes/autoscaler/pull/8586). Users of VPA may encounter the following issues if they enable VPA for a workload with a pod-level resources stanza defined:
> * When the admission-controller tries to set a container-level limit higher than the pod-level limit, the operation is prohibited, and the pod will not be created.
> * When the admission-controller tries to set the newly recommended container-level requests, their total value cannot exceed the pod-level request or limit. For example, if the pod-level memory request is 100Mi, but the newly recommended total of container-level requests is 250Mi, the pod will fail to be created.

## Development and testing

Have a look at the [Development and testing](./docs/development-and-testing.md) page for more information about developing on the VPA.

## Related links

- [Design proposal](https://github.com/kubernetes/design-proposals-archive/blob/main/autoscaling/vertical-pod-autoscaler.md)
- [API definition](pkg/apis/autoscaling.k8s.io/v1/types.go)
- [API reference](./docs/api.md)
