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

package gce

import (
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"gopkg.in/gcfg.v1"

	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gce "google.golang.org/api/compute/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	provider_gce "k8s.io/kubernetes/pkg/cloudprovider/providers/gce"
)

const (
	operationWaitTimeout  = 5 * time.Second
	operationPollInterval = 100 * time.Millisecond
)

type migInformation struct {
	config   *Mig
	basename string
}

// GceManager is handles gce communication and data caching.
type GceManager struct {
	migs     []*migInformation
	migCache map[GceRef]*Mig

	service    *gce.Service
	cacheMutex sync.Mutex
}

// CreateGceManager constructs gceManager object.
func CreateGceManager(configReader io.Reader) (*GceManager, error) {
	// Create Google Compute Engine token.
	tokenSource := google.ComputeTokenSource("")
	if configReader != nil {
		var cfg provider_gce.Config
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			glog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
		if cfg.Global.TokenURL == "" {
			glog.Warning("Empty tokenUrl in cloud config")
		} else {
			glog.Infof("Using TokenSource from config %#v", tokenSource)
			tokenSource = provider_gce.NewAltTokenSource(cfg.Global.TokenURL, cfg.Global.TokenBody)
		}
	} else {
		glog.Infof("Using default TokenSource %#v", tokenSource)
	}

	// Create Google Compute Engine service.
	client := oauth2.NewClient(oauth2.NoContext, tokenSource)
	gceService, err := gce.New(client)
	if err != nil {
		return nil, err
	}
	manager := &GceManager{
		migs:     make([]*migInformation, 0),
		service:  gceService,
		migCache: make(map[GceRef]*Mig),
	}
	go wait.Forever(func() {
		manager.cacheMutex.Lock()
		defer manager.cacheMutex.Unlock()
		if err := manager.regenerateCache(); err != nil {
			glog.Errorf("Error while regenerating Mig cache: %v", err)
		}
	}, time.Hour)
	return manager, nil
}

// RegisterMig registers mig in Gce Manager.
func (m *GceManager) RegisterMig(mig *Mig) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.migs = append(m.migs, &migInformation{
		config: mig,
	})

	template, err := m.getMigTemplate(mig)
	if err != nil {
		glog.Errorf("Failed to build template for %s", mig.Name)
	}

	node, err := m.buildNodeFromTemplate(mig, template)
	if err != nil {
		glog.Errorf("Failed to build template for %s", mig.Name)
	}
	glog.V(4).Infof("Node template for mig %s - %#v", mig.Name, node)
}

// GetMigSize gets MIG size.
func (m *GceManager) GetMigSize(mig *Mig) (int64, error) {
	igm, err := m.service.InstanceGroupManagers.Get(mig.Project, mig.Zone, mig.Name).Do()
	if err != nil {
		return -1, err
	}
	return igm.TargetSize, nil
}

// SetMigSize sets MIG size.
func (m *GceManager) SetMigSize(mig *Mig, size int64) error {
	glog.V(0).Infof("Setting mig size %s to %d", mig.Id(), size)
	op, err := m.service.InstanceGroupManagers.Resize(mig.Project, mig.Zone, mig.Name, size).Do()
	if err != nil {
		return err
	}
	if err := m.waitForOp(op, mig.Project, mig.Zone); err != nil {
		return err
	}
	return nil
}

