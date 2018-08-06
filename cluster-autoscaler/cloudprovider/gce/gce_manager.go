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
	"regexp"
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
	gke_beta "google.golang.org/api/container/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
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
	gkeOperationWaitTimeout    = 120 * time.Second
	refreshInterval            = 1 * time.Minute
	machinesRefreshInterval    = 1 * time.Hour
	httpTimeout                = 30 * time.Second
	nodeAutoprovisioningPrefix = "nap"
	napMaxNodes                = 1000
	napMinNodes                = 0
	scaleToZeroSupported       = true
)

var (
	defaultOAuthScopes []string = []string{
		"https://www.googleapis.com/auth/compute",
		"https://www.googleapis.com/auth/devstorage.read_only",
		"https://www.googleapis.com/auth/service.management.readonly",
		"https://www.googleapis.com/auth/servicecontrol"}
	supportedResources = map[string]bool{}
)

func init() {
	supportedResources[cloudprovider.ResourceNameCores] = true
	supportedResources[cloudprovider.ResourceNameMemory] = true
	for _, gpuType := range supportedGpuTypes {
		supportedResources[gpuType] = true
	}
}

type migInformation struct {
	config   *Mig
	basename string
}

type machineTypeKey struct {
	zone        string
	machineType string
}

// GceManager handles gce communication and data caching.
type GceManager interface {
	// RegisterMig registers mig in GceManager. Returns true if the node group didn't exist before.
	RegisterMig(mig *Mig) bool
	// UnregisterMig unregisters mig in GceManager. Returns true if the node group has been removed.
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
	findMigsNamed(name *regexp.Regexp) ([]string, error)
	getMigTemplateNode(mig *Mig) (*apiv1.Node, error)
	getCpuAndMemoryForMachineType(machineType string, zone string) (cpu int64, mem int64, err error)
}

// gceManagerImpl handles gce communication and data caching.
type gceManagerImpl struct {
	migs     []*migInformation
	migCache map[GceRef]*Mig

	gkeService     *gke.Service
	gkeBetaService *gke_beta.Service
	GceService     AutoscalingGceClient

	cacheMutex sync.Mutex
	migsMutex  sync.Mutex

	location              string
	projectId             string
	clusterName           string
	mode                  GcpCloudProviderMode
	templates             *templateBuilder
	interrupt             chan struct{}
	regional              bool
	explicitlyConfigured  map[GceRef]bool
	migAutoDiscoverySpecs []cloudprovider.MIGAutoDiscoveryConfig
	resourceLimiter       *cloudprovider.ResourceLimiter
	lastRefresh           time.Time

	machinesCache            map[machineTypeKey]*gce.MachineType
	machinesCacheLastRefresh time.Time
}

