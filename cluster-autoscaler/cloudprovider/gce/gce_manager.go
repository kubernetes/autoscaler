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
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	gcfg "gopkg.in/gcfg.v1"

	"cloud.google.com/go/compute/metadata"
	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gce "google.golang.org/api/compute/v1"
	gke "google.golang.org/api/container/v1"
	gke_alpha "google.golang.org/api/container/v1alpha1"
	gke_beta "google.golang.org/api/container/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	provider_gce "k8s.io/kubernetes/pkg/cloudprovider/providers/gce"
)

// TODO(krzysztof-jastrzebski): Move to main.go.
var (
	gkeAPIEndpoint = flag.String("gke-api-endpoint", "", "GKE API endpoint address. This flag is used by developers only. Users shouldn't change this flag.")
)

var (
	// This makes me so sad
	taintEffectsMap = map[apiv1.TaintEffect]string{
		apiv1.TaintEffectNoSchedule:       "NO_SCHEDULE",
		apiv1.TaintEffectPreferNoSchedule: "PREFER_NO_SCHEDULE",
		apiv1.TaintEffectNoExecute:        "NO_EXECUTE",
	}
)

// GcpCloudProviderMode allows to pass information whether the cluster is GCE or GKE.
type GcpCloudProviderMode string

const (
	// ModeGCE means that the cluster is running on gce (or using the legacy gke setup).
	ModeGCE GcpCloudProviderMode = "gce"

	// ModeGKE means that the cluster is running
	ModeGKE GcpCloudProviderMode = "gke"

	// ModeGKENAP means that the cluster is running on GKE with autoprovisioning enabled.
	// TODO(maciekpytel): remove this when NAP API is available in normal client
	ModeGKENAP GcpCloudProviderMode = "gke_nap"
)

const (
	operationWaitTimeout       = 5 * time.Second
	gkeOperationWaitTimeout    = 120 * time.Second
	operationPollInterval      = 100 * time.Millisecond
	refreshInterval            = 1 * time.Minute
	nodeAutoprovisioningPrefix = "nap"
	napMaxNodes                = 1000
	napMinNodes                = 0
)

var (
	defaultOAuthScopes []string = []string{
		"https://www.googleapis.com/auth/compute",
		"https://www.googleapis.com/auth/devstorage.read_only",
		"https://www.googleapis.com/auth/service.management.readonly",
		"https://www.googleapis.com/auth/servicecontrol"}
	supportedResources = map[string]bool{cloudprovider.ResourceNameCores: true, cloudprovider.ResourceNameMemory: true}
)

type migInformation struct {
	config   *Mig
	basename string
}

// GceManager handles gce communication and data caching.
type GceManager interface {
	// RegisterMig registers mig in Gce Manager. Returns true if the node group didn't exist before.
	RegisterMig(mig *Mig) bool
	// UnregisterMig unregisters mig in Gce Manager. Returns true if the node group has been removed.
	UnregisterMig(toBeRemoved *Mig) bool
	// GetMigSize gets MIG size.
	GetMigSize(mig *Mig) (int64, error)
	// SetMigSize sets MIG size.
	SetMigSize(mig *Mig, size int64) error
	// DeleteInstances deletes the given instances. All instances must be controlled by the same MIG.
	DeleteInstances(instances []*GceRef) error
	// GetMigForInstance returns MigConfig of the given Instance
	GetMigForInstance(instance *GceRef) (*Mig, error)
	// GetMigNodes returns mig nodes.
	GetMigNodes(mig *Mig) ([]string, error)
	// Refresh updates config by calling GKE API (in GKE mode only).
	Refresh() error
	// GetResourceLimiter returns resource limiter.
	GetResourceLimiter() (*cloudprovider.ResourceLimiter, error)
	// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
	Cleanup() error
	getMigs() []*migInformation
	createNodePool(mig *Mig) error
	deleteNodePool(toBeRemoved *Mig) error
	getLocation() string
	getProjectId() string
	getMode() GcpCloudProviderMode
	getTemplates() *templateBuilder
}

