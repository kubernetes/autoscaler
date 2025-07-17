/*
Copyright 2018 The Kubernetes Authors.

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

package flags

import (
	"flag"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/config"
	kubelet_config "k8s.io/kubernetes/pkg/kubelet/apis/config"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestParseSingleGpuLimit(t *testing.T) {
	type testcase struct {
		input                string
		expectError          bool
		expectedLimits       config.GpuLimits
		expectedErrorMessage string
	}

	testcases := []testcase{
		{
			input:       "gpu:1:10",
			expectError: false,
			expectedLimits: config.GpuLimits{
				GpuType: "gpu",
				Min:     1,
				Max:     10,
			},
		},
		{
			input:                "gpu:1",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit specification: gpu:1",
		},
		{
			input:                "gpu:1:10:x",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit specification: gpu:1:10:x",
		},
		{
			input:                "gpu:x:10",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit - min is not integer: gpu:x:10",
		},
		{
			input:                "gpu:1:y",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit - max is not integer: gpu:1:y",
		},
		{
			input:                "gpu:-1:10",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit - min is less than 0; gpu:-1:10",
		},
		{
			input:                "gpu:1:-10",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit - max is less than 0; gpu:1:-10",
		},
		{
			input:                "gpu:10:1",
			expectError:          true,
			expectedErrorMessage: "incorrect gpu limit - min is greater than max; gpu:10:1",
		},
	}

	for _, testcase := range testcases {
		limits, err := parseSingleGpuLimit(testcase.input)
		if testcase.expectError {
			assert.NotNil(t, err)
			if err != nil {
				assert.Equal(t, testcase.expectedErrorMessage, err.Error())
			}
		} else {
			assert.Equal(t, testcase.expectedLimits, limits)
		}
	}
}

func TestParseShutdownGracePeriodsAndPriorities(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []kubelet_config.ShutdownGracePeriodByPodPriority
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "Incorrect string - incorrect priority grace period pairs",
			input: "1:2,34",
			want:  nil,
		},
		{
			name:  "Incorrect string - trailing ,",
			input: "1:2, 3:4,",
			want:  nil,
		},
		{
			name:  "Incorrect string - trailing space",
			input: "1:2,3:4 ",
			want:  nil,
		},
		{
			name:  "Non integers - 1",
			input: "1:2,3:a",
			want:  nil,
		},
		{
			name:  "Non integers - 2",
			input: "1:2,3:23.2",
			want:  nil,
		},
		{
			name:  "parsable input",
			input: "1:2,3:4",
			want: []kubelet_config.ShutdownGracePeriodByPodPriority{
				{1, 2},
				{3, 4},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shutdownGracePeriodByPodPriority := parseShutdownGracePeriodsAndPriorities(tc.input)
			assert.Equal(t, tc.want, shutdownGracePeriodByPodPriority)
		})
	}
}

func TestCreateAutoscalingOptions(t *testing.T) {
	for _, tc := range []struct {
		testName            string
		flags               []string
		wantOptionsAsserter func(t *testing.T, gotOptions config.AutoscalingOptions)
	}{
		{
			testName: "DrainPriorityConfig defaults to an empty list when the flag isn't passed",
			flags:    []string{},
			wantOptionsAsserter: func(t *testing.T, gotOptions config.AutoscalingOptions) {
				if diff := cmp.Diff([]kubelet_config.ShutdownGracePeriodByPodPriority{}, gotOptions.DrainPriorityConfig, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("createAutoscalingOptions(): unexpected DrainPriorityConfig field (-want +got): %s", diff)
				}
			},
		},
		{
			testName: "DrainPriorityConfig is parsed correctly when the flag passed",
			flags:    []string{"--drain-priority-config", "5000:60,3000:50,0:40"},
			wantOptionsAsserter: func(t *testing.T, gotOptions config.AutoscalingOptions) {
				wantConfig := []kubelet_config.ShutdownGracePeriodByPodPriority{
					{Priority: 5000, ShutdownGracePeriodSeconds: 60},
					{Priority: 3000, ShutdownGracePeriodSeconds: 50},
					{Priority: 0, ShutdownGracePeriodSeconds: 40},
				}
				if diff := cmp.Diff(wantConfig, gotOptions.DrainPriorityConfig); diff != "" {
					t.Errorf("createAutoscalingOptions(): unexpected DrainPriorityConfig field (-want +got): %s", diff)
				}
			},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			pflag.CommandLine = pflag.NewFlagSet("test", pflag.ExitOnError)
			pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
			err := pflag.CommandLine.Parse(tc.flags)
			if err != nil {
				t.Errorf("pflag.CommandLine.Parse() got unexpected error: %v", err)
			}

			gotOptions := createAutoscalingOptions()
			tc.wantOptionsAsserter(t, gotOptions)
		})
	}
}
