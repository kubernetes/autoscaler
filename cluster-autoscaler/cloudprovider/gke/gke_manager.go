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

package gke

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	provider_gce "k8s.io/kubernetes/pkg/cloudprovider/providers/gce"

	"cloud.google.com/go/compute/metadata"
	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gce_api "google.golang.org/api/compute/v1"
	gcfg "gopkg.in/gcfg.v1"
)

// GcpCloudProviderMode allows to pass information whether the cluster is in NAP mode.
type GcpCloudProviderMode string

const (
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

// GkeManager handles gce communication and data caching.
type GkeManager interface {
	// GetMigSize gets MIG size.
	GetMigSize(mig gce.Mig) (int64, error)
	// SetMigSize sets MIG size.
	SetMigSize(mig gce.Mig, size int64) error
	// DeleteInstances deletes the given instances. All instances must be controlled by the same MIG.
	DeleteInstances(instances []*gce.GceRef) error
	// GetMigForInstance returns MigConfig of the given Instance
	GetMigForInstance(instance *gce.GceRef) (gce.Mig, error)
	// GetMigNodes returns mig nodes.
	GetMigNodes(mig gce.Mig) ([]string, error)
	// Refresh updates config by calling GKE API (in GKE mode only).
	Refresh() error
	// GetResourceLimiter returns resource limiter.
	GetResourceLimiter() (*cloudprovider.ResourceLimiter, error)
	// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
	Cleanup() error
	GetMigs() []*gce.MigInformation
	CreateNodePool(mig *GkeMig) (*GkeMig, error)
	DeleteNodePool(toBeRemoved *GkeMig) error
	GetLocation() string
	GetProjectId() string
	GetClusterName() string
	GetMigTemplateNode(mig *GkeMig) (*apiv1.Node, error)
}

// gkeManagerImpl handles gce communication and data caching.
type gkeManagerImpl struct {
	cache                    gce.GceCache
	lastRefresh              time.Time
	machinesCacheLastRefresh time.Time

	GkeService AutoscalingGkeClient
	GceService gce.AutoscalingGceClient

	location    string
	projectId   string
	clusterName string
	mode        GcpCloudProviderMode
	templates   *GkeTemplateBuilder
	interrupt   chan struct{}
	regional    bool
}

// CreateGkeManager constructs gkeManager object.
func CreateGkeManager(configReader io.Reader, mode GcpCloudProviderMode, clusterName string, regional bool) (GkeManager, error) {
	// Create Google Compute Engine token.
	var err error
	tokenSource := google.ComputeTokenSource("")
	if len(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")) > 0 {
		tokenSource, err = google.DefaultTokenSource(oauth2.NoContext, gce_api.ComputeScope)
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
	gceService, err := gce.NewAutoscalingGceClientV1(client, projectId)
	if err != nil {
		return nil, err
	}
	manager := &gkeManagerImpl{
		cache:       gce.NewGceCache(gceService),
		GceService:  gceService,
		location:    location,
		regional:    regional,
		projectId:   projectId,
		clusterName: clusterName,
		mode:        mode,
		templates:   &GkeTemplateBuilder{},
		interrupt:   make(chan struct{}),
	}

	switch mode {
	case ModeGKE:
		gkeService, err := NewAutoscalingGkeClientV1(client, projectId, location, clusterName)
		if err != nil {
			return nil, err
		}
		manager.GkeService = gkeService
	case ModeGKENAP:
		gkeBetaService, err := NewAutoscalingGkeClientV1beta1(client, projectId, location, clusterName)
		if err != nil {
			return nil, err
		}
		manager.GkeService = gkeBetaService
		glog.V(1).Info("Using GKE-NAP mode")
	}

	if err := manager.forceRefresh(); err != nil {
		return nil, err
	}

	go wait.Until(func() {
		if err := manager.cache.RegenerateInstancesCache(); err != nil {
			glog.Errorf("Error while regenerating Mig cache: %v", err)
		}
	}, time.Hour, manager.interrupt)

	return manager, nil
}

// Cleanup closes the channel to signal the go routine to stop that is handling the cache
func (m *gkeManagerImpl) Cleanup() error {
	close(m.interrupt)
	return nil
}

func (m *gkeManagerImpl) assertGKE() {
	if m.mode != ModeGKE && m.mode != ModeGKENAP {
		glog.Fatalf("This should run only in GKE mode")
	}
}

func (m *gkeManagerImpl) assertGKENAP() {
	if m.mode != ModeGKENAP {
		glog.Fatalf("This should run only in GKE mode with autoprovisioning enabled")
	}
}

func (m *gkeManagerImpl) refreshNodePools() error {
	m.assertGKE()

	nodePools, err := m.GkeService.FetchNodePools()
	if err != nil {
		return err
	}

	existingMigs := map[gce.GceRef]struct{}{}
	changed := false

	for _, nodePool := range nodePools {
		for _, igurl := range nodePool.InstanceGroupUrls {
			project, zone, name, err := gce.ParseIgmUrl(igurl)
			if err != nil {
				return err
			}
			mig := &GkeMig{
				gceRef: gce.GceRef{
					Name:    name,
					Zone:    zone,
					Project: project,
				},
				gkeManager:      m,
				exist:           true,
				autoprovisioned: nodePool.Autoprovisioned,
				nodePoolName:    nodePool.Name,
				minSize:         int(nodePool.MinNodeCount),
				maxSize:         int(nodePool.MaxNodeCount),
			}
			existingMigs[mig.GceRef()] = struct{}{}

			if m.registerMig(mig) {
				changed = true
			}
		}
	}
	for _, mig := range m.cache.GetMigs() {
		if _, found := existingMigs[mig.Config.GceRef()]; !found {
			m.cache.UnregisterMig(mig.Config)
			changed = true
		}
	}
	if changed {
		return m.cache.RegenerateInstancesCache()
	}
	return nil
}

// RegisterMig registers mig in GceManager. Returns true if the node group didn't exist before or its config has changed.
func (m *gkeManagerImpl) registerMig(mig *GkeMig) bool {
	changed := m.cache.RegisterMig(mig)
	if changed {
		// Try to build a node from template to validate that this group
		// can be scaled up from 0 nodes.
		// We may never need to do it, so just log error if it fails.
		if _, err := m.GetMigTemplateNode(mig); err != nil {
			glog.Errorf("Can't build node from template for %s, won't be able to scale from 0: %v", mig.GceRef().String(), err)
		}
	}
	return changed
}

func (m *gkeManagerImpl) DeleteNodePool(toBeRemoved *GkeMig) error {
	m.assertGKENAP()
	if !toBeRemoved.Autoprovisioned() {
		return fmt.Errorf("only autoprovisioned node pools can be deleted")
	}
	// TODO: handle multi-zonal node pools.
	err := m.GkeService.DeleteNodePool(toBeRemoved.NodePoolName())
	if err != nil {
		return err
	}
	return m.refreshNodePools()
}

func (m *gkeManagerImpl) CreateNodePool(mig *GkeMig) (*GkeMig, error) {
	m.assertGKENAP()

	err := m.GkeService.CreateNodePool(mig)
	if err != nil {
		return nil, err
	}
	err = m.refreshNodePools()
	if err != nil {
		return nil, err
	}
	// TODO(aleksandra-malinowska): support multi-zonal node pools.
	for _, existingMig := range m.cache.GetMigs() {
		gkeMig, ok := existingMig.Config.(*GkeMig)
		if !ok {
			// This is "should never happen" branch.
			// Report error as InternalError since it would signify a
			// serious bug in autoscaler code.
			errMsg := fmt.Sprintf("Mig %s is not GkeMig: got %v, want GkeMig", existingMig.Config.GceRef().String(), reflect.TypeOf(existingMig.Config))
			glog.Error(errMsg)
			return nil, errors.NewAutoscalerError(errors.InternalError, errMsg)
		}
		if gkeMig.NodePoolName() == mig.NodePoolName() {
			return gkeMig, nil
		}
	}
	return nil, fmt.Errorf("node pool %s not found", mig.NodePoolName())
}

func (m *gkeManagerImpl) fetchMachinesCache() error {
	if m.machinesCacheLastRefresh.Add(machinesRefreshInterval).After(time.Now()) {
		return nil
	}
	var locations []string
	locations, err := m.GkeService.FetchLocations()
	if err != nil {
		return err
	}
	machinesCache := make(map[gce.MachineTypeKey]*gce_api.MachineType)
	for _, location := range locations {
		machineTypes, err := m.GceService.FetchMachineTypes(location)
		if err != nil {
			return err
		}
		for _, machineType := range machineTypes {
			machinesCache[gce.MachineTypeKey{location, machineType.Name}] = machineType
		}

	}
	m.cache.SetMachinesCache(machinesCache)
	nextRefresh := time.Now()
	m.machinesCacheLastRefresh = nextRefresh
	glog.V(2).Infof("Refreshed machine types, next refresh after %v", nextRefresh)
	return nil
}

// GetMigSize gets MIG size.
func (m *gkeManagerImpl) GetMigSize(mig gce.Mig) (int64, error) {
	targetSize, err := m.GceService.FetchMigTargetSize(mig.GceRef())
	if err != nil {
		return -1, err
	}
	return targetSize, nil
}

// SetMigSize sets MIG size.
func (m *gkeManagerImpl) SetMigSize(mig gce.Mig, size int64) error {
	glog.V(0).Infof("Setting mig size %s to %d", mig.Id(), size)
	return m.GceService.ResizeMig(mig.GceRef(), size)
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same MIG.
func (m *gkeManagerImpl) DeleteInstances(instances []*gce.GceRef) error {
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

	return m.GceService.DeleteInstances(commonMig.GceRef(), instances)
}

func (m *gkeManagerImpl) GetMigs() []*gce.MigInformation {
	return m.cache.GetMigs()
}

// GetMigForInstance returns MigConfig of the given Instance
func (m *gkeManagerImpl) GetMigForInstance(instance *gce.GceRef) (gce.Mig, error) {
	return m.cache.GetMigForInstance(instance)
}

// GetMigNodes returns mig nodes.
func (m *gkeManagerImpl) GetMigNodes(mig gce.Mig) ([]string, error) {
	instances, err := m.GceService.FetchMigInstances(mig.GceRef())
	if err != nil {
		return []string{}, err
	}
	result := make([]string, 0)
	for _, ref := range instances {
		result = append(result, fmt.Sprintf("gce://%s/%s/%s", ref.Project, ref.Zone, ref.Name))
	}
	return result, nil
}

func (m *gkeManagerImpl) GetLocation() string {
	return m.location
}

func (m *gkeManagerImpl) GetProjectId() string {
	return m.projectId
}

func (m *gkeManagerImpl) GetClusterName() string {
	return m.clusterName
}

func (m *gkeManagerImpl) Refresh() error {
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.forceRefresh()
}

func (m *gkeManagerImpl) forceRefresh() error {
	switch m.mode {
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
		if err := m.refreshNodePools(); err != nil {
			glog.Errorf("Failed to fetch node pools: %v", err)
			return err
		}
	}
	m.lastRefresh = time.Now()
	glog.V(2).Infof("Refreshed GCE resources, next refresh after %v", m.lastRefresh.Add(refreshInterval))
	return nil
}

func (m *gkeManagerImpl) buildMigFromFlag(flag string) (gce.Mig, error) {
	s, err := dynamic.SpecFromString(flag, scaleToZeroSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	return m.buildMigFromSpec(s)
}

func (m *gkeManagerImpl) buildMigFromSpec(s *dynamic.NodeGroupSpec) (gce.Mig, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid node group spec: %v", err)
	}
	project, zone, name, err := gce.ParseMigUrl(s.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mig url: %s got error: %v", s.Name, err)
	}
	mig := &GkeMig{
		gceRef: gce.GceRef{
			Project: project,
			Name:    name,
			Zone:    zone,
		},
		gkeManager: m,
		minSize:    s.MinSize,
		maxSize:    s.MaxSize,
		exist:      true,
	}
	return mig, nil
}

func (m *gkeManagerImpl) fetchResourceLimiter() error {
	if m.mode == ModeGKENAP {
		resourceLimiter, err := m.GkeService.FetchResourceLimits()
		if err != nil {
			return err
		}
		if resourceLimiter != nil {
			glog.V(2).Infof("Refreshed resource limits: %s", resourceLimiter.String())
			m.cache.SetResourceLimiter(resourceLimiter)
		} else {
			oldLimits, _ := m.cache.GetResourceLimiter()
			glog.Errorf("Resource limits should always be defined in NAP mode, but they appear to be empty. Using possibly outdated limits: %v", oldLimits.String())
		}
	}
	return nil
}

// GetResourceLimiter returns resource limiter.
func (m *gkeManagerImpl) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return m.cache.GetResourceLimiter()
}

func (m *gkeManagerImpl) clearMachinesCache() {
	if m.machinesCacheLastRefresh.Add(machinesRefreshInterval).After(time.Now()) {
		return
	}

	machinesCache := make(map[gce.MachineTypeKey]*gce_api.MachineType)
	m.cache.SetMachinesCache(machinesCache)
	nextRefresh := time.Now()
	m.machinesCacheLastRefresh = nextRefresh
	glog.V(2).Infof("Cleared machine types cache, next clear after %v", nextRefresh)
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

func (m *gkeManagerImpl) GetMigTemplateNode(mig *GkeMig) (*apiv1.Node, error) {
	if mig.Exist() {
		template, err := m.GceService.FetchMigTemplate(mig.GceRef())
		if err != nil {
			return nil, err
		}
		cpu, mem, err := m.getCpuAndMemoryForMachineType(template.Properties.MachineType, mig.GceRef().Zone)
		if err != nil {
			return nil, err
		}
		return m.templates.BuildNodeFromTemplate(mig, template, cpu, mem)
	} else if mig.Autoprovisioned() {
		cpu, mem, err := m.getCpuAndMemoryForMachineType(mig.Spec().MachineType, mig.GceRef().Zone)
		if err != nil {
			return nil, err
		}
		return m.templates.BuildNodeFromMigSpec(mig, cpu, mem)
	}
	return nil, fmt.Errorf("unable to get node info for %s", mig.GceRef().String())
}

func (m *gkeManagerImpl) getCpuAndMemoryForMachineType(machineType string, zone string) (cpu int64, mem int64, err error) {
	if strings.HasPrefix(machineType, "custom-") {
		return parseCustomMachineType(machineType)
	}
	machine := m.cache.GetMachineFromCache(machineType, zone)
	if machine == nil {
		machine, err = m.GceService.FetchMachineType(zone, machineType)
		if err != nil {
			return 0, 0, err
		}
		m.cache.AddMachineToCache(machineType, zone, machine)
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
