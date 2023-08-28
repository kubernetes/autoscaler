package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/addon-resizer/nanny"
	"k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig"
	nannyconfigalpha "k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig/v1alpha1"
)

func TestConfigureAndRunNanny(t *testing.T) {
	nannyConfigurationFromFlags := &nannyconfigalpha.NannyConfiguration{
		BaseCPU:       "300m",
		CPUPerNode:    "350m",
		BaseMemory:    "200Mi",
		MemoryPerNode: "20Mi",
	}
	*estimator = "linear"
	*baseStorage = nannyconfig.NoValue
	n := nanny.Nanny{}

	f, err := os.Create("NannyConfiguration")
	defer os.Remove(f.Name())
	assert.NoError(t, err)
	_, err = f.WriteString(`
apiVersion: nannyconfig/v1alpha1
kind: NannyConfiguration
baseCPU: 264m	
`)
	assert.NoError(t, err)
	configFilePath, err := filepath.Abs(filepath.Dir(f.Name()))
	assert.NoError(t, err)

	expectedEstimator := nanny.LinearEstimator{
		Resources: []nanny.Resource{
			{
				Name:         "cpu",
				Base:         resource.MustParse("264m"),
				ExtraPerNode: resource.MustParse("350m"),
			},
			{
				Base:         resource.MustParse("200Mi"),
				ExtraPerNode: resource.MustParse("20Mi"),
				Name:         "memory",
			},
		},
	}
	nannyEstimator, err := getNannyEstimator(n, *nannyConfigurationFromFlags, configFilePath)
	assert.NoError(t, err)
	assert.Equal(t, expectedEstimator, nannyEstimator)
}
