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
	"k8s.io/apimachinery/pkg/util/wait"
	provider_gce "k8s.io/kubernetes/pkg/cloudprovider/providers/gce"
)

// GcpCloudProviderMode allows to pass information whether the cluster is GCE or GKE.
type GcpCloudProviderMode string

const (
	// ModeGCE means that the cluster is running on gce (or using the legacy gke setup).
	ModeGCE GcpCloudProviderMode = "gce"

	// ModeGKE means that the cluster is running
	ModeGKE GcpCloudProviderMode = "gke"
)

const (
	operationWaitTimeout       = 5 * time.Second
	operationPollInterval      = 100 * time.Millisecond
	nodeAutoprovisioningPrefix = "nodeautoprovisioning"
)

var (
	defaultOAuthScopes []string = []string{
		"https://www.googleapis.com/auth/compute",
		"https://www.googleapis.com/auth/devstorage.read_only",
		"https://www.googleapis.com/auth/service.management.readonly",
		"https://www.googleapis.com/auth/servicecontrol"}
)

type migInformation struct {
	config   *Mig
	basename string
}

// GceManager is handles gce communication and data caching.
type GceManager struct {
	migs     []*migInformation
	migCache map[GceRef]*Mig

	gceService *gce.Service
	gkeService *gke.Service

	cacheMutex sync.Mutex
	migsMutex  sync.Mutex

	zone        string
	projectId   string
	clusterName string
	mode        GcpCloudProviderMode
	templates   *templateBuilder
}

// CreateGceManager constructs gceManager object.
func CreateGceManager(configReader io.Reader, mode GcpCloudProviderMode, clusterName string) (*GceManager, error) {
	// Create Google Compute Engine token.
	var err error
	tokenSource := google.ComputeTokenSource("")
	if len(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")) > 0 {
		tokenSource, err = google.DefaultTokenSource(oauth2.NoContext, defaultOAuthScopes...)
		if err != nil {
			return nil, err
		}
	}
	var projectId, zone string
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
		zone = cfg.Global.LocalZone
	} else {
		glog.V(1).Infof("Using default TokenSource %#v", tokenSource)
	}
	if len(projectId) == 0 || len(zone) == 0 {
		projectId, zone, err = getProjectAndZone()
		if err != nil {
			return nil, err
		}
	}
	glog.V(1).Infof("GCE projectId=%s zone=%s", projectId, zone)

	// Create Google Compute Engine service.
	client := oauth2.NewClient(oauth2.NoContext, tokenSource)
	gceService, err := gce.New(client)
	if err != nil {
		return nil, err
	}
	manager := &GceManager{
		migs:        make([]*migInformation, 0),
		gceService:  gceService,
		migCache:    make(map[GceRef]*Mig),
		zone:        zone,
		projectId:   projectId,
		clusterName: clusterName,
		mode:        mode,
		templates: &templateBuilder{
			projectId: projectId,
			zone:      zone,
			service:   gceService,
		},
	}

	if mode == ModeGKE {
		gkeService, err := gke.New(client)
		if err != nil {
			return nil, err
		}
		manager.gkeService = gkeService
		err = manager.fetchAllNodePools()
		if err != nil {
			glog.Errorf("Failed to fetch node pools: %v", err)
			return nil, err
		}
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

func (m *GceManager) assertGKE() {
	if m.mode != ModeGKE {
		glog.Fatalf("This should run only in GKE mode")
	}
}

// Gets all registered node pools
func (m *GceManager) fetchAllNodePools() error {
	m.assertGKE()

	nodePoolsResponse, err := m.gkeService.Projects.Zones.Clusters.NodePools.List(m.projectId, m.zone, m.clusterName).Do()
	if err != nil {
		return err
	}

	existingMigs := map[GceRef]struct{}{}

	for _, nodePool := range nodePoolsResponse.NodePools {
		autoprovisioned := strings.Contains(nodePool.Name, nodeAutoprovisioningPrefix)
		autoscaled := nodePool.Autoscaling != nil && nodePool.Autoscaling.Enabled
		if !autoprovisioned && !autoscaled {
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
			}
			existingMigs[mig.GceRef] = struct{}{}

			if autoscaled {
				mig.minSize = int(nodePool.Autoscaling.MinNodeCount)
				mig.maxSize = int(nodePool.Autoscaling.MaxNodeCount)
			} else if autoprovisioned {
				mig.minSize = minAutoprovisionedSize
				mig.maxSize = maxAutoprovisionedSize
			}
			m.RegisterMig(mig)
		}
	}
	for _, mig := range m.getMigs() {
		if _, found := existingMigs[mig.config.GceRef]; !found {
			m.UnregisterMig(mig.config)
		}
	}
	return nil
}

