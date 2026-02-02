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
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	kwokProviderConfig = `
apiVersion: v1alpha1
readNodesFrom: configmap
nodegroups:
  fromNodeLabelKey: "kwok-nodegroup"
nodes:
  skipTaint: false
configmap:
  name: kwok-provider-templates
`

	kwokProviderTemplates = `
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Node
  metadata:
    annotations:
      node.alpha.kubernetes.io/ttl: "0"
      kwok.x-k8s.io/node: fake
    labels:
      beta.kubernetes.io/arch: amd64
      beta.kubernetes.io/os: linux
      kubernetes.io/arch: amd64
      kubernetes.io/hostname: kwok-worker
      kwok-nodegroup: kwok-worker
      kubernetes.io/os: linux
    name: kwok-worker
  status:
    allocatable:
      cpu: "4"
      memory: 32Gi
      pods: "110"
    capacity:
      cpu: "4"
      memory: 32Gi
      pods: "110"
    phase: Running
    conditions:
    - type: Ready
      status: "True"
      lastHeartbeatTime: "2023-05-31T04:40:17Z"
      lastTransitionTime: "2023-05-31T04:40:05Z"
      message: kubelet is posting ready status
      reason: KubeletReady
`

	caRbac = `
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    k8s-addon: cluster-autoscaler.addons.k8s.io
    app: cluster-autoscaler
  name: cluster-autoscaler
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-autoscaler
  labels:
    k8s-addon: cluster-autoscaler.addons.k8s.io
    app: cluster-autoscaler
rules:
  - apiGroups: [""]
    resources: ["events", "endpoints"]
    verbs: ["create", "patch"]
  - apiGroups: [""]
    resources: ["pods/eviction"]
    verbs: ["create"]
  - apiGroups: [""]
    resources: ["pods/status"]
    verbs: ["update"]
  - apiGroups: [""]
    resources: ["endpoints"]
    resourceNames: ["cluster-autoscaler"]
    verbs: ["get", "update"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["watch", "list", "get", "update", "create", "delete"] 
  - apiGroups: [""]
    resources:
      - "namespaces"
      - "pods"
      - "services"
      - "replicationcontrollers"
      - "persistentvolumeclaims"
      - "persistentvolumes"
    verbs: ["watch", "list", "get"]
  - apiGroups: ["extensions"]
    resources: ["replicasets", "daemonsets"]
    verbs: ["watch", "list", "get"]
  - apiGroups: ["policy"]
    resources: ["poddisruptionbudgets"]
    verbs: ["watch", "list"]
  - apiGroups: ["apps"]
    resources: ["statefulsets", "replicasets", "daemonsets"]
    verbs: ["watch", "list", "get"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses", "csinodes", "csidrivers", "csistoragecapacities", "volumeattachments"]
    verbs: ["watch", "list", "get"]
  - apiGroups: ["resource.k8s.io"]
    resources: ["resourceclaims", "resourceslices", "deviceclasses"]
    verbs: ["watch", "list", "get"]
  - apiGroups: ["batch", "extensions"]
    resources: ["jobs"]
    verbs: ["get", "list", "watch", "patch"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["create"]
  - apiGroups: ["coordination.k8s.io"]
    resourceNames: ["cluster-autoscaler"]
    resources: ["leases"]
    verbs: ["get", "update"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["create","list","watch"]
  - apiGroups: [""]
    resources: ["configmaps"]
    resourceNames: ["kwok-provider-config", "kwok-provider-templates", "cluster-autoscaler-status"]
    verbs: ["get", "update", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-autoscaler
  labels:
    k8s-addon: cluster-autoscaler.addons.k8s.io
    app: cluster-autoscaler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-autoscaler
subjects:
  - kind: ServiceAccount
    name: cluster-autoscaler
    namespace: default
`

	caDeployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-autoscaler
  namespace: default
  labels:
    app: cluster-autoscaler
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cluster-autoscaler
  template:
    metadata:
      labels:
        app: cluster-autoscaler
    spec:
      serviceAccountName: cluster-autoscaler
      containers:
        - image: cluster-autoscaler:dev
          name: cluster-autoscaler
          imagePullPolicy: IfNotPresent
          command:
            - ./cluster-autoscaler
            - --cloud-provider=kwok
            - --v=4
            - --stderrthreshold=info
            - --namespace=default
          env:
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: KWOK_PROVIDER_CONFIGMAP
              value: kwok-provider-config
`

	pendingPod = `
apiVersion: v1
kind: Pod
metadata:
  name: fake-pod
  labels:
    app: fake-pod
spec:
  containers:
  - name: fake-container
    image: fake-image
    resources:
      requests:
        cpu: "100m"
        memory: "100Mi"
  nodeSelector:
    kwok-nodegroup: kwok-worker
  tolerations:
  - key: "kwok-provider"
    operator: "Exists"
    effect: "NoSchedule"
