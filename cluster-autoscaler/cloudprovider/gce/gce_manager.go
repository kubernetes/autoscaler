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
	"os"
	"regexp"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	provider_gce "k8s.io/kubernetes/pkg/cloudprovider/providers/gce"

	"cloud.google.com/go/compute/metadata"
	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gce "google.golang.org/api/compute/v1"
	gcfg "gopkg.in/gcfg.v1"
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

// GceManager handles gce communication and data caching.
type GceManager interface {
	// GetMigSize gets MIG size.
	GetMigSize(mig Mig) (int64, error)
	// SetMigSize sets MIG size.
	SetMigSize(mig Mig, size int64) error
	// DeleteInstances deletes the given instances. All instances must be controlled by the same MIG.
	DeleteInstances(instances []*GceRef) error
	// GetMigForInstance returns MigConfig of the given Instance
	GetMigForInstance(instance *GceRef) (Mig, error)
	// GetMigNodes returns mig nodes.
	GetMigNodes(mig Mig) ([]string, error)
	// Refresh updates config by calling GKE API (in GKE mode only).
	Refresh() error
	// GetResourceLimiter returns resource limiter.
	GetResourceLimiter() (*cloudprovider.ResourceLimiter, error)
	// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
	Cleanup() error
	getMigs() []*MigInformation
	createNodePool(mig Mig) (Mig, error)
	deleteNodePool(toBeRemoved Mig) error
	getLocation() string
	getProjectId() string
	getClusterName() string
	getMode() GcpCloudProviderMode
	findMigsNamed(name *regexp.Regexp) ([]string, error)
	getMigTemplateNode(mig Mig) (*apiv1.Node, error)
	getCpuAndMemoryForMachineType(machineType string, zone string) (cpu int64, mem int64, err error)
}