// RegisterMig registers mig in Gce Manager. Returns true if the node group didn't exist before.
func (m *GceManager) RegisterMig(mig *Mig) bool {
	m.migsMutex.Lock()
	defer m.migsMutex.Unlock()

	updated := false
	for i := range m.migs {
		if m.migs[i].config.GceRef == mig.GceRef {
			m.migs[i].config = mig
			glog.V(8).Infof("Updated Mig %s/%s/%s", mig.GceRef.Project, mig.GceRef.Zone, mig.GceRef.Name)
			updated = true
		}
	}

	if !updated {
		glog.V(1).Infof("Registering %s/%s/%s", mig.GceRef.Project, mig.GceRef.Zone, mig.GceRef.Name)
		m.migs = append(m.migs, &migInformation{
			config: mig,
		})
	}

	template, err := m.templates.getMigTemplate(mig)
	if err != nil {
		glog.Errorf("Failed to build template for %s", mig.Name)
	} else {
		_, err = m.templates.buildNodeFromTemplate(mig, template)
		if err != nil {
			glog.Errorf("Failed to build template for %s", mig.Name)
		}
	}
	return !updated
}

// UnregisterMig unregisters mig in Gce Manager. Returns true if the node group has been removed.
func (m *GceManager) UnregisterMig(toBeRemoved *Mig) bool {
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

func (m *GceManager) deleteNodePool(toBeRemoved *Mig) error {
	m.assertGKE()
	if !toBeRemoved.Autoprovisioned() {
		return fmt.Errorf("only autoprovisioned node pools can be deleted")
	}
	// TODO: handle multi-zonal node pools.
	deleteOp, err := m.gkeService.Projects.Zones.Clusters.NodePools.Delete(m.projectId, m.zone, m.clusterName,
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

func (m *GceManager) createNodePool(mig *Mig) error {
	m.assertGKE()

	// TODO: handle preemptable
	// TODO: handle ssd
	// TODO: handle taints

	config := gke.NodeConfig{
		MachineType: mig.spec.machineType,
		OauthScopes: defaultOAuthScopes,
		Labels:      mig.spec.labels,
	}

	createRequest := gke.CreateNodePoolRequest{
		NodePool: &gke.NodePool{
			Name:             mig.nodePoolName,
			InitialNodeCount: 0,
			Config:           &config,
		},
	}

	createOp, err := m.gkeService.Projects.Zones.Clusters.NodePools.Create(m.projectId, m.zone, m.clusterName,
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

// GetMigSize gets MIG size.
func (m *GceManager) GetMigSize(mig *Mig) (int64, error) {
	if mig.unmanaged {
		return m.GetMigInstanceSize(mig)
	}
	igm, err := m.gceService.InstanceGroupManagers.Get(mig.Project, mig.Zone, mig.Name).Do()
	if err != nil {
		return -1, err
	}
	return igm.TargetSize, nil
}

// SetMigSize sets MIG size.
func (m *GceManager) SetMigSize(mig *Mig, size int64) error {
	glog.V(0).Infof("Setting mig size %s to %d", mig.Id(), size)
	if mig.unmanaged {
		return m.SetMigInstanceSize(mig, size)
	}
	op, err := m.gceService.InstanceGroupManagers.Resize(mig.Project, mig.Zone, mig.Name, size).Do()
	if err != nil {
		return err
	}
	return m.waitForOp(op, mig.Project, mig.Zone)
}

// GCE
func (m *GceManager) waitForOp(operation *gce.Operation, project string, zone string) error {
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
func (m *GceManager) waitForGkeOp(operation *gke.Operation) error {
	for start := time.Now(); time.Since(start) < operationWaitTimeout; time.Sleep(operationPollInterval) {
		glog.V(4).Infof("Waiting for operation %s %s %s", m.projectId, m.zone, operation.Name)
		if op, err := m.gkeService.Projects.Zones.Operations.Get(m.projectId, m.zone, operation.Name).Do(); err == nil {
			glog.V(4).Infof("Operation %s %s %s status: %s", m.projectId, m.zone, operation.Name, op.Status)
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

	op, err := m.gceService.InstanceGroupManagers.DeleteInstances(commonMig.Project, commonMig.Zone, commonMig.Name, &req).Do()
	if err != nil {
		return err
	}
	return m.waitForOp(op, commonMig.Project, commonMig.Zone)
}

func (m *GceManager) getMigs() []*migInformation {
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

func (m *GceManager) updateMigBasename(ref GceRef, basename string) {
	m.migsMutex.Lock()
	defer m.migsMutex.Unlock()
	for _, mig := range m.migs {
		if mig.config.GceRef == ref {
			mig.basename = basename
		}
	}
}

// GetMigForInstance returns MigConfig of the given Instance
func (m *GceManager) GetMigForInstance(instance *GceRef) (*Mig, error) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	if mig, found := m.migCache[*instance]; found {
		return mig, nil
	}

	refreshedUnmanaged := false
	for _, mig := range m.getMigs() {
		if mig.config.Project != instance.Project || mig.config.Zone != instance.Zone {
			continue
		}
		// refresh the cache exactly once in case this instance is new, otherwise ignore
		if mig.config.unmanaged {
			if !refreshedUnmanaged {
				refreshedUnmanaged = true
				if err := m.regenerateCache(); err != nil {
					return nil, fmt.Errorf("Error while looking for MIG for instance %+v, error: %v", *instance, err)
				}
				if mig, found := m.migCache[*instance]; found {
					return mig, nil
				}
			}
			// we cannot depend on basename
			continue
		}
		if strings.HasPrefix(instance.Name, mig.basename) {
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

	for _, migInfo := range m.getMigs() {
		mig := migInfo.config
		glog.V(4).Infof("Regenerating MIG information for %s %s %s", mig.Project, mig.Zone, mig.Name)

		if mig.unmanaged {
			instances, running, _, err := m.ListInstances(mig)
			if err != nil {
				return err
			}
			glog.V(4).Infof("Found %d/%d running instances for group %s", len(running), len(instances), mig.Name)
			for _, instance := range instances {
				project, zone, name, err := ParseInstanceUrl(instance)
				if err != nil {
					return err
				}
				newMigCache[GceRef{Project: project, Zone: zone, Name: name}] = mig
			}
		} else {
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
	}

	m.migCache = newMigCache
	return nil
}

// GetMigNodes returns mig nodes.
func (m *GceManager) GetMigNodes(mig *Mig) ([]string, error) {
	if mig.unmanaged {
		_, running, _, err := m.ListInstances(mig)
		if err != nil {
			return nil, err
		}
		result := make([]string, 0)
		for _, instance := range running {
			project, zone, name, err := ParseInstanceUrl(instance)
			if err != nil {
				return []string{}, err
			}
			result = append(result, fmt.Sprintf("gce://%s/%s/%s", project, zone, name))
		}
		return result, nil
	}

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

// Code borrowed from gce cloud provider. Reuse the original as soon as it becomes public.
func getProjectAndZone() (string, string, error) {
	result, err := metadata.Get("instance/zone")
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(result, "/")
	if len(parts) != 4 {
		return "", "", fmt.Errorf("unexpected response: %s", result)
	}
	zone := parts[3]
	projectID, err := metadata.ProjectID()
	if err != nil {
		return "", "", err
	}
	return projectID, zone, nil
}
