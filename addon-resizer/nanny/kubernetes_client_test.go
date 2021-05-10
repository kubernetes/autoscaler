package nanny

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestMergeResources(t *testing.T) {
	testCases := []struct {
		name    string
		current *corev1.ResourceRequirements
		new     *corev1.ResourceRequirements
		want    *corev1.ResourceRequirements
	}{
		{
			name:    "Overwrite empty",
			current: &corev1.ResourceRequirements{},
			new: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
		}, {
			name: "Overwrite all",
			current: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("200m")},
			},
			new: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
		}, {
			name: "Add limits",
			current: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
			},
			new: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
		}, {
			name: "Don't remove existing when new is empty",
			current: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("200m")},
			},
			new: &corev1.ResourceRequirements{},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("200m")},
			},
		}, {
			name: "Don't remove additional existing",
			current: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"cpu":    resource.MustParse("100m"),
					"memory": resource.MustParse("300Mi"),
				},
				Limits: corev1.ResourceList{
					"cpu":    resource.MustParse("200m"),
					"memory": resource.MustParse("900Mi"),
				},
			},
			new: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"cpu":    resource.MustParse("200m"),
					"memory": resource.MustParse("300Mi"),
				},
				Limits: corev1.ResourceList{
					"cpu":    resource.MustParse("600m"),
					"memory": resource.MustParse("900Mi"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeResources(tc.current, tc.new)
			verifyResources(t, "limits", got.Limits, tc.want.Limits)
			verifyResources(t, "requests", got.Requests, tc.want.Requests)
		})
	}
}