func (m *GceManager) waitForOp(operation *gce.Operation, project string, zone string) error {
	for start := time.Now(); time.Since(start) < operationWaitTimeout; time.Sleep(operationPollInterval) {
		glog.V(4).Infof("Waiting for operation %s %s %s", project, zone, operation.Name)
		if op, err := m.service.ZoneOperations.Get(project, zone, operation.Name).Do(); err == nil {
			glog.V(4).Infof("Operation %s %s %s status: %s", project, zone, operation.Name, op.Status)
			if op.Status == "DONE" {
				return nil
			}
		} else {
			glog.Warningf("Error while getting operation %s on %s: %v", operation.Name, operation.TargetLink, err)
		}
	}
	return fmt.Errorf("Timeout while waiting for operation %s on %s to complete.", operation.Name, operation.TargetLink)
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same MIG.
func (m *GceManager) DeleteInstances(instances []*GceRef) error {
	if len(instances) == 0 {
		return nil
	}
	commonMig, err := m.GetMigForInstance(instances[0])
	if err != nil {
		return err
	}
	for _, instance := range instances {
		mig, err := m.GetMigForInstance(instance)
		if err != nil {
			return err
		}
		if mig != commonMig {
			return fmt.Errorf("Connot delete instances which don't belong to the same MIG.")
		}
	}

	req := gce.InstanceGroupManagersDeleteInstancesRequest{
		Instances: []string{},
	}
	for _, instance := range instances {
		req.Instances = append(req.Instances, GenerateInstanceUrl(instance.Project, instance.Zone, instance.Name))
	}

	op, err := m.service.InstanceGroupManagers.DeleteInstances(commonMig.Project, commonMig.Zone, commonMig.Name, &req).Do()
	if err != nil {
		return err
	}
	if err := m.waitForOp(op, commonMig.Project, commonMig.Zone); err != nil {
		return err
	}
	return nil
}

// GetMigForInstance returns MigConfig of the given Instance
func (m *GceManager) GetMigForInstance(instance *GceRef) (*Mig, error) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	if mig, found := m.migCache[*instance]; found {
		return mig, nil
	}

	for _, mig := range m.migs {
		if mig.config.Project == instance.Project &&
			mig.config.Zone == instance.Zone &&
			strings.HasPrefix(instance.Name, mig.basename) {
			if err := m.regenerateCache(); err != nil {
				return nil, fmt.Errorf("Error while looking for MIG for instance %+v, error: %v", *instance, err)
			}
			if mig, found := m.migCache[*instance]; found {
				return mig, nil
			}
			return nil, fmt.Errorf("Instance %+v does not belong to any configured MIG", *instance)
		}
	}
	// Instance doesn't belong to any configured mig.
	return nil, nil
}

func (m *GceManager) regenerateCache() error {
	newMigCache := make(map[GceRef]*Mig)

	for _, migInfo := range m.migs {
		mig := migInfo.config
		glog.V(4).Infof("Regenerating MIG information for %s %s %s", mig.Project, mig.Zone, mig.Name)

		instanceGroupManager, err := m.service.InstanceGroupManagers.Get(mig.Project, mig.Zone, mig.Name).Do()
		if err != nil {
			return err
		}
		migInfo.basename = instanceGroupManager.BaseInstanceName

		instances, err := m.service.InstanceGroupManagers.ListManagedInstances(mig.Project, mig.Zone, mig.Name).Do()
		if err != nil {
			glog.V(4).Infof("Failed MIG info request for %s %s %s: %v", mig.Project, mig.Zone, mig.Name, err)
			return err
		}
		for _, instance := range instances.ManagedInstances {
			project, zone, name, err := ParseInstanceUrl(instance.Instance)
			if err != nil {
				return err
			}
			newMigCache[GceRef{Project: project, Zone: zone, Name: name}] = mig
		}
	}

	m.migCache = newMigCache
	return nil
}

// GetMigNodes returns mig nodes.
func (m *GceManager) GetMigNodes(mig *Mig) ([]string, error) {
	instances, err := m.service.InstanceGroupManagers.ListManagedInstances(mig.Project, mig.Zone, mig.Name).Do()
	if err != nil {
		return []string{}, err
	}
	result := make([]string, 0)
	for _, instance := range instances.ManagedInstances {
		project, zone, name, err := ParseInstanceUrl(instance.Instance)
		if err != nil {
			return []string{}, err
		}
		result = append(result, fmt.Sprintf("gce://%s/%s/%s", project, zone, name))
	}
	return result, nil
}

func (m *GceManager) getMigTemplate(mig *Mig) (*gce.InstanceTemplate, error) {
	igm, err := m.service.InstanceGroupManagers.Get(mig.Project, mig.Zone, mig.Name).Do()
	if err != nil {
		return nil, err
	}
	templateUrl, err := url.Parse(igm.InstanceTemplate)
	if err != nil {
		return nil, err
	}
	_, templateName := path.Split(templateUrl.EscapedPath())
	instanceTemplate, err := m.service.InstanceTemplates.Get(mig.Project, templateName).Do()
	if err != nil {
		return nil, err
	}
	return instanceTemplate, nil
}

