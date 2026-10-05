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

package routines

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	recommender_config "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/config"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
)

// configWithPrometheusFlagsSet returns a config with every Prometheus history
// provider flag set to a distinct non-zero value.
func configWithPrometheusFlagsSet() *recommender_config.RecommenderConfig {
	config := recommender_config.DefaultRecommenderConfig()
	config.PrometheusAddress = "http://prometheus.example:9090"
	config.PrometheusInsecure = true
	config.QueryTimeout = "7m"
	config.HistoryLength = "9d"
	config.HistoryResolution = "2h"
	config.PodLabelPrefix = "label_"
	config.PodLabelsMetricName = "kube_pod_labels"
	config.PodNamespaceLabel = "namespace"
	config.PodNameLabel = "pod"
	config.CtrNamespaceLabel = "container_namespace"
	config.CtrPodNameLabel = "container_pod"
	config.CtrNameLabel = "container"
	config.PrometheusJobName = "cadvisor-job"
	config.CommonFlags.VpaObjectNamespace = "vpa-namespace"
	config.HistoryCPUMetric = "custom_cpu_metric"
	config.HistoryMemoryMetric = "custom_memory_metric"
	config.PrometheusBearerToken = "bearer-token"
	config.Username = "user"
	config.Password = "pass"
	return config
}

func configWithInvalidQueryTimeout() *recommender_config.RecommenderConfig {
	config := recommender_config.DefaultRecommenderConfig()
	config.QueryTimeout = "not-a-duration"
	return config
}

func TestNewPrometheusHistoryProviderConfig(t *testing.T) {
	tests := map[string]struct {
		config *recommender_config.RecommenderConfig
		// expectErr asserts that building the config fails.
		expectErr bool
		// expected, when set, is compared for equality against the result.
		expected *history.PrometheusHistoryProviderConfig
		// checkAllFieldsSet asserts that no field of the result is left at its
		// zero value, guarding against a flag being added without being wired up.
		checkAllFieldsSet bool
	}{
		"maps every flag to the correct field": {
			config: configWithPrometheusFlagsSet(),
			expected: &history.PrometheusHistoryProviderConfig{
				Address:                "http://prometheus.example:9090",
				Insecure:               true,
				QueryTimeout:           7 * time.Minute,
				HistoryLength:          "9d",
				HistoryResolution:      "2h",
				PodLabelPrefix:         "label_",
				PodLabelsMetricName:    "kube_pod_labels",
				PodNamespaceLabel:      "namespace",
				PodNameLabel:           "pod",
				CtrNamespaceLabel:      "container_namespace",
				CtrPodNameLabel:        "container_pod",
				CtrNameLabel:           "container",
				CadvisorMetricsJobName: "cadvisor-job",
				Namespace:              "vpa-namespace",
				CPUMetricName:          "custom_cpu_metric",
				MemoryMetricName:       "custom_memory_metric",
				Authentication: history.PrometheusCredentials{
					BearerToken: "bearer-token",
					Username:    "user",
					Password:    "pass",
				},
			},
		},
		"populates every field of the provider config": {
			config:            configWithPrometheusFlagsSet(),
			checkAllFieldsSet: true,
		},
		"returns an error for an invalid query timeout": {
			config:    configWithInvalidQueryTimeout(),
			expectErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			histConfig, err := newPrometheusHistoryProviderConfig(tc.config)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tc.expected != nil {
				assert.Equal(t, *tc.expected, histConfig)
			}
			if tc.checkAllFieldsSet {
				for _, field := range zeroValuedFields(reflect.ValueOf(histConfig), "PrometheusHistoryProviderConfig") {
					t.Errorf("%s was not populated; did a flag get added without wiring it up in newPrometheusHistoryProviderConfig?", field)
				}
			}
		})
	}
}

// zeroValuedFields returns the dotted paths of any zero-valued fields, recursing
// into nested structs.
func zeroValuedFields(v reflect.Value, path string) []string {
	var zero []string
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		name := path + "." + v.Type().Field(i).Name
		if field.Kind() == reflect.Struct {
			zero = append(zero, zeroValuedFields(field, name)...)
			continue
		}
		if field.IsZero() {
			zero = append(zero, name)
		}
	}
	return zero
}
