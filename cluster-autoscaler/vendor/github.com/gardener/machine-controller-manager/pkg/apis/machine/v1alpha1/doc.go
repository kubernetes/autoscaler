// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha1 is the v1alpha1 version of the API.
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=github.com/gardener/machine-controller-manager/pkg/apis/machine
// +k8s:openapi-gen=true
// +k8s:defaulter-gen=TypeMeta
// +groupName=machine.sapcloud.io
// +kubebuilder:object:generate=true
// Package v1alpha1 is a version of the API.
package v1alpha1