func (m *GceManager) buildNodeFromTemplate(mig *Mig, template *gce.InstanceTemplate) (*apiv1.Node, error) {

	if template.Properties == nil {
		return nil, fmt.Errorf("instance template %s has no properties", template.Name)
	}

	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-template-%d", template.Name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}
	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}
	// TODO: get a real value.
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)

	// TODO: handle custom !!!!
	// TODO: handle GPU
	if strings.HasPrefix(template.Properties.MachineType, "custom-") {
		cpu, mem, err := parseCustomMachineType(template.Properties.MachineType)
		if err != nil {
			return nil, err
		}
		node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(cpu, resource.DecimalSI)
		node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
	} else {
		machineType, err := m.service.MachineTypes.Get(mig.Project, mig.Zone, template.Properties.MachineType).Do()
		if err != nil {
			return nil, err
		}
		node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(machineType.GuestCpus, resource.DecimalSI)
		node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(machineType.MemoryMb*1024*1024, resource.DecimalSI)
	}

	// TODO: use proper allocatable!!
	node.Status.Allocatable = node.Status.Capacity

	// KubeEnvLabels
	if template.Properties.Metadata == nil {
		return nil, fmt.Errorf("instance template %s has no metadata", template.Name)
	}
	for _, item := range template.Properties.Metadata.Items {
		if item.Key == "kube-env" {
			if item.Value == nil {
				return nil, fmt.Errorf("no kube-env content in metadata")
			}
			kubeEnvLabels, err := extractLabelsFromKubeEnv(*item.Value)
			if err != nil {
				return nil, err
			}
			node.Labels = cloudprovider.JoinStringMaps(node.Labels, kubeEnvLabels)
		}
	}
	// GenericLabels
	labels, err := buildGenericLabels(mig.GceRef, template.Properties.MachineType, nodeName)
	if err != nil {
		return nil, err
	}
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, labels)

	// Ready status
	node.Status.Conditions = cloudprovider.BuildReadyConditions()

	return &node, nil
}

func buildGenericLabels(ref GceRef, machineType string, nodeName string) (map[string]string, error) {
	result := make(map[string]string)

	// TODO: extract it somehow
	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS

	result[kubeletapis.LabelInstanceType] = machineType
	ix := strings.LastIndex(ref.Zone, "-")
	if ix == -1 {
		return nil, fmt.Errorf("unexpected zone: %s", ref.Zone)
	}
	result[kubeletapis.LabelZoneRegion] = ref.Zone[:ix]
	result[kubeletapis.LabelZoneFailureDomain] = ref.Zone
	result[kubeletapis.LabelHostname] = nodeName
	return result, nil
}

func parseCustomMachineType(machineType string) (cpu, mem int64, err error) {
	// example custom-2-2816
	var count int
	count, err = fmt.Sscanf(machineType, "custom-%d-%d", &cpu, &mem)
	if err != nil {
		return
	}
	if count != 2 {
		return 0, 0, fmt.Errorf("failed to parse all params in %s", machineType)
	}
	// Mb to bytes
	mem = mem * 1024 * 1024
	return
}

func extractLabelsFromKubeEnv(kubeEnv string) (map[string]string, error) {
	result := make(map[string]string)

	for line, env := range strings.Split(kubeEnv, "\n") {
		env = strings.Trim(env, " ")
		if len(env) == 0 {
			continue
		}
		items := strings.SplitN(env, ":", 2)
		if len(items) != 2 {
			return nil, fmt.Errorf("wrong content in kube-env at line: %d", line)
		}
		key := strings.Trim(items[0], " ")
		value := strings.Trim(items[1], " \"'")
		if key == "NODE_LABELS" {
			for _, label := range strings.Split(value, ",") {
				labelItems := strings.SplitN(label, "=", 2)
				if len(labelItems) != 2 {
					return nil, fmt.Errorf("error while parsing label: %s", label)
				}
				result[labelItems[0]] = labelItems[1]
			}
		}
	}
	return result, nil
}