// gceManagerImpl handles gce communication and data caching.
type gceManagerImpl struct {
	migs     []*migInformation
	migCache map[GceRef]*Mig

	gceService      *gce.Service
	gkeService      *gke.Service
	gkeAlphaService *gke_alpha.Service
	gkeBetaService  *gke_beta.Service

	cacheMutex sync.Mutex
	migsMutex  sync.Mutex

	location        string
	projectId       string
	clusterName     string
	mode            GcpCloudProviderMode
	templates       *templateBuilder
	interrupt       chan struct{}
	isRegional      bool
	resourceLimiter *cloudprovider.ResourceLimiter
	lastRefresh     time.Time
}

// CreateGceManager constructs gceManager object.
func CreateGceManager(configReader io.Reader, mode GcpCloudProviderMode, clusterName string) (GceManager, error) {
	// Create Google Compute Engine token.
	var err error
	tokenSource := google.ComputeTokenSource("")
	if len(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")) > 0 {
		tokenSource, err = google.DefaultTokenSource(oauth2.NoContext, gce.ComputeScope)
		if err != nil {
			return nil, err
		}
	}
	var projectId, location string
	var isRegional bool
	if configReader != nil {
		var cfg provider_gce.ConfigFile
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			glog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
		if cfg.Global.TokenURL == "" {
			glog.Warning("Empty tokenUrl in cloud config")
		} else {
			tokenSource = provider_gce.NewAltTokenSource(cfg.Global.TokenURL, cfg.Global.TokenBody)
			glog.V(1).Infof("Using TokenSource from config %#v", tokenSource)
		}
		projectId = cfg.Global.ProjectID
		isRegional = cfg.Global.Multizone
		location = cfg.Global.LocalZone
	} else {
		glog.V(1).Infof("Using default TokenSource %#v", tokenSource)
	}
	if len(projectId) == 0 || len(location) == 0 {
		// XXX: On GKE discoveredProjectId is hosted master project and
		// not the project we want to use, however, zone seems to not
		// be specified in config. For now we can just assume that hosted
		// master project is in the same zone as cluster and only use
		// discoveredZone.
		discoveredProjectId, discoveredLocation, err := getProjectAndLocation(isRegional)
		if err != nil {
			return nil, err
		}
		if len(projectId) == 0 {
			projectId = discoveredProjectId
		}
		if len(location) == 0 {
			location = discoveredLocation
		}
	}
	glog.V(1).Infof("GCE projectId=%s location=%s", projectId, location)

	// Create Google Compute Engine service.
	client := oauth2.NewClient(oauth2.NoContext, tokenSource)
	gceService, err := gce.New(client)
	if err != nil {
		return nil, err
	}
	manager := &gceManagerImpl{
		migs:        make([]*migInformation, 0),
		gceService:  gceService,
		migCache:    make(map[GceRef]*Mig),
		location:    location,
		isRegional:  isRegional,
		projectId:   projectId,
		clusterName: clusterName,
		mode:        mode,
		templates: &templateBuilder{
			projectId: projectId,
			service:   gceService,
		},
		interrupt: make(chan struct{}),
	}

	if mode == ModeGKE {
		gkeService, err := gke.New(client)
		if err != nil {
			return nil, err
		}
		if *gkeAPIEndpoint != "" {
			gkeService.BasePath = *gkeAPIEndpoint
		}
		manager.gkeService = gkeService
		if manager.isRegional {
			gkeBetaService, err := gke_beta.New(client)
			if err != nil {
				return nil, err
			}
			if *gkeAPIEndpoint != "" {
				gkeBetaService.BasePath = *gkeAPIEndpoint
			}
			manager.gkeBetaService = gkeBetaService
		}
		err = manager.fetchAllNodePools()
		if err != nil {
			glog.Errorf("Failed to fetch node pools: %v", err)
			return nil, err
		}
	}

	if mode == ModeGKENAP {
		gkeAlphaService, err := gke_alpha.New(client)
		if err != nil {
			return nil, err
		}
		if *gkeAPIEndpoint != "" {
			gkeAlphaService.BasePath = *gkeAPIEndpoint
		}
		manager.gkeAlphaService = gkeAlphaService
		err = manager.fetchAllNodePools()
		if err != nil {
			glog.Errorf("Failed to fetch node pools: %v", err)
			return nil, err
		}
		err = manager.fetchResourceLimiter()
		if err != nil {
			glog.Errorf("Failed to fetch resource limits: %v", err)
			return nil, err
		}
		glog.V(1).Info("Using GKE-NAP mode")
	}

	manager.lastRefresh = time.Now()

	go wait.Until(func() {
		manager.cacheMutex.Lock()
		defer manager.cacheMutex.Unlock()
		if err := manager.regenerateCache(); err != nil {
			glog.Errorf("Error while regenerating Mig cache: %v", err)
		}
	}, time.Hour, manager.interrupt)

	return manager, nil
}