// CreateGceManager constructs gceManager object.
func CreateGceManager(configReader io.Reader, mode GcpCloudProviderMode, clusterName string, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, regional bool) (GceManager, error) {
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
		discoveredProjectId, discoveredLocation, err := getProjectAndLocation(regional)
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
	client.Timeout = httpTimeout
	gceService, err := NewAutoscalingGceClientV1(client, projectId)
	if err != nil {
		return nil, err
	}
	manager := &gceManagerImpl{
		migs:        make([]*migInformation, 0),
		GceService:  gceService,
		migCache:    make(map[GceRef]*Mig),
		location:    location,
		regional:    regional,
		projectId:   projectId,
		clusterName: clusterName,
		mode:        mode,
		templates: &templateBuilder{
			projectId: projectId,
		},
		interrupt:            make(chan struct{}),
		explicitlyConfigured: make(map[GceRef]bool),
		machinesCache:        make(map[machineTypeKey]*gce.MachineType),
	}

	switch mode {
	case ModeGCE:
		var err error
		if err = manager.fetchExplicitMigs(discoveryOpts.NodeGroupSpecs); err != nil {
			return nil, fmt.Errorf("failed to fetch MIGs: %v", err)
		}
		if manager.migAutoDiscoverySpecs, err = discoveryOpts.ParseMIGAutoDiscoverySpecs(); err != nil {
			return nil, err
		}
	case ModeGKE:
		gkeService, err := gke.New(client)
		if err != nil {
			return nil, err
		}
		if *gkeAPIEndpoint != "" {
			gkeService.BasePath = *gkeAPIEndpoint
		}
		manager.gkeService = gkeService
		if manager.regional {
			gkeBetaService, err := gke_beta.New(client)
			if err != nil {
				return nil, err
			}
			if *gkeAPIEndpoint != "" {
				gkeBetaService.BasePath = *gkeAPIEndpoint
			}
			manager.gkeBetaService = gkeBetaService
		}
	case ModeGKENAP:
		gkeBetaService, err := gke_beta.New(client)
		if err != nil {
			return nil, err
		}
		if *gkeAPIEndpoint != "" {
			gkeBetaService.BasePath = *gkeAPIEndpoint
		}
		manager.gkeBetaService = gkeBetaService
		glog.V(1).Info("Using GKE-NAP mode")
	}

	if err := manager.forceRefresh(); err != nil {
		return nil, err
	}

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

func (m *gceManagerImpl) assertGCE() {
	if m.mode != ModeGCE {
		glog.Fatalf("This should run only in GCE mode")
	}
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
	if m.regional {
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
		for _, igurl := range nodePool.InstanceGroupUrls {
			project, zone, name, err := ParseIgmUrl(igurl)
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
		for _, igurl := range nodePool.InstanceGroupUrls {
			project, zone, name, err := ParseIgmUrl(igurl)
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

	nodePoolsResponse, err := m.gkeBetaService.Projects.Zones.Clusters.NodePools.List(m.projectId, m.location, m.clusterName).Do()
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
		for _, igurl := range nodePool.InstanceGroupUrls {
			project, zone, name, err := ParseIgmUrl(igurl)
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

// RegisterMig registers mig in GceManager. Returns true if the node group didn't exist before or its config has changed.
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

	template, err := m.GceService.FetchMigTemplate(mig.GceRef)
	if err != nil {
		glog.Errorf("Failed to fetch template for %s", mig.Name)
	} else {
		cpu, mem, err := m.getCpuAndMemoryForMachineType(template.Properties.MachineType, mig.GceRef.Zone)
		if err != nil {
			glog.Errorf("Failed to get cpu and memory for machine type: %v.", err)
			return false
		}
		_, err = m.templates.buildNodeFromTemplate(mig, template, cpu, mem)
		if err != nil {
			glog.Errorf("Failed to build template for %s", mig.Name)
		}
	}
	return true
}

// UnregisterMig unregisters mig in GceManager. Returns true if the node group has been removed.
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
	deleteOp, err := m.gkeBetaService.Projects.Zones.Clusters.NodePools.Delete(m.projectId, m.location, m.clusterName,
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

	// TODO: handle preemptible
	// TODO: handle SSDs

	accelerators := []*gke_beta.AcceleratorConfig{}

	if gpuRequest, found := mig.spec.extraResources[gpu.ResourceNvidiaGPU]; found {
		gpuType, found := mig.spec.labels[gpu.GPULabel]
		if !found {
			return fmt.Errorf("failed to create node pool %v with gpu request of unspecified type", mig.nodePoolName)
		}
		gpuConfig := &gke_beta.AcceleratorConfig{
			AcceleratorType:  gpuType,
			AcceleratorCount: gpuRequest.Value(),
		}
		accelerators = append(accelerators, gpuConfig)

	}

	taints := []*gke_beta.NodeTaint{}
	for _, taint := range mig.spec.taints {
		if taint.Key == gpu.ResourceNvidiaGPU {
			continue
		}
		effect, found := taintEffectsMap[taint.Effect]
		if !found {
			effect = "EFFECT_UNSPECIFIED"
		}
		taint := &gke_beta.NodeTaint{
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

	config := gke_beta.NodeConfig{
		MachineType:  mig.spec.machineType,
		OauthScopes:  defaultOAuthScopes,
		Labels:       labels,
		Accelerators: accelerators,
		Taints:       taints,
	}

	autoscaling := gke_beta.NodePoolAutoscaling{
		Enabled:         true,
		MinNodeCount:    napMinNodes,
		MaxNodeCount:    napMaxNodes,
		Autoprovisioned: true,
	}

	createRequest := gke_beta.CreateNodePoolRequest{
		NodePool: &gke_beta.NodePool{
			Name:             mig.nodePoolName,
			InitialNodeCount: 0,
			Config:           &config,
			Autoscaling:      &autoscaling,
		},
	}

	createOp, err := m.gkeBetaService.Projects.Zones.Clusters.NodePools.Create(m.projectId, m.location, m.clusterName,
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

func (m *gceManagerImpl) fetchMachinesCache() error {
	if m.machinesCacheLastRefresh.Add(machinesRefreshInterval).After(time.Now()) {
		return nil
	}
	var locations []string
	if m.mode == ModeGKENAP {
		cluster, err := m.gkeBetaService.Projects.Locations.Clusters.Get(fmt.Sprintf("projects/%s/locations/%s/clusters/%s", m.projectId, m.location, m.clusterName)).Do()
		if err != nil {
			return err
		}
		locations = cluster.Locations
	} else {
		if m.regional {
			cluster, err := m.gkeBetaService.Projects.Locations.Clusters.Get(fmt.Sprintf("projects/%s/locations/%s/clusters/%s", m.projectId, m.location, m.clusterName)).Do()
			if err != nil {
				return err
			}
			locations = cluster.Locations
		} else {
			cluster, err := m.gkeService.Projects.Zones.Clusters.Get(m.projectId, m.location, m.clusterName).Do()
			if err != nil {
				return err
			}
			locations = cluster.Locations
		}
	}
	machinesCache := make(map[machineTypeKey]*gce.MachineType)
	for _, location := range locations {
		machineTypes, err := m.GceService.FetchMachineTypes(location)
		if err != nil {
			return err
		}
		for _, machineType := range machineTypes {
			machinesCache[machineTypeKey{location, machineType.Name}] = machineType
		}

	}
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.machinesCache = machinesCache
	m.machinesCacheLastRefresh = time.Now()
	glog.V(2).Infof("Refreshed machine types, next refresh after %v", m.machinesCacheLastRefresh.Add(machinesRefreshInterval))
	return nil
}

// End of v1alpha1/v1beta1 mess

// GetMigSize gets MIG size.
func (m *gceManagerImpl) GetMigSize(mig *Mig) (int64, error) {
	targetSize, err := m.GceService.FetchMigTargetSize(mig.GceRef)
	if err != nil {
		return -1, err
	}
	return targetSize, nil
}

// SetMigSize sets MIG size.
func (m *gceManagerImpl) SetMigSize(mig *Mig, size int64) error {
	glog.V(0).Infof("Setting mig size %s to %d", mig.Id(), size)
	return m.GceService.ResizeMig(mig.GceRef, size)
}

//  GKE
func (m *gceManagerImpl) waitForGkeOp(operation *gke_beta.Operation) error {
	for start := time.Now(); time.Since(start) < gkeOperationWaitTimeout; time.Sleep(defaultOperationPollInterval) {
		glog.V(4).Infof("Waiting for operation %s %s %s", m.projectId, m.location, operation.Name)
		if op, err := m.gkeBetaService.Projects.Zones.Operations.Get(m.projectId, m.location, operation.Name).Do(); err == nil {
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
			return fmt.Errorf("Cannot delete instances which don't belong to the same MIG.")
		}
	}

	return m.GceService.DeleteInstances(commonMig.GceRef, instances)
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

		basename, err := m.GceService.FetchMigBasename(mig.GceRef)
		if err != nil {
			return err
		}
		m.updateMigBasename(mig.GceRef, basename)

		instances, err := m.GceService.FetchMigInstances(mig.GceRef)
		if err != nil {
			glog.V(4).Infof("Failed MIG info request for %s %s %s: %v", mig.Project, mig.Zone, mig.Name, err)
			return err
		}
		for _, ref := range instances {
			newMigCache[ref] = mig
		}
	}

	m.migCache = newMigCache
	return nil
}

// GetMigNodes returns mig nodes.
func (m *gceManagerImpl) GetMigNodes(mig *Mig) ([]string, error) {
	instances, err := m.GceService.FetchMigInstances(mig.GceRef)
	if err != nil {
		return []string{}, err
	}
	result := make([]string, 0)
	for _, ref := range instances {
		result = append(result, fmt.Sprintf("gce://%s/%s/%s", ref.Project, ref.Zone, ref.Name))
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

func (m *gceManagerImpl) Refresh() error {
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.forceRefresh()
}

func (m *gceManagerImpl) forceRefresh() error {
	switch m.mode {
	case ModeGCE:
		if err := m.clearMachinesCache(); err != nil {
			glog.Errorf("Failed to clear machine types cache: %v", err)
			return err
		}
		if err := m.fetchAutoMigs(); err != nil {
			glog.Errorf("Failed to fetch MIGs: %v", err)
			return err
		}
	case ModeGKENAP:
		if err := m.fetchResourceLimiter(); err != nil {
			glog.Errorf("Failed to fetch resource limits: %v", err)
			return err
		}
		fallthrough
	case ModeGKE:
		if err := m.fetchMachinesCache(); err != nil {
			glog.Errorf("Failed to fetch machine types: %v", err)
			return err
		}
		if err := m.fetchAllNodePools(); err != nil {
			glog.Errorf("Failed to fetch node pools: %v", err)
			return err
		}
	}
	m.lastRefresh = time.Now()
	glog.V(2).Infof("Refreshed GCE resources, next refresh after %v", m.lastRefresh.Add(refreshInterval))
	return nil
}

// Fetch explicitly configured MIGs. These MIGs should never be unregistered
// during refreshes, even if they no longer exist in GCE.
func (m *gceManagerImpl) fetchExplicitMigs(specs []string) error {
	m.assertGCE()

	changed := false
	for _, spec := range specs {
		mig, err := m.buildMigFromSpec(spec)
		if err != nil {
			return err
		}
		if m.RegisterMig(mig) {
			changed = true
		}
		m.explicitlyConfigured[mig.GceRef] = true
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

func (m *gceManagerImpl) buildMigFromSpec(spec string) (*Mig, error) {
	s, err := dynamic.SpecFromString(spec, scaleToZeroSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	mig := &Mig{gceManager: m, minSize: s.MinSize, maxSize: s.MaxSize, exist: true}
	if mig.Project, mig.Zone, mig.Name, err = ParseMigUrl(s.Name); err != nil {
		return nil, fmt.Errorf("failed to parse mig url: %s got error: %v", s.Name, err)
	}
	return mig, nil
}

// Fetch automatically discovered MIGs. These MIGs should be unregistered if
// they no longer exist in GCE.
func (m *gceManagerImpl) fetchAutoMigs() error {
	m.assertGCE()

	exists := make(map[GceRef]bool)
	changed := false
	for _, cfg := range m.migAutoDiscoverySpecs {
		links, err := m.findMigsNamed(cfg.Re)
		if err != nil {
			return fmt.Errorf("cannot autodiscover managed instance groups: %s", err)
		}
		for _, link := range links {
			mig, err := m.buildMigFromAutoCfg(link, cfg)
			if err != nil {
				return err
			}
			exists[mig.GceRef] = true
			if m.explicitlyConfigured[mig.GceRef] {
				// This MIG was explicitly configured, but would also be
				// autodiscovered. We want the explicitly configured min and max
				// nodes to take precedence.
				glog.V(3).Infof("Ignoring explicitly configured MIG %s for autodiscovery.", mig.GceRef.Name)
				continue
			}
			if m.RegisterMig(mig) {
				glog.V(3).Infof("Autodiscovered MIG %s using regexp %s", mig.GceRef.Name, cfg.Re.String())
				changed = true
			}
		}
	}

	for _, mig := range m.getMigs() {
		if !exists[mig.config.GceRef] && !m.explicitlyConfigured[mig.config.GceRef] {
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

func (m *gceManagerImpl) buildMigFromAutoCfg(link string, cfg cloudprovider.MIGAutoDiscoveryConfig) (*Mig, error) {
	spec := dynamic.NodeGroupSpec{
		Name:               link,
		MinSize:            cfg.MinSize,
		MaxSize:            cfg.MaxSize,
		SupportScaleToZero: scaleToZeroSupported,
	}
	if verr := spec.Validate(); verr != nil {
		return nil, fmt.Errorf("failed to create node group spec: %v", verr)
	}
	mig := &Mig{gceManager: m, minSize: spec.MinSize, maxSize: spec.MaxSize, exist: true}
	var err error
	if mig.Project, mig.Zone, mig.Name, err = ParseMigUrl(spec.Name); err != nil {
		return nil, fmt.Errorf("failed to parse mig url: %s got error: %v", spec.Name, err)
	}
	return mig, nil
}

func (m *gceManagerImpl) fetchResourceLimiter() error {
	if m.mode == ModeGKENAP {
		cluster, err := m.gkeBetaService.Projects.Zones.Clusters.Get(m.projectId, m.location, m.clusterName).Do()
		if err != nil {
			return err
		}
		if cluster.Autoscaling == nil {
			return nil
		}

		minLimits := make(map[string]int64)
		maxLimits := make(map[string]int64)
		for _, limit := range cluster.Autoscaling.ResourceLimits {
			if _, found := supportedResources[limit.ResourceType]; !found {
				glog.Warningf("Unsupported limit defined %s: %d - %d", limit.ResourceType, limit.Minimum, limit.Maximum)
			}
			minLimits[limit.ResourceType] = limit.Minimum
			maxLimits[limit.ResourceType] = limit.Maximum
		}

		// GKE API provides memory in GB, but ResourceLimiter expects them in bytes
		if _, found := minLimits[cloudprovider.ResourceNameMemory]; found {
			minLimits[cloudprovider.ResourceNameMemory] = minLimits[cloudprovider.ResourceNameMemory] * units.Gigabyte
		}
		if _, found := maxLimits[cloudprovider.ResourceNameMemory]; found {
			maxLimits[cloudprovider.ResourceNameMemory] = maxLimits[cloudprovider.ResourceNameMemory] * units.Gigabyte
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

func (m *gceManagerImpl) clearMachinesCache() error {
	if m.machinesCacheLastRefresh.Add(machinesRefreshInterval).After(time.Now()) {
		return nil
	}
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.machinesCache = make(map[machineTypeKey]*gce.MachineType)
	m.machinesCacheLastRefresh = time.Now()
	glog.V(2).Infof("Cleared machine types cache, next clear after %v", m.machinesCacheLastRefresh.Add(machinesRefreshInterval))
	return nil
}

// Code borrowed from gce cloud provider. Reuse the original as soon as it becomes public.
func getProjectAndLocation(regional bool) (string, string, error) {
	result, err := metadata.Get("instance/zone")
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(result, "/")
	if len(parts) != 4 {
		return "", "", fmt.Errorf("unexpected response: %s", result)
	}
	location := parts[3]
	if regional {
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

func (m *gceManagerImpl) findMigsNamed(name *regexp.Regexp) ([]string, error) {
	if m.regional {
		return m.findMigsInRegion(m.location, name)
	}
	return m.GceService.FetchMigsWithName(m.location, name)
}

func (m *gceManagerImpl) getZones(region string) ([]string, error) {
	zones, err := m.GceService.FetchZones(region)
	if err != nil {
		return nil, fmt.Errorf("cannot get zones for GCE region %s: %v", region, err)
	}
	return zones, nil
}

func (m *gceManagerImpl) findMigsInRegion(region string, name *regexp.Regexp) ([]string, error) {
	links := make([]string, 0)
	zones, err := m.getZones(region)
	if err != nil {
		return nil, err
	}
	for _, z := range zones {
		zl, err := m.GceService.FetchMigsWithName(z, name)
		if err != nil {
			return nil, err
		}
		for _, link := range zl {
			links = append(links, link)
		}
	}

	return links, nil
}

func (m *gceManagerImpl) getMigTemplateNode(mig *Mig) (*apiv1.Node, error) {
	if mig.Exist() {
		template, err := m.GceService.FetchMigTemplate(mig.GceRef)
		if err != nil {
			return nil, err
		}
		cpu, mem, err := m.getCpuAndMemoryForMachineType(template.Properties.MachineType, mig.GceRef.Zone)
		if err != nil {
			return nil, err
		}
		return m.templates.buildNodeFromTemplate(mig, template, cpu, mem)
	} else if mig.Autoprovisioned() {
		cpu, mem, err := m.getCpuAndMemoryForMachineType(mig.spec.machineType, mig.GceRef.Zone)
		if err != nil {
			return nil, err
		}
		return m.templates.buildNodeFromAutoprovisioningSpec(mig, cpu, mem)
	}
	return nil, fmt.Errorf("unable to get node info for %s/%s/%s", mig.Project, mig.Zone, mig.Name)
}

func (m *gceManagerImpl) getMachineFromCache(machineType string, zone string) *gce.MachineType {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	return m.machinesCache[machineTypeKey{zone, machineType}]
}

func (m *gceManagerImpl) addMachineToCache(machineType string, zone string, machine *gce.MachineType) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.machinesCache[machineTypeKey{zone, machineType}] = machine
}

func (m *gceManagerImpl) getCpuAndMemoryForMachineType(machineType string, zone string) (cpu int64, mem int64, err error) {
	if strings.HasPrefix(machineType, "custom-") {
		return parseCustomMachineType(machineType)
	}
	machine := m.getMachineFromCache(machineType, zone)
	if machine == nil {
		machine, err = m.GceService.FetchMachineType(zone, machineType)
		if err != nil {
			return 0, 0, err
		}
		m.addMachineToCache(machineType, zone, machine)
	}
	return machine.GuestCpus, machine.MemoryMb * bytesPerMB, nil
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
	mem = mem * bytesPerMB
	return
}
