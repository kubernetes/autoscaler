/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// models-schema is a tool to generate the OpenAPI schema JSON from the
// generated openapi definitions. This is used by applyconfiguration-gen
// to generate proper apply configurations with Extract functions.
//
// Adapted from:
// https://github.com/kubernetes/kubernetes/blob/master/pkg/generated/openapi/cmd/models-schema/main.go
// https://github.com/kubernetes-sigs/gateway-api/blob/main/pkg/generated/openapi/cmd/modelschema/main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/validation/spec"

	openapi "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/generated/openapi"
)

func main() {
	if err := output(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		os.Exit(1)
	}
}

func output() error {
	refFunc := func(name string) spec.Ref {
		return spec.MustCreateRef(fmt.Sprintf("#/definitions/%s", friendlyName(name)))
	}

	defs := openapi.GetOpenAPIDefinitions(refFunc)

	schemaDefs := make(map[string]spec.Schema, len(defs))
	for k, v := range defs {
		// Use v2 schema if available (preferred for structured merge-diff)
		if schema, ok := v.Schema.Extensions[common.ExtensionV2Schema]; ok {
			if v2Schema, isOpenAPISchema := schema.(spec.Schema); isOpenAPISchema {
				schemaDefs[friendlyName(k)] = v2Schema
				continue
			}
		}
		schemaDefs[friendlyName(k)] = v.Schema
	}

	data, err := json.Marshal(&spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Definitions: schemaDefs,
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:   "VPA",
					Version: "unversioned",
				},
			},
			Swagger: "2.0",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal openapi schema: %w", err)
	}

	os.Stdout.Write(data)
	return nil
}

// friendlyName converts a Go package path to a friendly OpenAPI definition name.
// e.g., "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1.VerticalPodAutoscaler"
// becomes "io.k8s.autoscaler.vertical-pod-autoscaler.pkg.apis.autoscaling.k8s.io.v1.VerticalPodAutoscaler"
func friendlyName(name string) string {
	nameParts := strings.Split(name, "/")
	// Reverse the first part (e.g., "k8s.io" -> "io.k8s")
	if len(nameParts) > 0 && strings.Contains(nameParts[0], ".") {
		parts := strings.Split(nameParts[0], ".")
		for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
			parts[i], parts[j] = parts[j], parts[i]
		}
		nameParts[0] = strings.Join(parts, ".")
	}
	return strings.Join(nameParts, ".")
}