// Cleanup closes the channel to signal the go routine to stop that is handling the cache
func (m *gceManagerImpl) Cleanup() error {
	close(m.interrupt)
	return nil
}

func (m *gceManagerImpl) assertGKE() {
	if m.mode != ModeGKE {
		glog.Fatalf("This should run only in GKE mode")
	}
}

// Following section is a mess, because we need to use GKE v1alpha1 for NAP,
// v1beta1 for regional clusters and v1 in the regular case.
// TODO(maciekpytel): Clean this up once NAP fields are promoted to v1beta1

func (m *gceManagerImpl) assertGKENAP() {
	if m.mode != ModeGKENAP {
		glog.Fatalf("This should run only in GKE mode with autoprovisioning enabled")
	}
}

func (m *gceManagerImpl) fetchAllNodePools() error {
	if m.mode == ModeGKENAP {
		return m.fetchAllNodePoolsGkeNapImpl()
	}
	if m.isRegional {
		return m.fetchAllNodePoolsGkeRegionalImpl()
	}
	return m.fetchAllNodePoolsGkeImpl()
}

// Gets all registered node pools
func (m *gceManagerImpl) fetchAllNodePoolsGkeImpl() error {
	m.assertGKE()

	nodePoolsResponse, err := m.gkeService.Projects.Zones.Clusters.NodePools.List(m.projectId, m.location, m.clusterName).Do()
	if err != nil {
		return err
	}

	existingMigs := map[GceRef]struct{}{}
	changed := false

	for _, nodePool := range nodePoolsResponse.NodePools {
		autoscaled := nodePool.Autoscaling != nil && nodePool.Autoscaling.Enabled
		if !autoscaled {
			continue
		}
		// format is
		// "https://www.googleapis.com/compute/v1/projects/mwielgus-proj/zones/europe-west1-b/instanceGroupManagers/gke-cluster-1-default-pool-ba78a787-grp"
		for _, igurl := range nodePool.InstanceGroupUrls {
			project, zone, name, err := parseGceUrl(igurl, "instanceGroupManagers")
			if err != nil {
				return err
			}
			mig := &Mig{
				GceRef: GceRef{
					Name:    name,
					Zone:    zone,
					Project: project,
				},
				gceManager:      m,
				exist:           true,
				autoprovisioned: false, // NAP is disabled
				nodePoolName:    nodePool.Name,
				minSize:         int(nodePool.Autoscaling.MinNodeCount),
				maxSize:         int(nodePool.Autoscaling.MaxNodeCount),
			}
			existingMigs[mig.GceRef] = struct{}{}

			if m.RegisterMig(mig) {
				changed = true
			}
		}
	}
	for _, mig := range m.getMigs() {
		if _, found := existingMigs[mig.config.GceRef]; !found {
			m.UnregisterMig(mig.config)
			changed = true
		}
	}
	if changed {
		m.cacheMutex.Lock()
		defer m.cacheMutex.Unlock()

		if err := m.regenerateCache(); err != nil {
			return err
		}
	}
	return nil
}

