/*
Copyright 2019 The Kubernetes Authors.

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

package openstack

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

    "github.com/gophercloud/gophercloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	gcfg "gopkg.in/gcfg.v1"
)

const (
	refreshInterval         = 1 * time.Minute
	scaleToZeroSupported    = false
    // Nova
    compute_service         = "compute"
    // Heat
    orchestration_service   = "orchestration"
    // Magnum - Contaioner Orchestration Engine Service
    coe_service             = "coe"
    // Senlin
    cluster_service         = "cluster"
)

// OpenStackManager handles OpenStack communication and data caching.
type OpenStackManager interface {
	// Refresh triggers refresh of cached resources.
	Refresh() error
	// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
	Cleanup() error

	// GetASGs returns list of registered ASGs.
	GetASGs() []*ASGInformation
	// GetASGNodes returns asg nodes.
	GetASGNodes(asg ASG) ([]string, error)
	// GetASGForInstance returns ASG to which the given instance belongs.
	GetASGForInstance(instance *OpenStackRef) (ASG, error)
	// GetResourceLimiter returns resource limiter.
	GetResourceLimiter() (*cloudprovider.ResourceLimiter, error)
	// GetASGSize gets ASG size.
	GetASGSize(asg ASG) (int64, error)

	// SetASGSize sets ASG size.
	SetASGSize(asg ASG, size int64) error
	// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
	DeleteInstances(instances []*OpenStackRef) error
}

type openstackManagerImpl struct {
	cache                    OpenStackCache
	lastRefresh              time.Time
	machinesCacheLastRefresh time.Time

	OpenStackService AutoscalingOpenStackClient

	location              string
	projectId             string
	interrupt             chan struct{}
	explicitlyConfigured  map[OpenStackRef]bool
	asgAutoDiscoverySpecs []cloudprovider.OSASGAutoDiscoveryConfig
}

type AuthOptions struct {
    IdentityEndpoint string `gcfg:"identity_endpoint"`
    Username string `gcfg:"username"`
    UserID   string `gcfg:"user_id"`
    Password string `gcfg:"password"`
    DomainID   string `gcfg:"domain_id"`
    DomainName string `gcfg:"domainname"`
    TenantID   string `gcfg:"project_id"`
    TenantName string `gcfg:"projectname"`
    AllowReauth bool `gcfg:"allow_reauth"`
    TokenID string `gcfg:"token_id"`
}

type AuthScope struct {
    ProjectID   string `gcfg:"project_id"`
    ProjectName string `gcfg:"projectname"`
    DomainID    string `gcfg:"domain_id"`
    DomainName  string `gcfg:"domainname"`
}

type EndpointOpts struct {
    Type         string `gcfg:"type"`
    Name         string `gcfg:"name"`
    Region       string `gcfg:"region"`
    Availability gophercloud.Availability `gcfg:"availability"`
}

// ConfigFile is the struct used to parse the /etc/gce.conf configuration file.
type ConfigFile struct {
    Auth   AuthOptions `gcfg:"auth"`
    Endpoint EndpointOpts `gcfg:"endpoint"`
    ServiceProvider string `gcfg:"service_provider"`
}

// CreateOpenStackManager constructs OpenStackManager object.
func CreateOpenStackManager(configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (OpenStackManager, error) {
	if configReader != nil {
        var cfg ConfigFile
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			klog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
        var authOpts = gophercloud.AuthOptions{
            IdentityEndpoint: cfg.Auth.IdentityEndpoint,
            Username: cfg.Auth.Username,
            UserID: cfg.Auth.UserID,
            Password: cfg.Auth.Password,
            TenantName: cfg.Auth.TenantName,
            DomainName: cfg.Auth.DomainName,
            DomainID: cfg.Auth.DomainID,
            TenantID: cfg.Auth.TenantID,
            AllowReauth: cfg.Auth.AllowReauth,
            TokenID: cfg.Auth.TokenID,
        }
        var endpointOpts = gophercloud.EndpointOpts{
            Type: cfg.Endpoint.Type,
            Name: cfg.Endpoint.Name,
            Region: cfg.Endpoint.Region,
            Availability: cfg.Endpoint.Availability,
        }
        var openstackService AutoscalingOpenStackClient
        err := fmt.Errorf("Current Service Provider %v didn't supported.", cfg.ServiceProvider)
        switch cfg.ServiceProvider {
        case coe_service:
            openstackService, err = NewMagnumClient(authOpts,  endpointOpts)
            if err != nil {
                return nil, err
            }
        case orchestration_service:
            //TODO add Heat support
            return nil, err
            // openstackService, err = NewAutoscalingOrchestrationClient(authOpts,  endpointOpts)
            // if err != nil {
            //    return nil, err
            //}
        case compute_service:
            //TODO add Nova support
            return nil, err
        case cluster_service:
            //TODO add Senlin support
            return nil, err
        }
        cache := NewOpenStackCache(openstackService)
        manager := &openstackManagerImpl{
            OpenStackService:     openstackService,
            cache:                cache,
            interrupt:            make(chan struct{}),
            explicitlyConfigured: make(map[OpenStackRef]bool),
        }
        if err := manager.fetchExplicitASGs(discoveryOpts.NodeGroupSpecs); err != nil {
            return nil, fmt.Errorf("failed to fetch ASGs: %v", err)
        }
        if spec, err := discoveryOpts.ParseOSASGAutoDiscoverySpecs(); err != nil {
            return nil, err
        } else {
            manager.asgAutoDiscoverySpecs = spec
        }
        if err := manager.forceRefresh(); err != nil {
            return nil, err
        }
        go wait.Until(func() {
            if err := manager.cache.RegenerateInstancesCache(); err != nil {
                klog.Errorf("Error while regenerating ASG cache: %v", err)
            }
        }, time.Hour, manager.interrupt)
        return manager, nil
    }
    return nil, nil
}

// Cleanup closes the channel to stop the goroutine refreshing cache.
func (m *openstackManagerImpl) Cleanup() error {
	close(m.interrupt)
	return nil
}

// registerASG registers asg in OpenStackManager. Returns true if the node group didn't exist before or its config has changed.
func (m *openstackManagerImpl) registerASG(asg ASG) bool {
	changed := m.cache.RegisterASG(asg)
	return changed
}

// GetASGSize gets ASG size.
func (m *openstackManagerImpl) GetASGSize(asg ASG) (int64, error) {
	targetSize, err := m.OpenStackService.FetchASGTargetSize(asg.OpenStackRef())
	if err != nil {
		return -1, err
	}
	return targetSize, nil
}

// SetASGSize sets ASG size.
func (m *openstackManagerImpl) SetASGSize(asg ASG, size int64) error {
	klog.V(0).Infof("Setting asg size %s to %d", asg.Id(), size)
	return m.OpenStackService.ResizeASG(asg.OpenStackRef(), size)
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (m *openstackManagerImpl) DeleteInstances(instances []*OpenStackRef) error {
	if len(instances) == 0 {
		return nil
	}
	commonASG, err := m.GetASGForInstance(instances[0])
	if err != nil {
		return err
	}
	for _, instance := range instances {
		asg, err := m.GetASGForInstance(instance)
		if err != nil {
			return err
		}
		if asg != commonASG {
			return fmt.Errorf("Cannot delete instances which don't belong to the same ASG.")
		}
	}
	return m.OpenStackService.DeleteInstances(commonASG.OpenStackRef(), instances)
}

// GetASGs returns list of registered ASGs.
func (m *openstackManagerImpl) GetASGs() []*ASGInformation {
	return m.cache.GetASGs()
}

// GetASGForInstance returns ASG to which the given instance belongs.
func (m *openstackManagerImpl) GetASGForInstance(instance *OpenStackRef) (ASG, error) {
	return m.cache.GetASGForInstance(instance)
}

// GetASGNodes returns asg nodes.
func (m *openstackManagerImpl) GetASGNodes(asg ASG) ([]string, error) {
	instances, err := m.OpenStackService.FetchASGInstances(asg.OpenStackRef())
	if err != nil {
		return []string{}, err
	}
	result := make([]string, 0)
	for _, ref := range instances {
        result = append(result, fmt.Sprintf("%s/%s/%s/%s", ref.Project, ref.RootResource, ref.Resource, ref.Name))
	}
	return result, nil
}

// Refresh triggers refresh of cached resources.
func (m *openstackManagerImpl) Refresh() error {
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.forceRefresh()
}

func (m *openstackManagerImpl) forceRefresh() error {
	if err := m.fetchAutoASGs(); err != nil {
		klog.Errorf("Failed to fetch ASGs: %v", err)
		return err
	}
	m.lastRefresh = time.Now()
	klog.V(2).Infof("Refreshed OpenStack resources, next refresh after %v", m.lastRefresh.Add(refreshInterval))
	return nil
}

// Fetch explicitly configured ASGs. These ASGs should never be unregistered
// during refreshes, even if they no longer exist in OpenStack.
func (m *openstackManagerImpl) fetchExplicitASGs(specs []string) error {
	changed := false
	for _, spec := range specs {
		asg, err := m.buildASGFromFlag(spec)
		if err != nil {
			return err
		}
		if m.registerASG(asg) {
			changed = true
		}
		m.explicitlyConfigured[asg.OpenStackRef()] = true
	}

	if changed {
		return m.cache.RegenerateInstancesCache()
	}
	return nil
}

func (m *openstackManagerImpl) buildASGFromFlag(flag string) (ASG, error) {
	s, err := dynamic.SpecFromString(flag, scaleToZeroSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	return m.buildASGFromSpec(s)
}

func (m *openstackManagerImpl) buildASGFromAutoCfg(link string, cfg cloudprovider.OSASGAutoDiscoveryConfig) (ASG, error) {
	s := &dynamic.NodeGroupSpec{
		Name:               link,
		MinSize:            cfg.MinSize,
		MaxSize:            cfg.MaxSize,
		SupportScaleToZero: scaleToZeroSupported,
	}
	return m.buildASGFromSpec(s)
}

func (m *openstackManagerImpl) buildASGFromSpec(s *dynamic.NodeGroupSpec) (ASG, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid node group spec: %v", err)
	}
	openStackRef, err := ParseASGUrl(s.Name)
	if err != nil {
        return nil, fmt.Errorf("failed to parse asg url: %s got error: %v", s.Name, err)
    }
    asg := &openstackASG{
        openstackRef: *openStackRef,
        openstackManager: m,
        minSize:    s.MinSize,
        maxSize:    s.MaxSize,
    }
	return asg, nil
}

func ParseASGUrl(url string) (ref *OpenStackRef, err error) {
	splitted := strings.Split(url, "/")
	if len(splitted) != 4 {
		return nil, fmt.Errorf("Wrong id: expected format <project-id>/<root_resource_id>/<resource_id>/<name>, got %v", url)
	}
	return &OpenStackRef{
		Project:        splitted[0],
		RootResource:   splitted[1],
		Resource:       splitted[2],
		Name:           splitted[3],
	}, nil
}

// Fetch automatically discovered ASGs. These ASGs should be unregistered if
// they no longer exist in OpenStack.
func (m *openstackManagerImpl) fetchAutoASGs() error {
	exists := make(map[OpenStackRef]bool)
	changed := false
	for _, cfg := range m.asgAutoDiscoverySpecs {
		links, err := m.findASGsNamed(cfg.Re)
		if err != nil {
			return fmt.Errorf("cannot autodiscover managed instance groups: %v", err)
		}
		for _, link := range links {
			asg, err := m.buildASGFromAutoCfg(link, cfg)
			if err != nil {
				return err
			}
			exists[asg.OpenStackRef()] = true
			if m.explicitlyConfigured[asg.OpenStackRef()] {
				// This ASG was explicitly configured, but would also be
				// autodiscovered. We want the explicitly configured min and max
				// nodes to take precedence.
				klog.V(3).Infof("Ignoring explicitly configured ASG %s in autodiscovery.", asg.OpenStackRef().String())
				continue
			}
			if m.registerASG(asg) {
				klog.V(3).Infof("Autodiscovered ASG %s using regexp %s", asg.OpenStackRef().String(), cfg.Re.String())
				changed = true
			}
		}
	}

	for _, asg := range m.GetASGs() {
		if !exists[asg.Config.OpenStackRef()] && !m.explicitlyConfigured[asg.Config.OpenStackRef()] {
			m.cache.UnregisterASG(asg.Config)
			changed = true
		}
	}

	if changed {
		return m.cache.RegenerateInstancesCache()
	}

	return nil
}

// GetResourceLimiter returns resource limiter from cache.
func (m *openstackManagerImpl) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return m.cache.GetResourceLimiter()
}

func (m *openstackManagerImpl) findASGsNamed(name *regexp.Regexp) ([]string, error) {
	return m.OpenStackService.FetchASGsWithName(name)
}
