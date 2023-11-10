/*
Copyright 2016 The Kubernetes Authors.

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

/*
Package nanny implements logic to poll the k8s apiserver for cluster status,
and update a deployment based on that status.
*/
package nanny

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	log "github.com/golang/glog"
	inf "gopkg.in/inf.v0"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/autoscaler/addon-resizer/healthcheck"
	"k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig"
	nannyscheme "k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig/scheme"
	nannyconfigalpha "k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig/v1alpha1"
)

type operation int

const (
	unknown               operation = iota
	scaleDown             operation = iota
	scaleUp               operation = iota
	ContainerProportional           = "container-proportional"
	NodeProportional                = "node-proportional"
)

type updateResult int

const (
	noChange  updateResult = iota
	postpone  updateResult = iota
	overwrite updateResult = iota
)

type Nanny struct {
	StopChan       chan struct{}
	Client         KubernetesClient
	PollPeriod     time.Duration
	ScaleDownDelay time.Duration
	ScaleUpDelay   time.Duration
	ScalingMode    string
	NannyCfdFlags  nannyconfigalpha.NannyConfiguration
	ConfigDir      string
	BaseStorage    string
	EstimatorType  string
	MinClusterSize uint64
	Threshold      uint64
	HealthCheck    *healthcheck.HealthCheck
	estimator      ResourceEstimator
}

// checkResource determines whether a specific resource needs to be over-written and determines type of the operation.
func (n *Nanny) checkResource(actual, expected corev1.ResourceList, res corev1.ResourceName) (bool, operation) {
	val, ok := actual[res]
	expVal, expOk := expected[res]
	if ok != expOk {
		return true, unknown
	}
	if !ok && !expOk {
		return false, unknown
	}
	q := new(inf.Dec).QuoRound(val.AsDec(), expVal.AsDec(), 2, inf.RoundDown)
	lower := inf.NewDec(100-int64(n.Threshold), 2)
	upper := inf.NewDec(100+int64(n.Threshold), 2)
	if q.Cmp(lower) == -1 {
		return true, scaleUp
	}
	if q.Cmp(upper) == 1 {
		return true, scaleDown
	}
	return false, unknown
}

// shouldOverwriteResources determines if we should over-write the container's resource limits and determines type of the operation.
// We'll over-write the resource limits if the limited resources are different, or if any limit is violated by a threshold.
// Returned operation type (scale up/down or unknown) is calculated based on values taken from the first resource which requires overwrite.
func (n *Nanny) shouldOverwriteResources(limits, reqs, expLimits, expReqs corev1.ResourceList) (bool, operation) {
	for _, resourceType := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory, corev1.ResourceStorage} {
		overwrite, op := n.checkResource(limits, expLimits, resourceType)
		if overwrite {
			return true, op
		}
		overwrite, op = n.checkResource(reqs, expReqs, resourceType)
		if overwrite {
			return true, op
		}
	}
	return false, unknown
}

// KubernetesClient is an object that performs the nanny's requisite interactions with Kubernetes.
type KubernetesClient interface {
	CountNodes() (uint64, error)
	CountContainers() (uint64, error)
	ContainerResources() (*corev1.ResourceRequirements, error)
	UpdateDeployment(resources *corev1.ResourceRequirements) error
}

// ResourceEstimator estimates ResourceRequirements for a given criteria.
type ResourceEstimator interface {
	scale(clusterSize uint64) *corev1.ResourceRequirements
}

// PollAPIServer periodically counts the size of the cluster, estimates the expected
// ResourceRequirements, compares them to the actual ResourceRequirements, and
// updates the deployment with the expected ResourceRequirements if necessary.
func (n *Nanny) pollAPIServer() {
	lastChange := time.Now()
	lastResult := noChange

	for i := 0; true; i++ {
		select {
		case <-n.StopChan:
			return
		default:
			if i != 0 {
				// Sleep for the poll period.
				time.Sleep(n.PollPeriod)
			}

			if lastResult = n.updateResources(time.Now(), lastChange, lastResult); lastResult == overwrite {
				lastChange = time.Now()
			}
			n.HealthCheck.UpdateLastActivity()
		}
	}
}

// Run would trigger the nanny to get the estimator based on the command line flags and configuration file
// and then trigger the API polling which periodically counts the number of nodes, estimates the expected
// ResourceRequirements, compares them to the actual ResourceRequirements, and
// updates the deployment with the expected ResourceRequirements if necessary.
func (n *Nanny) Run() {
	estimator, err := n.getEstimator()
	if err != nil {
		log.Fatal(err)
	}
	n.estimator = estimator
	go n.pollAPIServer()
}

func (n *Nanny) loadConfiguration(configDir string, defaultConfig *nannyconfigalpha.NannyConfiguration) (*nannyconfig.NannyConfiguration, error) {
	path := filepath.Join(configDir, "NannyConfiguration")
	_, codecs, err := nannyscheme.NewSchemeAndCodecs()
	if err != nil {
		return nil, err
	}
	// overwrite defaults with flag-specified parameters
	nannyconfigalpha.SetDefaults_NannyConfiguration(defaultConfig)
	// retrieve config map parameters if present
	configMapConfig := &nannyconfigalpha.NannyConfiguration{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.V(0).Infof("Failed to read data from config file %q: %v, using default parameters", path, err)
	} else if configMapConfig, err = n.decodeConfiguration(data, codecs); err != nil {
		configMapConfig = &nannyconfigalpha.NannyConfiguration{}
		log.V(0).Infof("Unable to decode Nanny Configuration from config map, using default parameters")
	}
	nannyconfigalpha.SetDefaults_NannyConfiguration(configMapConfig)
	// overwrite defaults with config map parameters
	nannyconfigalpha.FillInDefaults_NannyConfiguration(configMapConfig, defaultConfig)
	return n.convertConfiguration(configMapConfig), nil
}