// Gets all registered node pools
func (m *gceManagerImpl) fetchAllNodePoolsGkeRegionalImpl() error {
	m.assertGKE()

	nodePoolsResponse, err := m.gkeBetaService.Projects.Locations.Clusters.NodePools.List(fmt.Sprintf("projects/%s/locations/%s/clusters/%s", m.projectId, m.location, m.clusterName)).Do()
	if err != nil {
		return err
	}

	existingMigs := map[GceRef]struct{}{}
	changed := false

	for _, nodePool := range nodePoolsResponse.NodePools {
		autoscaled := nodePool.Autoscaling != nil && nodePool.Autoscaling.Enabled
		if !autoscaled {
			continue
		}
		// format is
		// "https://www.googleapis.com/compute/v1/projects/mwielgus-proj/zones/europe-west1-b/instanceGroupManagers/gke-cluster-1-default-pool-ba78a787-grp"
		for _, igurl := range nodePool.InstanceGroupUrls {
			project, zone, name, err := parseGceUrl(igurl, "instanceGroupManagers")
			if err != nil {
				return err
			}
			mig := &Mig{
				GceRef: GceRef{
					Name:    name,
					Zone:    zone,
					Project: project,
				},
				gceManager:      m,
				exist:           true,
				autoprovisioned: false, // NAP is disabled
				nodePoolName:    nodePool.Name,
				minSize:         int(nodePool.Autoscaling.MinNodeCount),
				maxSize:         int(nodePool.Autoscaling.MaxNodeCount),
			}
			existingMigs[mig.GceRef] = struct{}{}

			if m.RegisterMig(mig) {
				changed = true
			}
		}
	}
	for _, mig := range m.getMigs() {
		if _, found := existingMigs[mig.config.GceRef]; !found {
			m.UnregisterMig(mig.config)
			changed = true
		}
	}
	if changed {
		m.cacheMutex.Lock()
		defer m.cacheMutex.Unlock()

		if err := m.regenerateCache(); err != nil {
			return err
		}
	}
	return nil
}

// Gets all registered node pools
func (m *gceManagerImpl) fetchAllNodePoolsGkeNapImpl() error {
	m.assertGKENAP()

	nodePoolsResponse, err := m.gkeAlphaService.Projects.Zones.Clusters.NodePools.List(m.projectId, m.location, m.clusterName).Do()
	if err != nil {
		return err
	}

	existingMigs := map[GceRef]struct{}{}
	changed := false

	for _, nodePool := range nodePoolsResponse.NodePools {
		autoprovisioned := nodePool.Autoscaling != nil && nodePool.Autoscaling.Autoprovisioned
		autoscaled := nodePool.Autoscaling != nil && nodePool.Autoscaling.Enabled
		if !autoscaled {
			if autoprovisioned {
				glog.Warningf("NodePool %v has invalid config - autoprovisioned, but not autoscaled. Ignoring this NodePool.", nodePool.Name)
			}
			continue
		}
		// format is
		// "https://www.googleapis.com/compute/v1/projects/mwielgus-proj/zones/europe-west1-b/instanceGroupManagers/gke-cluster-1-default-pool-ba78a787-grp"
		for _, igurl := range nodePool.InstanceGroupUrls {
			project, zone, name, err := parseGceUrl(igurl, "instanceGroupManagers")
			if err != nil {
				return err
			}
			mig := &Mig{
				GceRef: GceRef{
					Name:    name,
					Zone:    zone,
					Project: project,
				},
				gceManager:      m,
				exist:           true,
				autoprovisioned: autoprovisioned,
				nodePoolName:    nodePool.Name,
				minSize:         int(nodePool.Autoscaling.MinNodeCount),
				maxSize:         int(nodePool.Autoscaling.MaxNodeCount),
			}
			existingMigs[mig.GceRef] = struct{}{}

			if m.RegisterMig(mig) {
				changed = true
			}
		}
	}
	for _, mig := range m.getMigs() {
		if _, found := existingMigs[mig.config.GceRef]; !found {
			m.UnregisterMig(mig.config)
			changed = true
		}
	}
	if changed {
		m.cacheMutex.Lock()
		defer m.cacheMutex.Unlock()

		if err := m.regenerateCache(); err != nil {
			return err
		}
	}
	return nil
}