// gceManagerImpl handles gce communication and data caching.
type gceManagerImpl struct {
	cache       GceCache
	lastRefresh time.Time

	GkeService AutoscalingGkeClient
	GceService AutoscalingGceClient

	location              string
	projectId             string
	clusterName           string
	mode                  GcpCloudProviderMode
	templates             *templateBuilder
	interrupt             chan struct{}
	regional              bool
	explicitlyConfigured  map[GceRef]bool
	migAutoDiscoverySpecs []cloudprovider.MIGAutoDiscoveryConfig
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
		cache:       NewGceCache(gceService),
		GceService:  gceService,
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
		if err := manager.cache.RegenerateCacheWithLock(); err != nil {
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
	if m.mode != ModeGKE && m.mode != ModeGKENAP {
		glog.Fatalf("This should run only in GKE mode")
	}
}

func (m *gceManagerImpl) assertGKENAP() {
	if m.mode != ModeGKENAP {
		glog.Fatalf("This should run only in GKE mode with autoprovisioning enabled")
	}
}

func (m *gceManagerImpl) refreshNodePools() error {
	m.assertGKE()

	nodePools, err := m.GkeService.FetchNodePools()
	if err != nil {
		return err
	}

	existingMigs := map[GceRef]struct{}{}
	changed := false

	for _, nodePool := range nodePools {
		for _, igurl := range nodePool.InstanceGroupUrls {
			project, zone, name, err := ParseIgmUrl(igurl)
			if err != nil {
				return err
			}
			mig := &gceMig{
				gceRef: GceRef{
					Name:    name,
					Zone:    zone,
					Project: project,
				},
				gceManager:      m,
				exist:           true,
				autoprovisioned: nodePool.Autoprovisioned,
				nodePoolName:    nodePool.Name,
				minSize:         int(nodePool.MinNodeCount),
				maxSize:         int(nodePool.MaxNodeCount),
			}
			existingMigs[mig.GceRef()] = struct{}{}

			if m.RegisterMig(mig) {
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
		return m.cache.RegenerateCacheWithLock()
	}
	return nil
}

// RegisterMig registers mig in GceManager. Returns true if the node group didn't exist before or its config has changed.
func (m *gceManagerImpl) RegisterMig(mig Mig) bool {
	changed := m.cache.RegisterMig(mig)
	if changed {
		// Try to build a node from template to validate that this group
		// can be scaled up from 0 nodes.
		// We may never need to do it, so just log error if it fails.
		if _, err := m.getMigTemplateNode(mig); err != nil {
			glog.Errorf("Can't build node from template for %s, won't be able to scale from 0: %v", mig.GceRef().String(), err)
		}
	}
	return changed
}

func (m *gceManagerImpl) deleteNodePool(toBeRemoved Mig) error {
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

func (m *gceManagerImpl) createNodePool(mig Mig) (Mig, error) {
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
		if existingMig.Config.NodePoolName() == mig.NodePoolName() {
			return existingMig.Config, nil
		}
	}
	return nil, fmt.Errorf("node pool %s not found", mig.NodePoolName())
}

func (m *gceManagerImpl) fetchMachinesCache() error {
	if m.cache.MachinesCacheFresh() {
		return nil
	}
	var locations []string
	locations, err := m.GkeService.FetchLocations()
	if err != nil {
		return err
	}
	machinesCache := make(map[MachineTypeKey]*gce.MachineType)
	for _, location := range locations {
		machineTypes, err := m.GceService.FetchMachineTypes(location)
		if err != nil {
			return err
		}
		for _, machineType := range machineTypes {
			machinesCache[MachineTypeKey{location, machineType.Name}] = machineType
		}

	}
	nextRefresh := m.cache.SetMachinesCache(machinesCache)
	glog.V(2).Infof("Refreshed machine types, next refresh after %v", nextRefresh)
	return nil
}

// GetMigSize gets MIG size.
func (m *gceManagerImpl) GetMigSize(mig Mig) (int64, error) {
	targetSize, err := m.GceService.FetchMigTargetSize(mig.GceRef())
	if err != nil {
		return -1, err
	}
	return targetSize, nil
}

// SetMigSize sets MIG size.
func (m *gceManagerImpl) SetMigSize(mig Mig, size int64) error {
	glog.V(0).Infof("Setting mig size %s to %d", mig.Id(), size)
	return m.GceService.ResizeMig(mig.GceRef(), size)
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

	return m.GceService.DeleteInstances(commonMig.GceRef(), instances)
}

func (m *gceManagerImpl) getMigs() []*MigInformation {
	return m.cache.GetMigs()
}

// GetMigForInstance returns MigConfig of the given Instance
func (m *gceManagerImpl) GetMigForInstance(instance *GceRef) (Mig, error) {
	return m.cache.GetMigForInstance(instance)
}

// GetMigNodes returns mig nodes.
func (m *gceManagerImpl) GetMigNodes(mig Mig) ([]string, error) {
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

func (m *gceManagerImpl) getLocation() string {
	return m.location
}
func (m *gceManagerImpl) getProjectId() string {
	return m.projectId
}
func (m *gceManagerImpl) getClusterName() string {
	return m.clusterName
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
		m.clearMachinesCache()
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
		if err := m.refreshNodePools(); err != nil {
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
		mig, err := m.buildMigFromFlag(spec)
		if err != nil {
			return err
		}
		if m.RegisterMig(mig) {
			changed = true
		}
		m.explicitlyConfigured[mig.GceRef()] = true
	}

	if changed {
		return m.cache.RegenerateCacheWithLock()
	}
	return nil
}

func (m *gceManagerImpl) buildMigFromFlag(flag string) (Mig, error) {
	s, err := dynamic.SpecFromString(flag, scaleToZeroSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	return m.buildMigFromSpec(s)
}

func (m *gceManagerImpl) buildMigFromAutoCfg(link string, cfg cloudprovider.MIGAutoDiscoveryConfig) (Mig, error) {
	s := &dynamic.NodeGroupSpec{
		Name:               link,
		MinSize:            cfg.MinSize,
		MaxSize:            cfg.MaxSize,
		SupportScaleToZero: scaleToZeroSupported,
	}
	return m.buildMigFromSpec(s)
}

func (m *gceManagerImpl) buildMigFromSpec(s *dynamic.NodeGroupSpec) (Mig, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid node group spec: %v", err)
	}
	project, zone, name, err := ParseMigUrl(s.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mig url: %s got error: %v", s.Name, err)
	}
	mig := &gceMig{
		gceRef: GceRef{
			Project: project,
			Name:    name,
			Zone:    zone,
		},
		gceManager: m,
		minSize:    s.MinSize,
		maxSize:    s.MaxSize,
		exist:      true,
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
			return fmt.Errorf("cannot autodiscover managed instance groups: %v", err)
		}
		for _, link := range links {
			mig, err := m.buildMigFromAutoCfg(link, cfg)
			if err != nil {
				return err
			}
			exists[mig.GceRef()] = true
			if m.explicitlyConfigured[mig.GceRef()] {
				// This MIG was explicitly configured, but would also be
				// autodiscovered. We want the explicitly configured min and max
				// nodes to take precedence.
				glog.V(3).Infof("Ignoring explicitly configured MIG %s in autodiscovery.", mig.GceRef().String())
				continue
			}
			if m.RegisterMig(mig) {
				glog.V(3).Infof("Autodiscovered MIG %s using regexp %s", mig.GceRef().String(), cfg.Re.String())
				changed = true
			}
		}
	}

	for _, mig := range m.getMigs() {
		if !exists[mig.Config.GceRef()] && !m.explicitlyConfigured[mig.Config.GceRef()] {
			m.cache.UnregisterMig(mig.Config)
			changed = true
		}
	}

	if changed {
		return m.cache.RegenerateCacheWithLock()
	}

	return nil
}

func (m *gceManagerImpl) fetchResourceLimiter() error {
	if m.mode == ModeGKENAP {
		resourceLimiter, err := m.GkeService.FetchResourceLimits()
		if err != nil {
			return err
		}

		glog.V(2).Infof("Refreshed resource limits: %s", resourceLimiter.String())
		m.cache.SetResourceLimiter(resourceLimiter)
	}
	return nil
}

// GetResourceLimiter returns resource limiter.
func (m *gceManagerImpl) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return m.cache.GetResourceLimiter()
}

func (m *gceManagerImpl) clearMachinesCache() {
	if m.cache.MachinesCacheFresh() {
		return
	}

	machinesCache := make(map[MachineTypeKey]*gce.MachineType)
	nextRefresh := m.cache.SetMachinesCache(machinesCache)
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

func (m *gceManagerImpl) getMigTemplateNode(mig Mig) (*apiv1.Node, error) {
	if mig.Exist() {
		template, err := m.GceService.FetchMigTemplate(mig.GceRef())
		if err != nil {
			return nil, err
		}
		cpu, mem, err := m.getCpuAndMemoryForMachineType(template.Properties.MachineType, mig.GceRef().Zone)
		if err != nil {
			return nil, err
		}
		return m.templates.buildNodeFromTemplate(mig, template, cpu, mem)
	} else if mig.Autoprovisioned() {
		cpu, mem, err := m.getCpuAndMemoryForMachineType(mig.Spec().machineType, mig.GceRef().Zone)
		if err != nil {
			return nil, err
		}
		return m.templates.buildNodeFromAutoprovisioningSpec(mig, cpu, mem)
	}
	return nil, fmt.Errorf("unable to get node info for %s", mig.GceRef().String())
}

func (m *gceManagerImpl) getCpuAndMemoryForMachineType(machineType string, zone string) (cpu int64, mem int64, err error) {
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