`
)

var _ = Describe("Cluster Autoscaler with Kwok", Ordered, func() {
	BeforeAll(func() {
		By("Creating kwok-provider-config ConfigMap")
		err := createConfigMap("kwok-provider-config", "kwok.yaml", kwokProviderConfig)
		Expect(err).NotTo(HaveOccurred())

		By("Creating kwok-provider-templates ConfigMap")
		err = createTemplatesConfigMap("kwok-provider-templates", "templates.yaml", kwokProviderTemplates)
		Expect(err).NotTo(HaveOccurred())

		By("Applying RBAC")
		err = applyManifest(caRbac, "rbac.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("Deploying Cluster Autoscaler")
		err = applyManifest(caDeployment, "deployment.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for Cluster Autoscaler to be ready")
		Eventually(func() string {
			cmd := exec.Command("kubectl", "get", "pods", "-l", "app=cluster-autoscaler", "-o", "jsonpath={.items[0].status.phase}")
			out, _ := Run(cmd)
			return out
		}, 2*time.Minute, 5*time.Second).Should(Equal("Running"))
	})

	AfterEach(func() {
		specReport := CurrentSpecReport()
		if specReport.Failed() {
			By("Fetching controller logs")
			cmd := exec.Command("kubectl", "logs", "-l", "app=cluster-autoscaler", "--tail=200")
			out, err := Run(cmd)
			if err == nil {
				fmt.Fprintf(GinkgoWriter, "Controller logs:\n%s\n", out)
			} else {
				fmt.Fprintf(GinkgoWriter, "Failed to get logs: %v\n", err)
			}

			By("Fetching pending pod events")
			cmd = exec.Command("kubectl", "get", "events", "--field-selector", "involvedObject.name=fake-pod")
			out, err = Run(cmd)
			fmt.Fprintf(GinkgoWriter, "Pod events:\n%s\n", out)

			By("Fetching pending pod status")
			cmd = exec.Command("kubectl", "describe", "pod", "-l", "app=fake-pod")
			out, err = Run(cmd)
			fmt.Fprintf(GinkgoWriter, "Pod description:\n%s\n", out)
		}
	})

	AfterAll(func() {
		By("Deleting pending pod")
		cmd := exec.Command("kubectl", "delete", "pod", "pending-pod", "--ignore-not-found")
		_, _ = Run(cmd)

		By("Deleting Cluster Autoscaler deployment")
		cmd = exec.Command("kubectl", "delete", "deployment", "cluster-autoscaler", "--ignore-not-found")
		_, _ = Run(cmd)

		By("Deleting RBAC")
		cmd = exec.Command("kubectl", "delete", "-f", "rbac.yaml", "--ignore-not-found")
		_, _ = Run(cmd)

		By("Deleting ConfigMaps")
		cmd = exec.Command("kubectl", "delete", "configmap", "kwok-provider-config", "kwok-provider-templates", "--ignore-not-found")
		_, _ = Run(cmd)
	})

	It("should scale up when a pod is pending", func() {
		By("Creating a pending pod")
		err := applyManifest(pendingPod, "pod.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for TriggeredScaleUp event")
		Eventually(func() string {
			cmd := exec.Command("kubectl", "get", "events", "--field-selector", "involvedObject.name=fake-pod,reason=TriggeredScaleUp")
			out, _ := Run(cmd)
			return out
		}, 2*time.Minute, 5*time.Second).Should(ContainSubstring("TriggeredScaleUp"))

		By("Waiting for the pod to be scheduled")
		Eventually(func() string {
			cmd := exec.Command("kubectl", "get", "pod", "fake-pod", "-o", "jsonpath={.spec.nodeName}")
			out, _ := Run(cmd)
			return out
		}, 2*time.Minute, 5*time.Second).ShouldNot(BeEmpty())

		By("Verifying new node is created")
		Eventually(func() string {
			cmd := exec.Command("kubectl", "get", "nodes", "-l", "kwok-nodegroup=kwok-worker")
			out, _ := Run(cmd)
			return out
		}, 2*time.Minute, 5*time.Second).Should(ContainSubstring("kwok-worker-"))
	})
})

func createConfigMap(name string, fileName string, content string) error {
	err := os.WriteFile(fileName, []byte(content), 0644)
	if err != nil {
		return err
	}
	// defer os.Remove(fileName)

	cmd := exec.Command("kubectl", "create", "configmap", name, fmt.Sprintf("--from-file=config=%s", fileName))
	_, err = Run(cmd)
	return err
}

func createTemplatesConfigMap(name string, fileName string, content string) error {
	err := os.WriteFile(fileName, []byte(content), 0644)
	if err != nil {
		return err
	}
	cmd := exec.Command("kubectl", "create", "configmap", name, fmt.Sprintf("--from-file=templates=%s", fileName))
	_, err = Run(cmd)
	return err
}

func applyManifest(content string, fileName string) error {
	err := os.WriteFile(fileName, []byte(content), 0644)
	if err != nil {
		return err
	}
	cmd := exec.Command("kubectl", "apply", "-f", fileName)
	_, err = Run(cmd)
	return err
}