// RegisterMig registers mig in Gce Manager. Returns true if the node group didn't exist before or its config has changed.
func (m *gceManagerImpl) RegisterMig(mig *Mig) bool {
	m.migsMutex.Lock()
	defer m.migsMutex.Unlock()

	for i := range m.migs {
		if oldMig := m.migs[i].config; oldMig.GceRef == mig.GceRef {
			if !reflect.DeepEqual(oldMig, mig) {
				m.migs[i].config = mig
				glog.V(4).Infof("Updated Mig %s/%s/%s", mig.GceRef.Project, mig.GceRef.Zone, mig.GceRef.Name)
				return true
			}
			return false
		}
	}

	glog.V(1).Infof("Registering %s/%s/%s", mig.GceRef.Project, mig.GceRef.Zone, mig.GceRef.Name)
	m.migs = append(m.migs, &migInformation{
		config: mig,
	})

	template, err := m.templates.getMigTemplate(mig)
	if err != nil {
		glog.Errorf("Failed to build template for %s", mig.Name)
	} else {
		_, err = m.templates.buildNodeFromTemplate(mig, template)
		if err != nil {
			glog.Errorf("Failed to build template for %s", mig.Name)
		}
	}
	return true
}

// UnregisterMig unregisters mig in Gce Manager. Returns true if the node group has been removed.
func (m *gceManagerImpl) UnregisterMig(toBeRemoved *Mig) bool {
	m.migsMutex.Lock()
	defer m.migsMutex.Unlock()

	newMigs := make([]*migInformation, 0, len(m.migs))
	found := false
	for _, mig := range m.migs {
		if mig.config.GceRef == toBeRemoved.GceRef {
			glog.V(1).Infof("Unregistered Mig %s/%s/%s", toBeRemoved.GceRef.Project, toBeRemoved.GceRef.Zone,
				toBeRemoved.GceRef.Name)
			found = true
		} else {
			newMigs = append(newMigs, mig)
		}
	}
	m.migs = newMigs
	return found
}

func (m *gceManagerImpl) deleteNodePool(toBeRemoved *Mig) error {
	m.assertGKENAP()
	if !toBeRemoved.Autoprovisioned() {
		return fmt.Errorf("only autoprovisioned node pools can be deleted")
	}
	// TODO: handle multi-zonal node pools.
	deleteOp, err := m.gkeAlphaService.Projects.Zones.Clusters.NodePools.Delete(m.projectId, m.location, m.clusterName,
		toBeRemoved.nodePoolName).Do()
	if err != nil {
		return err
	}
	err = m.waitForGkeOp(deleteOp)
	if err != nil {
		return err
	}
	return m.fetchAllNodePools()
}

func (m *gceManagerImpl) createNodePool(mig *Mig) error {
	m.assertGKENAP()

	// TODO: handle preemptable
	// TODO: handle ssd

	accelerators := []*gke_alpha.AcceleratorConfig{}

	if gpuRequest, found := mig.spec.extraResources[gpu.ResourceNvidiaGPU]; found {
		gpuType, found := mig.spec.labels[gpu.GPULabel]
		if !found {
			return fmt.Errorf("failed to create node pool %v with gpu request of unspecified type", mig.nodePoolName)
		}
		gpuConfig := &gke_alpha.AcceleratorConfig{
			AcceleratorType:  gpuType,
			AcceleratorCount: gpuRequest.Value(),
		}
		accelerators = append(accelerators, gpuConfig)

	}

	taints := []*gke_alpha.NodeTaint{}
	for _, taint := range mig.spec.taints {
		effect, found := taintEffectsMap[taint.Effect]
		if !found {
			effect = "EFFECT_UNSPECIFIED"
		}
		taint := &gke_alpha.NodeTaint{
			Effect: effect,
			Key:    taint.Key,
			Value:  taint.Value,
		}
		taints = append(taints, taint)
	}

	labels := make(map[string]string)
	for k, v := range mig.spec.labels {
		if k != gpu.GPULabel {
			labels[k] = v
		}
	}

	config := gke_alpha.NodeConfig{
		MachineType:  mig.spec.machineType,
		OauthScopes:  defaultOAuthScopes,
		Labels:       labels,
		Accelerators: accelerators,
		Taints:       taints,
	}

	autoscaling := gke_alpha.NodePoolAutoscaling{
		Enabled:         true,
		MinNodeCount:    napMinNodes,
		MaxNodeCount:    napMaxNodes,
		Autoprovisioned: true,
	}

	createRequest := gke_alpha.CreateNodePoolRequest{
		NodePool: &gke_alpha.NodePool{
			Name:             mig.nodePoolName,
			InitialNodeCount: 0,
			Config:           &config,
			Autoscaling:      &autoscaling,
		},
	}

	createOp, err := m.gkeAlphaService.Projects.Zones.Clusters.NodePools.Create(m.projectId, m.location, m.clusterName,
		&createRequest).Do()
	if err != nil {
		return err
	}
	err = m.waitForGkeOp(createOp)
	if err != nil {
		return err
	}
	err = m.fetchAllNodePools()
	if err != nil {
		return err
	}
	for _, existingMig := range m.getMigs() {
		if existingMig.config.nodePoolName == mig.nodePoolName {
			*mig = *existingMig.config
			return nil
		}
	}
	return fmt.Errorf("node pool %s not found", mig.nodePoolName)
}