func (n *Nanny) convertConfiguration(configAlpha *nannyconfigalpha.NannyConfiguration) *nannyconfig.NannyConfiguration {
	if configAlpha == nil {
		return nil
	}
	return &nannyconfig.NannyConfiguration{
		TypeMeta:      configAlpha.TypeMeta,
		BaseCPU:       configAlpha.BaseCPU,
		CPUPerNode:    configAlpha.CPUPerNode,
		BaseMemory:    configAlpha.BaseMemory,
		MemoryPerNode: configAlpha.MemoryPerNode,
	}
}

func (n *Nanny) decodeConfiguration(data []byte, codecs *serializer.CodecFactory) (*nannyconfigalpha.NannyConfiguration, error) {
	obj, err := runtime.Decode(codecs.UniversalDecoder(nannyconfigalpha.SchemeGroupVersion), data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode, error: %v", err)
	}
	externalHC, ok := obj.(*nannyconfigalpha.NannyConfiguration)
	if !ok {
		return nil, fmt.Errorf("failed to cast object to NannyConfiguration, object: %#v", obj)
	}
	return externalHC, nil
}

func (n *Nanny) getEstimator() (ResourceEstimator, error) {
	nannycfg, err := n.loadConfiguration(n.ConfigDir, &n.ConfigurationFlags)
	if err != nil {
		return nil, err
	}
	log.Infof("cpu: %s, extra_cpu: %s, memory: %s, extra_memory: %s", nannycfg.BaseCPU, nannycfg.CPUPerNode, nannycfg.BaseMemory, nannycfg.MemoryPerNode)

	var resources []Resource

	// Monitor only the resources specified.
	if nannycfg.BaseCPU != nannyconfig.NoValue {
		resources = append(resources, Resource{
			Base:             resource.MustParse(nannycfg.BaseCPU),
			ExtraPerResource: resource.MustParse(nannycfg.CPUPerNode),
			Name:             "cpu",
		})
	}

	if nannycfg.BaseMemory != nannyconfig.NoValue {
		resources = append(resources, Resource{
			Base:             resource.MustParse(nannycfg.BaseMemory),
			ExtraPerResource: resource.MustParse(nannycfg.MemoryPerNode),
			Name:             "memory",
		})
	}

	if n.BaseStorage != nannyconfig.NoValue {
		resources = append(resources, Resource{
			Base:             resource.MustParse(n.BaseStorage),
			ExtraPerResource: resource.MustParse(nannycfg.MemoryPerNode),
			Name:             "storage",
		})
	}

	log.Infof("Resources: %+v", resources)

	var est ResourceEstimator
	if n.EstimatorType == "linear" {
		est = LinearEstimator{
			Resources: resources,
		}
	} else if n.EstimatorType == "exponential" {
		est = ExponentialEstimator{
			Resources:      resources,
			ScaleFactor:    1.5,
			MinClusterSize: n.MinClusterSize,
		}
	} else {
		log.Fatalf("Estimator %s not supported", n.EstimatorType)
	}

	return est, nil
}

func (n *Nanny) Stop() {
	select {
	case n.StopChan <- struct{}{}:
	default:
	}
}

// updateResources counts the cluster size, estimates the expected
// ResourceRequirements, compares them to the actual ResourceRequirements, and
// updates the deployment with the expected ResourceRequirements if necessary.
// It returns overwrite if deployment has been updated, postpone if the change
// could not be applied due to scale up/down delay and noChange if the estimated
// expected ResourceRequirements are in line with the actual ResourceRequirements.
func (n *Nanny) updateResources(now, lastChange time.Time, prevResult updateResult) updateResult {
	// Query the apiserver for the cluster size.
	var num uint64
	var err error
	if n.ScalingMode == ContainerProportional {
		num, err = n.Client.CountContainers()
	} else {
		num, err = n.Client.CountNodes()
	}

	if err != nil {
		log.Error(err)
		return noChange
	}
	log.V(4).Infof("The cluster size is %d", num)

	// Query the apiserver for this pod's information.
	resources, err := n.Client.ContainerResources()
	if err != nil {
		log.Errorf("Error while querying apiserver for resources: %v", err)
		return noChange
	}

	// Get the expected resource limits.
	expResources := n.estimator.scale(num)

	// If there's a difference, go ahead and set the new values.
	overwriteRes, op := n.shouldOverwriteResources(resources.Limits, resources.Requests, expResources.Limits, expResources.Requests)
	if !overwriteRes {
		if log.V(4) {
			log.V(4).Infof("Resources are within the expected limits. Actual: %+v Expected: %+v", jsonOrValue(*resources), jsonOrValue(*expResources))
		}
		return noChange
	}

	if (op == scaleDown && now.Before(lastChange.Add(n.ScaleDownDelay))) ||
		(op == scaleUp && now.Before(lastChange.Add(n.ScaleUpDelay))) {
		if prevResult != postpone {
			log.Infof("Resources are not within the expected limits, Actual: %+v, accepted range: %+v. Skipping resource update because of scale up/down delay", jsonOrValue(*resources), jsonOrValue(*expResources))
		}
		return postpone
	}

	log.Infof("Resources are not within the expected limits, updating the deployment. Actual: %+v Expected: %+v. Resource will be updated.", jsonOrValue(*resources), jsonOrValue(*expResources))
	if err := n.Client.UpdateDeployment(expResources); err != nil {
		log.Error(err)
		return noChange
	}
	return overwrite
}

func jsonOrValue(val interface{}) interface{} {
	bytes, err := json.Marshal(val)
	if err != nil {
		return val
	}
	return string(bytes)
}
