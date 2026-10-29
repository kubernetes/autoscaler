/*
Copyright The Kubernetes Authors.

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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const testImageTag = "test@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func scriptFixture(t *testing.T, name string) (string, string) {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash is required to test shell entrypoints")
	}
	root := t.TempDir()
	script := filepath.Join(root, "hack", "e2e", name)
	writeTestFile(t, script, readTestFile(t, name))
	if err := os.MkdirAll(filepath.Join(root, "e2e"), 0755); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(root, "tool bin")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("ARTIFACTS", filepath.Join(root, "artifacts"))
	t.Setenv("KUBECONFIG", filepath.Join(root, "workload-kubeconfig"))
	t.Setenv("SKIP", "")
	t.Setenv("TIMEOUT", "")
	return root, script
}

func TestE2ERunnerArguments(t *testing.T) {
	for _, extra := range [][]string{
		nil,
		{"--provider=local", "--non-blocking-taints=CriticalAddonsOnly", "--kubectl-path=/path with spaces/kubectl"},
	} {
		name := "existing invocation"
		if len(extra) > 0 {
			name = "test binary arguments"
		}
		t.Run(name, func(t *testing.T) {
			root, script := scriptFixture(t, "run-e2e-tests.sh")
			argsFile := filepath.Join(root, "ginkgo-args")
			t.Setenv("GINKGO_ARGS_FILE", argsFile)
			writeTestFile(t, filepath.Join(root, "tool bin", "go"), "#!/usr/bin/env bash\nexit 0\n")
			writeTestFile(t, filepath.Join(root, "e2e", "_output", "bin", "ginkgo"), `#!/usr/bin/env bash
set -euo pipefail
if [[ "$1" == build ]]; then
  exit 0
fi
printf '%s\0' "$@" > "$GINKGO_ARGS_FILE"
`)
			args := []string{script, "--report-prefix=standard_", "Standard Autoscaling"}
			if len(extra) > 0 {
				args = append(args, "--")
				args = append(args, extra...)
			}
			if output, err := exec.Command("bash", args...).CombinedOutput(); err != nil {
				t.Fatalf("runner failed: %v\n%s", err, output)
			}
			got := strings.Split(strings.TrimSuffix(readTestFile(t, argsFile), "\x00"), "\x00")
			want := []string{
				"-v", "--timeout=400m", "--focus=Standard Autoscaling", "./e2e.test", "--",
				"--report-dir=" + filepath.Join(root, "artifacts"),
				"--report-prefix=standard_", "--disable-log-dump",
			}
			want = append(want, extra...)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("ginkgo arguments:\ngot  %q\nwant %q", got, want)
			}
		})
	}
}

func azureFixture(t *testing.T) (string, string) {
	t.Helper()
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq is required to test the Azure shell entrypoint")
	}
	root, script := scriptFixture(t, "run-e2e-azure.sh")
	bin := filepath.Join(root, "tool bin")
	for key, value := range map[string]string{
		"KUBECTL":                      filepath.Join(bin, "kubectl"),
		"HELM":                         filepath.Join(bin, "helm"),
		"AZURE_SUBSCRIPTION_ID":        "test-subscription",
		"AZURE_TENANT_ID":              "test-tenant",
		"AZURE_RESOURCE_GROUP":         "MC_test-node-resource-group",
		"CLUSTER_AUTOSCALER_CLIENT_ID": "test-autoscaler-identity",
		"CLUSTER_AUTOSCALER_MIN_NODES": "3",
		"CLUSTER_AUTOSCALER_MAX_NODES": "6",
		"CALLS_FILE":                   filepath.Join(root, "calls"),
		"NODES_FILE":                   filepath.Join(root, "nodes.json"),
		"HELM_VALUES_FILE":             filepath.Join(root, "helm-values.json"),
		"HELM_VALUES_PATH_FILE":        filepath.Join(root, "helm-values-path"),
		"STANDARD_EXIT":                "0",
		"DRA_EXIT":                     "0",
		"LOG_EXIT":                     "0",
		"FINAL_LOG_EXIT":               "0",
	} {
		t.Setenv(key, value)
	}

	nodes := make([]map[string]map[string]string, 3)
	for i := range nodes {
		nodes[i] = map[string]map[string]string{"spec": {
			"providerID": fmt.Sprintf("azure:///subscriptions/test-subscription/resourceGroups/MC_test-node-resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/aks-linux-vmss/virtualMachines/%d", i),
		}}
	}
	data, err := json.Marshal(map[string][]map[string]map[string]string{"items": nodes})
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(root, "nodes.json"), string(data))
	writeTestFile(t, filepath.Join(bin, "kubectl"), `#!/usr/bin/env bash
set -euo pipefail
printf 'kubectl %s\n' "$*" >> "$CALLS_FILE"
if [[ "$1" == get && "$2" == nodes ]]; then
  cat "$NODES_FILE"
elif [[ "$*" == *" logs "* ]]; then
  echo "autoscaler diagnostics"
  if grep -q '^suite 90m ' "$CALLS_FILE"; then
    exit "$FINAL_LOG_EXIT"
  fi
  exit "$LOG_EXIT"
fi
`)
	writeTestFile(t, filepath.Join(bin, "helm"), `#!/usr/bin/env bash
set -euo pipefail
printf 'helm %s\n' "$*" >> "$CALLS_FILE"
while [[ $# -gt 0 ]]; do
  if [[ "$1" == --values ]]; then
    cp "$2" "$HELM_VALUES_FILE"
    printf '%s' "$2" > "$HELM_VALUES_PATH_FILE"
    shift
  fi
  shift
done
`)
	writeTestFile(t, filepath.Join(root, "hack", "e2e", "run-e2e-tests.sh"), `#!/usr/bin/env bash
set -euo pipefail
printf 'suite %s %s\n' "$TIMEOUT" "$*" >> "$CALLS_FILE"
if [[ "$2" == *"Standard Autoscaling" ]]; then
  exit "$STANDARD_EXIT"
elif [[ "$2" == *"DRA Autoscaling" ]]; then
  exit "$DRA_EXIT"
fi
echo "unexpected test focus" >&2
exit 99
`)
	return root, script
}

func TestAzureE2ERunner(t *testing.T) {
	helmPath, helmErr := exec.LookPath("helm")
	for _, tc := range []struct {
		name         string
		standardExit string
		draExit      string
		logExit      string
		finalLogExit string
		wantExit     int
		wantDRA      bool
	}{
		{name: "both suites", standardExit: "0", draExit: "0", logExit: "0", finalLogExit: "0", wantDRA: true},
		{name: "standard failure", standardExit: "17", draExit: "0", logExit: "0", finalLogExit: "0", wantExit: 17},
		{name: "DRA failure", standardExit: "0", draExit: "23", logExit: "0", finalLogExit: "0", wantExit: 23, wantDRA: true},
		{name: "failed diagnostics preserve test failure", standardExit: "17", draExit: "0", logExit: "31", finalLogExit: "0", wantExit: 17},
		{name: "failed diagnostics are reported", standardExit: "0", draExit: "0", logExit: "31", finalLogExit: "0", wantExit: 31},
		{name: "failed final diagnostics fail the job", standardExit: "0", draExit: "0", logExit: "0", finalLogExit: "31", wantExit: 1, wantDRA: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, script := azureFixture(t)
			t.Setenv("STANDARD_EXIT", tc.standardExit)
			t.Setenv("DRA_EXIT", tc.draExit)
			t.Setenv("LOG_EXIT", tc.logExit)
			t.Setenv("FINAL_LOG_EXIT", tc.finalLogExit)
			output, err := exec.Command("bash", script, "registry.example/cluster-autoscaler-azure-amd64", testImageTag).CombinedOutput()
			exitCode := 0
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				exitCode = exitErr.ExitCode()
			} else if err != nil {
				t.Fatalf("executing runner: %v\n%s", err, output)
			}
			if exitCode != tc.wantExit {
				t.Fatalf("exit code %d, want %d:\n%s", exitCode, tc.wantExit, output)
			}

			calls := readTestFile(t, filepath.Join(root, "calls"))
			if !strings.Contains(calls, "suite 300m --report-prefix=standard_") {
				t.Fatalf("standard suite was not invoked correctly:\n%s", calls)
			}
			if got := strings.Contains(calls, "suite 90m --report-prefix=dra_"); got != tc.wantDRA {
				t.Fatalf("DRA invoked=%v, want %v:\n%s", got, tc.wantDRA, calls)
			}
			if got := strings.Contains(calls, "kubectl apply -f"); got != tc.wantDRA {
				t.Fatalf("DRA driver installed=%v, want %v:\n%s", got, tc.wantDRA, calls)
			}
			if !strings.Contains(calls, "--provider=local --non-blocking-taints=CriticalAddonsOnly --kubectl-path=") {
				t.Fatalf("Azure test binary arguments are missing:\n%s", calls)
			}
			for _, phase := range []string{"standard", "dra"} {
				if phase == "dra" && !tc.wantDRA {
					continue
				}
				if _, err := os.Stat(filepath.Join(root, "artifacts", "cluster-autoscaler-"+phase+".log")); err != nil {
					t.Fatalf("missing %s diagnostics: %v", phase, err)
				}
			}
			valuesPath := readTestFile(t, filepath.Join(root, "helm-values-path"))
			if _, err := os.Stat(valuesPath); !os.IsNotExist(err) {
				t.Fatalf("temporary Helm values were not removed: %v", err)
			}
			checkAzureHelmValues(t, filepath.Join(root, "helm-values.json"))
			if tc.name == "both suites" {
				t.Run("render chart", func(t *testing.T) {
					if helmErr != nil {
						t.Skip("helm is required to validate the rendered deployment")
					}
					chart, err := filepath.Abs("../../charts/cluster-autoscaler")
					if err != nil {
						t.Fatal(err)
					}
					rendered, err := exec.Command(helmPath, "template", "cluster-autoscaler", chart,
						"--namespace=kube-system", "--kube-version=1.37.1", "--kubeconfig=/dev/null",
						"--values="+filepath.Join(root, "helm-values.json")).CombinedOutput()
					if err != nil {
						t.Fatalf("rendering Azure deployment: %v\n%s", err, rendered)
					}
					for _, want := range []string{
						"registry.example/cluster-autoscaler-azure-amd64:" + testImageTag,
						"--cloud-provider=azure",
						"--nodes=3:6:aks-linux-vmss",
						"--bypassed-scheduler-names=non-existing-bypassed-scheduler",
						"name: ARM_USE_WORKLOAD_IDENTITY_EXTENSION",
					} {
						if !strings.Contains(string(rendered), want) {
							t.Errorf("rendered deployment is missing %q", want)
						}
					}
				})
			}
		})
	}
}

func checkAzureHelmValues(t *testing.T, path string) {
	t.Helper()
	var values struct {
		FullnameOverride                  string
		CloudProvider                     string
		AzureResourceGroup                string
		AzureUseWorkloadIdentityExtension bool
		Image                             struct{ Repository, Tag string }
		NodeSelector                      map[string]string
		AutoscalingGroups                 []struct {
			Name             string
			MinSize, MaxSize int
		}
		RBAC struct {
			ServiceAccount struct {
				Name        string
				Annotations map[string]string
			}
		}
	}
	if err := json.Unmarshal([]byte(readTestFile(t, path)), &values); err != nil {
		t.Fatal(err)
	}
	if values.CloudProvider != "azure" || values.FullnameOverride != "cluster-autoscaler" ||
		!values.AzureUseWorkloadIdentityExtension || values.AzureResourceGroup != "MC_test-node-resource-group" {
		t.Fatalf("incorrect Azure deployment values: %+v", values)
	}
	if values.Image.Repository != "registry.example/cluster-autoscaler-azure-amd64" || values.Image.Tag != testImageTag {
		t.Fatalf("image identity not preserved: %+v", values.Image)
	}
	if values.NodeSelector["kubernetes.azure.com/mode"] != "system" || values.NodeSelector["kubernetes.io/os"] != "linux" {
		t.Fatalf("autoscaler is not isolated on the Linux system pool: %v", values.NodeSelector)
	}
	if len(values.AutoscalingGroups) != 1 || values.AutoscalingGroups[0].Name != "aks-linux-vmss" ||
		values.AutoscalingGroups[0].MinSize != 3 || values.AutoscalingGroups[0].MaxSize != 6 {
		t.Fatalf("incorrect VMSS bounds: %+v", values.AutoscalingGroups)
	}
	if values.RBAC.ServiceAccount.Name != "cluster-autoscaler" ||
		values.RBAC.ServiceAccount.Annotations["azure.workload.identity/client-id"] != "test-autoscaler-identity" ||
		values.RBAC.ServiceAccount.Annotations["azure.workload.identity/tenant-id"] != "test-tenant" {
		t.Fatalf("incorrect workload identity configuration: %+v", values.RBAC)
	}
}

func TestAzureE2ERunnerRejectsInvalidConfiguration(t *testing.T) {
	for _, tc := range []struct{ name, variable, value string }{
		{"missing workload kubeconfig", "KUBECONFIG", ""},
		{"missing workload identity", "CLUSTER_AUTOSCALER_CLIENT_ID", ""},
		{"zero minimum", "CLUSTER_AUTOSCALER_MIN_NODES", "0"},
		{"noninteger minimum", "CLUSTER_AUTOSCALER_MIN_NODES", "three"},
		{"maximum below minimum", "CLUSTER_AUTOSCALER_MAX_NODES", "2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, script := azureFixture(t)
			t.Setenv(tc.variable, tc.value)
			if output, err := exec.Command("bash", script, "registry.example/autoscaler", "test").CombinedOutput(); err == nil {
				t.Fatalf("invalid configuration succeeded:\n%s", output)
			}
			if _, err := os.Stat(filepath.Join(root, "calls")); !os.IsNotExist(err) {
				t.Fatalf("invalid configuration reached cluster commands: %v", err)
			}
		})
	}
}

func TestAzureE2ERunnerRejectsInvalidTopology(t *testing.T) {
	for _, tc := range []struct{ name, before, after, message string }{
		{"multiple VMSS", "aks-linux-vmss/virtualMachines/2", "aks-other-vmss/virtualMachines/2", "expected exactly one Linux user VMSS"},
		{"invalid provider ID", "virtualMachines/2", "invalid-instance-path", "expected an Azure VMSS provider ID on every Linux user node"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, script := azureFixture(t)
			path := filepath.Join(root, "nodes.json")
			writeTestFile(t, path, strings.Replace(readTestFile(t, path), tc.before, tc.after, 1))
			output, err := exec.Command("bash", script, "registry.example/autoscaler", "test").CombinedOutput()
			if err == nil || !strings.Contains(string(output), tc.message) {
				t.Fatalf("expected a topology error, got %v:\n%s", err, output)
			}
			if strings.Contains(readTestFile(t, filepath.Join(root, "calls")), "helm ") {
				t.Fatal("invalid topology reached Helm deployment")
			}
		})
	}
}