// End of v1alpha1/v1beta1 mess

// GetMigSize gets MIG size.
func (m *gceManagerImpl) GetMigSize(mig *Mig) (int64, error) {
	igm, err := m.gceService.InstanceGroupManagers.Get(mig.Project, mig.Zone, mig.Name).Do()
	if err != nil {
		return -1, err
	}
	return igm.TargetSize, nil
}

// SetMigSize sets MIG size.
func (m *gceManagerImpl) SetMigSize(mig *Mig, size int64) error {
	glog.V(0).Infof("Setting mig size %s to %d", mig.Id(), size)
	op, err := m.gceService.InstanceGroupManagers.Resize(mig.Project, mig.Zone, mig.Name, size).Do()
	if err != nil {
		return err
	}
	return m.waitForOp(op, mig.Project, mig.Zone)
}

// GCE
func (m *gceManagerImpl) waitForOp(operation *gce.Operation, project string, zone string) error {
	for start := time.Now(); time.Since(start) < operationWaitTimeout; time.Sleep(operationPollInterval) {
		glog.V(4).Infof("Waiting for operation %s %s %s", project, zone, operation.Name)
		if op, err := m.gceService.ZoneOperations.Get(project, zone, operation.Name).Do(); err == nil {
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

//  GKE
func (m *gceManagerImpl) waitForGkeOp(operation *gke_alpha.Operation) error {
	for start := time.Now(); time.Since(start) < gkeOperationWaitTimeout; time.Sleep(operationPollInterval) {
		glog.V(4).Infof("Waiting for operation %s %s %s", m.projectId, m.location, operation.Name)
		if op, err := m.gkeAlphaService.Projects.Zones.Operations.Get(m.projectId, m.location, operation.Name).Do(); err == nil {
			glog.V(4).Infof("Operation %s %s %s status: %s", m.projectId, m.location, operation.Name, op.Status)
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
func (m *gceManagerImpl) DeleteInstances(instances []*GceRef) error {
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

	op, err := m.gceService.InstanceGroupManagers.DeleteInstances(commonMig.Project, commonMig.Zone, commonMig.Name, &req).Do()
	if err != nil {
		return err
	}
	return m.waitForOp(op, commonMig.Project, commonMig.Zone)
}

func (m *gceManagerImpl) getMigs() []*migInformation {
	m.migsMutex.Lock()
	defer m.migsMutex.Unlock()
	migs := make([]*migInformation, 0, len(m.migs))
	for _, mig := range m.migs {
		migs = append(migs, &migInformation{
			basename: mig.basename,
			config:   mig.config,
		})
	}
	return migs
}
func (m *gceManagerImpl) updateMigBasename(ref GceRef, basename string) {
	m.migsMutex.Lock()
	defer m.migsMutex.Unlock()
	for _, mig := range m.migs {
		if mig.config.GceRef == ref {
			mig.basename = basename
		}
	}
}

// GetMigForInstance returns MigConfig of the given Instance
func (m *gceManagerImpl) GetMigForInstance(instance *GceRef) (*Mig, error) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	if mig, found := m.migCache[*instance]; found {
		return mig, nil
	}

	for _, mig := range m.getMigs() {
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

func (m *gceManagerImpl) regenerateCache() error {
	newMigCache := make(map[GceRef]*Mig)

	for _, migInfo := range m.getMigs() {
		mig := migInfo.config
		glog.V(4).Infof("Regenerating MIG information for %s %s %s", mig.Project, mig.Zone, mig.Name)

		instanceGroupManager, err := m.gceService.InstanceGroupManagers.Get(mig.Project, mig.Zone, mig.Name).Do()
		if err != nil {
			return err
		}
		m.updateMigBasename(migInfo.config.GceRef, instanceGroupManager.BaseInstanceName)

		instances, err := m.gceService.InstanceGroupManagers.ListManagedInstances(mig.Project, mig.Zone, mig.Name).Do()
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
func (m *gceManagerImpl) GetMigNodes(mig *Mig) ([]string, error) {
	instances, err := m.gceService.InstanceGroupManagers.ListManagedInstances(mig.Project, mig.Zone, mig.Name).Do()
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

func (m *gceManagerImpl) getLocation() string {
	return m.location
}
func (m *gceManagerImpl) getProjectId() string {
	return m.projectId
}
func (m *gceManagerImpl) getMode() GcpCloudProviderMode {
	return m.mode
}
func (m *gceManagerImpl) getTemplates() *templateBuilder {
	return m.templates
}

func (m *gceManagerImpl) Refresh() error {
	if m.mode == ModeGCE {
		return nil
	}
	if m.lastRefresh.Add(refreshInterval).Before(time.Now()) {
		err := m.fetchAllNodePools()
		if err != nil {
			return err
		}

		err = m.fetchResourceLimiter()
		if err != nil {
			return err
		}

		m.lastRefresh = time.Now()
		glog.V(2).Infof("Refreshed NodePools list and resource limits, next refresh after %v", m.lastRefresh.Add(refreshInterval))
		return nil
	}
	return nil
}

func (m *gceManagerImpl) fetchResourceLimiter() error {
	if m.mode == ModeGKENAP {
		cluster, err := m.gkeAlphaService.Projects.Zones.Clusters.Get(m.projectId, m.location, m.clusterName).Do()
		if err != nil {
			return err
		}
		if cluster.Autoscaling == nil {
			return nil
		}

		minLimits := make(map[string]int64)
		maxLimits := make(map[string]int64)
		for _, limit := range cluster.Autoscaling.ResourceLimits {
			if _, found := supportedResources[limit.Name]; !found {
				glog.Warning("Unsupported limit defined %s: %d - %d", limit.Name, limit.Minimum, limit.Maximum)
			}
			minLimits[limit.Name] = limit.Minimum
			maxLimits[limit.Name] = limit.Maximum
		}

		// GKE API provides memory in GB, but ResourceLimiter expects them in MB
		if _, found := minLimits[cloudprovider.ResourceNameMemory]; found {
			minLimits[cloudprovider.ResourceNameMemory] = minLimits[cloudprovider.ResourceNameMemory] * 1024
		}
		if _, found := maxLimits[cloudprovider.ResourceNameMemory]; found {
			maxLimits[cloudprovider.ResourceNameMemory] = maxLimits[cloudprovider.ResourceNameMemory] * 1024
		}

		resourceLimiter := cloudprovider.NewResourceLimiter(minLimits, maxLimits)
		glog.V(2).Infof("Refreshed resource limits: %s", resourceLimiter.String())

		m.cacheMutex.Lock()
		defer m.cacheMutex.Unlock()
		m.resourceLimiter = resourceLimiter
	}
	return nil
}

// GetResourceLimiter returns resource limiter.
func (m *gceManagerImpl) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	return m.resourceLimiter, nil
}

// Code borrowed from gce cloud provider. Reuse the original as soon as it becomes public.
func getProjectAndLocation(isRegional bool) (string, string, error) {
	result, err := metadata.Get("instance/zone")
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(result, "/")
	if len(parts) != 4 {
		return "", "", fmt.Errorf("unexpected response: %s", result)
	}
	location := parts[3]
	if isRegional {
		location, err = provider_gce.GetGCERegion(location)
		if err != nil {
			return "", "", err
		}
	}
	projectID, err := metadata.ProjectID()
	if err != nil {
		return "", "", err
	}
	return projectID, location, nil
}
