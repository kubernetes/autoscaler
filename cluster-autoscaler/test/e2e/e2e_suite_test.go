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

package e2e

import (
	"fmt"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	// projectImage is the name of the image which will be build and loaded
	// with the code source changes to be tested.
	projectImage = "cluster-autoscaler:dev"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	_, _ = fmt.Fprintf(GinkgoWriter, "Starting cluster-autoscaler integration test suite\n")
	RunSpecs(t, "e2e suite")
}

var _ = BeforeSuite(func() {
	By("building the cluster-autoscaler image")
	cmd := exec.Command("make", "docker-build-e2e", "IMAGE=cluster-autoscaler", "TAG=dev")
	_, err := Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to build the cluster-autoscaler image")

	By("loading the cluster-autoscaler image on Kind")
	err = LoadImageToKindClusterWithName(projectImage)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to load the cluster-autoscaler image into Kind")
})
