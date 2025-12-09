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

package verdacloud

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda"
	klog "k8s.io/klog/v2"
)

// ASG constants for hostname parsing and capacity management.
const (
	// ASG_SEPARATOR_MAGIC_NUMBER is the magic string used in hostname format: {prefix}-vm-{location}-{random}
	ASG_SEPARATOR_MAGIC_NUMBER = "vm"

	NO_CAPACITY_BACKOFF_DURATION      = 5 * time.Minute  // wait before retrying after no_capacity
	NO_CAPACITY_CLEANUP_AGE           = 10 * time.Minute // delete stuck no_capacity instances after this
	NO_CAPACITY_MAP_ENTRY_TTL         = 1 * time.Hour    // prevent unbounded map growth
	MAX_CONCURRENT_INSTANCE_CREATIONS = 10
)

// ASG_SEPARATOR is the separator pattern used to parse hostnames.
var ASG_SEPARATOR = fmt.Sprintf("-%s-", ASG_SEPARATOR_MAGIC_NUMBER)

type autoScalingGroups struct {
	registeredAsgs    map[AsgRef]*Asg
	asgToInstances    map[AsgRef][]InstanceRef
	instanceToAsg     map[InstanceRef]*Asg
	asgNodeGroupSpecs map[AsgRef]string
	cfg               *cloudConfig
	dcService         *verdacloudWrapper

	noCapacityInstances map[string]time.Time // tracks no_capacity instances for backoff
	lastNoCapacityCheck map[AsgRef]time.Time

	cacheMutex sync.RWMutex
}

func newAutoScalingGroups(dcService *verdacloudWrapper, nodeGroupSpecs []string, cfg *cloudConfig) (*autoScalingGroups, error) {
	registry := &autoScalingGroups{
		registeredAsgs:      make(map[AsgRef]*Asg),
		asgToInstances:      make(map[AsgRef][]InstanceRef),
		instanceToAsg:       make(map[InstanceRef]*Asg),
		asgNodeGroupSpecs:   make(map[AsgRef]string),
		noCapacityInstances: make(map[string]time.Time),
		lastNoCapacityCheck: make(map[AsgRef]time.Time),
		cfg:                 cfg,
		dcService:           dcService,
	}

	if err := registry.parseASGNodeGroupSpecs(nodeGroupSpecs); err != nil {
		return nil, err
	}

	return registry, nil
}

func (m *autoScalingGroups) getAsgs() []*Asg {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()
	asgs := make([]*Asg, 0, len(m.registeredAsgs))
	for _, asg := range m.registeredAsgs {
		asgs = append(asgs, asg)
	}
	return asgs
}

func (m *autoScalingGroups) GetAsgByRef(ref AsgRef) (*Asg, error) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()
	asg, exists := m.registeredAsgs[ref]
	if !exists {
		return nil, fmt.Errorf("ASG not found for ref: %s", ref.Name)
	}
	return asg, nil
}

func (m *autoScalingGroups) Register(asg *Asg) *Asg {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	if existing, asgExists := m.registeredAsgs[asg.AsgRef]; asgExists {
		if reflect.DeepEqual(existing, asg) {
			return existing
		}

		// Don't overwrite user-specified size/type from node group specs
		if _, asgExists := m.asgNodeGroupSpecs[asg.AsgRef]; !asgExists {
			existing.minSize = asg.minSize
			existing.maxSize = asg.maxSize
			existing.instanceType = asg.instanceType
		}

		existing.AvailabilityLocations = asg.AvailabilityLocations
		return existing
	}

	m.registeredAsgs[asg.AsgRef] = asg
	return asg
}

func (m *autoScalingGroups) FindASGForInstance(ref *InstanceRef) (*Asg, error) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	if asg, exists := m.instanceToAsg[*ref]; exists {
		return asg, nil
	}
	// ProviderID format can vary; try hostname match
	for cachedRef, asg := range m.instanceToAsg {
		if cachedRef.Hostname == ref.Hostname {
			return asg, nil
		}
	}
	klog.V(4).Infof("Instance %s not found in cache", ref.Hostname)
	return nil, nil
}

// findCachedRefByHostname requires cacheMutex held.
func (m *autoScalingGroups) findCachedRefByHostname(hostname string) (*InstanceRef, *Asg) {
	for ref, asg := range m.instanceToAsg {
		if ref.Hostname == hostname {
			return &ref, asg
		}
	}
	return nil, nil
}

func (m *autoScalingGroups) regenerate() error {
	allInstances, err := m.dcService.ListInstances("")
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	m.cacheMutex.Lock()
	specs := make([]string, 0, len(m.asgNodeGroupSpecs))
	for _, spec := range m.asgNodeGroupSpecs {
		specs = append(specs, spec)
	}
	existingAsgs := make(map[AsgRef]*Asg)
	for ref, asg := range m.registeredAsgs {
		existingAsgs[ref] = asg
	}
	existingAsgToInstances := m.asgToInstances
	m.cacheMutex.Unlock()

	newRegisteredAsgs := make(map[AsgRef]*Asg)
	newInstanceToAsg := make(map[InstanceRef]*Asg)
	newAsgToInstances := make(map[AsgRef][]InstanceRef)

	for _, spec := range specs {
		asg, err := m.buildOrReuseAsg(spec, existingAsgs)
		if err != nil {
			klog.Errorf("Failed to build ASG from spec: %v", err)
			continue
		}
		newRegisteredAsgs[asg.AsgRef] = asg

		asgInstances := m.filterInstancesForAsg(allInstances, asg)
		for _, inst := range asgInstances {
			ref := InstanceRef{
				Hostname:   inst.Hostname,
				ProviderID: verdacloudProviderIDPrefix + inst.Location + "/" + inst.Hostname,
			}
			newInstanceToAsg[ref] = asg
			newAsgToInstances[asg.AsgRef] = append(newAsgToInstances[asg.AsgRef], ref)
		}

		// Only increase; decreasing would cause duplicate scale-ups
		if len(asgInstances) > asg.curSize {
			klog.V(4).Infof("ASG %s curSize: %d -> %d", asg.Name, asg.curSize, len(asgInstances))
			asg.curSize = len(asgInstances)
		}

		m.preserveCachedInstances(asg, existingAsgToInstances, newInstanceToAsg, newAsgToInstances)
	}

	m.cacheMutex.Lock()
	m.registeredAsgs = newRegisteredAsgs
	m.instanceToAsg = newInstanceToAsg
	m.asgToInstances = newAsgToInstances
	m.cacheMutex.Unlock()

	m.handleNoCapacityInstances(allInstances, newRegisteredAsgs)

	return nil
}

// buildOrReuseAsg reuses existing pointer to preserve curSize across regenerations.
func (m *autoScalingGroups) buildOrReuseAsg(spec string, existingAsgs map[AsgRef]*Asg) (*Asg, error) {
	builtAsg, err := m.buildASGFromSpec(spec)
	if err != nil {
		return nil, err
	}

	if existing, exists := existingAsgs[builtAsg.AsgRef]; exists {
		existing.AvailabilityLocations = builtAsg.AvailabilityLocations
		return existing, nil
	}
	return builtAsg, nil
}

func (m *autoScalingGroups) filterInstancesForAsg(allInstances []verda.Instance, asg *Asg) []verda.Instance {
	matchKey := asg.hostnamePrefix
	if matchKey == "" {
		matchKey = asg.Name
	}

	var result []verda.Instance
	for _, inst := range allInstances {
		if !isActiveStatus(inst.Status) {
			continue
		}
		prefix, err := extractAsgNameFromHostname(inst.Hostname)
		if err != nil {
			continue
		}
		if strings.EqualFold(prefix, matchKey) {
			result = append(result, inst)
		}
	}
	return result
}

// preserveCachedInstances keeps provisioning instances that haven't appeared in API yet.
func (m *autoScalingGroups) preserveCachedInstances(asg *Asg, existingAsgToInstances map[AsgRef][]InstanceRef, newInstanceToAsg map[InstanceRef]*Asg, newAsgToInstances map[AsgRef][]InstanceRef) {
	if len(newAsgToInstances[asg.AsgRef]) >= asg.curSize {
		return
	}

	apiHostnames := make(map[string]bool)
	for _, ref := range newAsgToInstances[asg.AsgRef] {
		apiHostnames[ref.Hostname] = true
	}

	for _, existingRef := range existingAsgToInstances[asg.AsgRef] {
		if len(newAsgToInstances[asg.AsgRef]) >= asg.curSize {
			break
		}
		if !apiHostnames[existingRef.Hostname] {
			newInstanceToAsg[existingRef] = asg
			newAsgToInstances[asg.AsgRef] = append(newAsgToInstances[asg.AsgRef], existingRef)
		}
	}
}

func (m *autoScalingGroups) handleNoCapacityInstances(allInstances []verda.Instance, asgs map[AsgRef]*Asg) {
	now := time.Now()

	m.cacheMutex.Lock()
	for hostname, markedTime := range m.noCapacityInstances {
		if now.Sub(markedTime) > NO_CAPACITY_MAP_ENTRY_TTL {
			delete(m.noCapacityInstances, hostname)
		}
	}
	m.cacheMutex.Unlock()

	for _, asg := range asgs {
		matchKey := asg.hostnamePrefix
		if matchKey == "" {
			matchKey = asg.Name
		}

		hasNoCapacity := false
		for _, inst := range allInstances {
			if strings.ToLower(inst.Status) != verda.StatusNoCapacity {
				continue
			}
			prefix, err := extractAsgNameFromHostname(inst.Hostname)
			if err != nil || !strings.EqualFold(prefix, matchKey) {
				continue
			}

			hasNoCapacity = true
			m.processNoCapacityInstance(inst, now)
		}

		if hasNoCapacity {
			m.cacheMutex.Lock()
			m.lastNoCapacityCheck[asg.AsgRef] = now
			m.cacheMutex.Unlock()
			klog.Warningf("ASG %s has no_capacity instances, backoff until %v", asg.Name, now.Add(NO_CAPACITY_BACKOFF_DURATION))
		}
	}
}

func (m *autoScalingGroups) processNoCapacityInstance(inst verda.Instance, now time.Time) {
	m.cacheMutex.Lock()
	markedTime, exists := m.noCapacityInstances[inst.Hostname]
	if !exists {
		m.noCapacityInstances[inst.Hostname] = now
		m.cacheMutex.Unlock()
		klog.Warningf("Found no_capacity instance: %s", inst.Hostname)
		return
	}
	m.cacheMutex.Unlock()

	if now.Sub(markedTime) > NO_CAPACITY_CLEANUP_AGE {
		if err := m.dcService.DeleteInstance(inst.ID); err != nil {
			klog.Errorf("Failed to delete no_capacity instance %s: %v", inst.Hostname, err)
		} else {
			m.cacheMutex.Lock()
			delete(m.noCapacityInstances, inst.Hostname)
			m.cacheMutex.Unlock()
			klog.Infof("Deleted old no_capacity instance %s", inst.Hostname)
		}
	}
}

func (m *autoScalingGroups) buildASGFromSpec(spec string) (*Asg, error) {
	asgSpec, err := parseAsgSpec(spec)
	if err != nil {
		return nil, err
	}

	nodeConfig := m.cfg.GetNodeConfig(asgSpec.name)
	return &Asg{
		AsgRef:                AsgRef{Name: asgSpec.name},
		minSize:               asgSpec.minSize,
		maxSize:               asgSpec.maxSize,
		instanceType:          asgSpec.instanceType,
		hostnamePrefix:        asgSpec.hostnamePrefix,
		AvailabilityLocations: nodeConfig.AvailableLocations,
	}, nil
}

func (m *autoScalingGroups) parseASGNodeGroupSpecs(specs []string) error {
	for _, spec := range specs {
		asg, err := m.buildASGFromSpec(spec)
		if err != nil {
			return err
		}
		m.Register(asg)
		m.asgNodeGroupSpecs[asg.AsgRef] = spec
	}
	klog.V(4).Infof("Registered %d ASGs", len(m.asgNodeGroupSpecs))
	return nil
}

func (m *autoScalingGroups) incrementTargetSize(asg *Asg, delta int) (int, error) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	if asg.curSize+delta > asg.maxSize {
		return 0, fmt.Errorf("size increase is too large - desired:%d max:%d",
			asg.curSize+delta, asg.maxSize)
	}

	asg.curSize += delta
	return asg.curSize, nil
}

func (m *autoScalingGroups) adjustTargetSize(asg *Asg, delta int) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	asg.curSize += delta
}

func (m *autoScalingGroups) updateCacheWithInstances(asg *Asg, refs []InstanceRef) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	for _, ref := range refs {
		m.instanceToAsg[ref] = asg
		m.asgToInstances[asg.AsgRef] = append(m.asgToInstances[asg.AsgRef], ref)
	}
}

func (m *autoScalingGroups) scaleUpAsg(asg *Asg, delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	asg.scaleMutex.Lock()
	defer asg.scaleMutex.Unlock()

	klog.Infof("Scale-up ASG %s by %d (curSize=%d, max=%d)", asg.Name, delta, asg.curSize, asg.maxSize)

	// Check no_capacity backoff
	if err := m.checkNoCapacityBackoff(asg); err != nil {
		return err
	}

	newSize, err := m.incrementTargetSize(asg, delta)
	if err != nil {
		return err
	}

	rollbackNeeded := true
	defer func() {
		if rollbackNeeded {
			m.adjustTargetSize(asg, -delta)
		}
	}()

	location, err := m.dcService.GetInstanceAvailabilityLocation(asg.instanceType, asg.AvailabilityLocations)
	if err != nil {
		return fmt.Errorf("availability check failed: %w", err)
	}
	if location == "" {
		return fmt.Errorf("instance type %s not available", asg.instanceType)
	}

	nodeConfig, err := m.getNodeConfigForAsg(asg)
	if err != nil {
		return fmt.Errorf("get node config failed: %w", err)
	}

	rollbackNeeded = false // partial success is ok from here on

	refs, errs := m.createInstances(asg, nodeConfig, location, delta)
	if len(refs) > 0 {
		m.updateCacheWithInstances(asg, refs)
	}

	failedCount := delta - len(refs)
	if failedCount > 0 {
		m.adjustTargetSize(asg, -failedCount)
	}

	klog.Infof("Scale-up ASG %s complete: created %d/%d instances (curSize=%d)", asg.Name, len(refs), delta, newSize)

	if len(errs) > 0 {
		return fmt.Errorf("failed to create %d instances", len(errs))
	}
	return nil
}

func (m *autoScalingGroups) checkNoCapacityBackoff(asg *Asg) error {
	m.cacheMutex.RLock()
	lastCheck, exists := m.lastNoCapacityCheck[asg.AsgRef]
	m.cacheMutex.RUnlock()

	if exists && time.Since(lastCheck) < NO_CAPACITY_BACKOFF_DURATION {
		remaining := NO_CAPACITY_BACKOFF_DURATION - time.Since(lastCheck)
		return fmt.Errorf("no_capacity backoff active for %v", remaining)
	}
	return nil
}

func (m *autoScalingGroups) createInstances(asg *Asg, nodeConfig *nodeConfig, location string, count int) ([]InstanceRef, []error) {
	sem := make(chan struct{}, MAX_CONCURRENT_INSTANCE_CREATIONS)
	var wg sync.WaitGroup
	refsCh := make(chan InstanceRef, count)
	errsCh := make(chan error, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() { <-sem; wg.Done() }()
			_, hostname, err := m.createInstanceForAsg(asg, nodeConfig, location)
			if err != nil {
				errsCh <- err
				return
			}
			refsCh <- InstanceRef{
				Hostname:   hostname,
				ProviderID: verdacloudProviderIDPrefix + location + "/" + hostname,
			}
		}()
	}

	wg.Wait()
	close(refsCh)
	close(errsCh)

	var refs []InstanceRef
	var errs []error
	for ref := range refsCh {
		refs = append(refs, ref)
	}
	for err := range errsCh {
		errs = append(errs, err)
	}
	return refs, errs
}

func (m *autoScalingGroups) getNodeConfigForAsg(asg *Asg) (*nodeConfig, error) {
	nodeConfig := m.cfg.GetNodeConfig(asg.Name)

	isGPU := isGPUInstanceType(asg.instanceType)
	if isGPU {
		nodeConfig.Image = m.cfg.Image.GPU
	} else {
		nodeConfig.Image = m.cfg.Image.CPU
	}

	return nodeConfig, nil
}

func (m *autoScalingGroups) createInstanceForAsg(asg *Asg, nodeConfig *nodeConfig, location string) (string, string, error) {
	baseName := asg.hostnamePrefix
	if baseName == "" {
		baseName = asg.Name
	}

	hostname := strings.ReplaceAll(
		fmt.Sprintf("%s%s%s-%02d", baseName, ASG_SEPARATOR, strings.ToLower(location), rand.Intn(100)),
		".", "-")

	providerID := fmt.Sprintf("verdacloud://%s/%s", location, hostname)
	klog.V(4).Infof("Creating instance %s with providerID=%s", hostname, providerID)

	startupScriptID, err := m.createOrGetStartupScript(asg, nodeConfig, providerID)
	if err != nil {
		return "", hostname, fmt.Errorf("create startup script failed: %w", err)
	}
	if startupScriptID == "" {
		return "", hostname, errors.New("startup script creation returned empty ID")
	}
	defer m.dcService.DeleteStartScript(startupScriptID)

	image := m.cfg.Image.CPU
	if isGPUInstanceType(asg.instanceType) {
		image = m.cfg.Image.GPU
	}
	if image == "" {
		return "", hostname, fmt.Errorf("no image configured for instance type %s", asg.instanceType)
	}

	input := verda.CreateInstanceRequest{
		InstanceType:    asg.instanceType,
		Image:           image,
		SSHKeyIDs:       nodeConfig.SSHKeyIDs,
		StartupScriptID: &startupScriptID,
		Hostname:        hostname,
		Description:     asg.Name,
		LocationCode:    location,
		IsSpot:          nodeConfig.IsSpot,
		Contract:        nodeConfig.Contract,
		Pricing:         nodeConfig.Price,
		OSVolume: &verda.OSVolumeCreateRequest{
			Name: fmt.Sprintf("%s-os-volume", hostname),
			Size: nodeConfig.OSVolumeSize,
		},
	}

	if len(nodeConfig.Volumes) > 0 {
		input.Volumes = make([]verda.VolumeCreateRequest, len(nodeConfig.Volumes))
		for i, vol := range nodeConfig.Volumes {
			input.Volumes[i] = verda.VolumeCreateRequest{Name: vol.Name, Size: vol.Size, Type: vol.Type}
		}
	}

	instance, err := m.dcService.CreateInstance(&input)
	if err != nil {
		return "", hostname, fmt.Errorf("create instance failed: %w", err)
	}

	klog.V(4).Infof("Created instance %s (id=%s) for ASG %s", hostname, instance.ID, asg.Name)
	return instance.ID, hostname, nil
}

func (m *autoScalingGroups) createOrGetStartupScript(asg *Asg, nodeConfig *nodeConfig, providerID string) (string, error) {
	scriptName := fmt.Sprintf("as-%s", asg.Name)
	_base64Script, err := base64.StdEncoding.DecodeString(nodeConfig.StartupScript)
	if err != nil {
		return "", fmt.Errorf("failed to decode startup script: %v", err)
	}

	startupScriptEnv := make(map[string]string, len(nodeConfig.StartupScriptEnv)+2)
	for k, v := range nodeConfig.StartupScriptEnv {
		startupScriptEnv[strings.ToUpper(k)] = v
	}
	startupScriptEnv["PROVIDER_ID"] = providerID
	labels := convertConfigLabelsToK8sLabels(nodeConfig.Labels, asg)
	startupScriptEnv["LABELS"] = labels

	klog.V(4).Infof("Patching startup script with PROVIDER_ID=%s, LABELS=%s", providerID, labels)

	_patchedBase64Script := patchScript(_base64Script, startupScriptEnv)
	_scriptsUtf8 := string(_patchedBase64Script)

	script, err := m.dcService.CreateStartScript(scriptName, _scriptsUtf8)
	if err != nil {
		klog.Errorf("CreateStartScript API call failed: %v", err)
		return "", fmt.Errorf("failed to create startup script: %v", err)
	}

	klog.V(4).Infof("Created startup script %s (id=%s)", scriptName, script.ID)
	return script.ID, nil
}

func (m *autoScalingGroups) scaleDownAsg(asg *Asg, count int) error {
	if count <= 0 {
		return nil
	}

	asg.scaleMutex.Lock()
	defer asg.scaleMutex.Unlock()

	instances, err := m.dcService.GetActiveInstancesForAsg(asg.Name, asg.hostnamePrefix)
	if err != nil {
		return fmt.Errorf("get instances failed: %w", err)
	}
	if len(instances) < count {
		return fmt.Errorf("cannot delete %d instances, only %d available", count, len(instances))
	}
	if len(instances)-count < asg.minSize {
		return fmt.Errorf("would go below min size %d", asg.minSize)
	}

	var wg sync.WaitGroup
	errsCh := make(chan error, count)
	deletedCh := make(chan string, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(inst verda.Instance) {
			defer wg.Done()
			if err := m.dcService.PerformInstanceAction(inst.ID, verda.ActionDelete); err != nil {
				errsCh <- err
			} else {
				deletedCh <- inst.Hostname
			}
		}(instances[i])
	}
	wg.Wait()
	close(errsCh)
	close(deletedCh)

	var errs []error
	var deleted []string
	for err := range errsCh {
		errs = append(errs, err)
	}
	for hn := range deletedCh {
		deleted = append(deleted, hn)
	}

	if len(deleted) == 0 && len(errs) > 0 {
		return fmt.Errorf("failed to delete any instances: %w", errors.Join(errs...))
	}

	m.cacheMutex.Lock()
	for _, hostname := range deleted {
		m.removeInstanceFromCache(hostname, asg)
	}
	m.cacheMutex.Unlock()

	klog.V(4).Infof("Scale-down ASG %s: deleted %d/%d instances", asg.Name, len(deleted), count)
	return nil
}

// removeInstanceFromCache requires cacheMutex held.
func (m *autoScalingGroups) removeInstanceFromCache(hostname string, asg *Asg) {
	ref, _ := m.findCachedRefByHostname(hostname)
	if ref == nil {
		return
	}
	delete(m.instanceToAsg, *ref)
	refs := m.asgToInstances[asg.AsgRef]
	for i, r := range refs {
		if r.Hostname == hostname {
			m.asgToInstances[asg.AsgRef] = append(refs[:i], refs[i+1:]...)
			break
		}
	}
	asg.curSize--
}

func (m *autoScalingGroups) InstanceRefsForAsg(ref AsgRef) ([]InstanceRef, error) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()
	return m.asgToInstances[ref], nil
}

func (m *autoScalingGroups) InstancesForAsg(ref AsgRef) ([]verda.Instance, error) {
	m.cacheMutex.RLock()
	refs := append([]InstanceRef(nil), m.asgToInstances[ref]...)
	m.cacheMutex.RUnlock()

	instances := make([]verda.Instance, 0, len(refs))
	for _, r := range refs {
		inst, err := m.dcService.GetInstanceByHostname(r.Hostname)
		if err != nil {
			// Not in API yet; return placeholder
			instances = append(instances, verda.Instance{
				ID:       r.ProviderID,
				Hostname: r.Hostname,
				Status:   verda.StatusOrdered,
			})
			continue
		}
		instances = append(instances, *inst)
	}
	return instances, nil
}

func (m *autoScalingGroups) DeleteAsg(ref AsgRef) error {
	m.cacheMutex.RLock()
	instanceRefs := append([]InstanceRef(nil), m.asgToInstances[ref]...)
	asg := m.registeredAsgs[ref]
	m.cacheMutex.RUnlock()

	var wg sync.WaitGroup
	errsCh := make(chan error, len(instanceRefs))
	for _, insRef := range instanceRefs {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()
			inst, err := m.dcService.GetInstanceByHostname(hostname)
			if err != nil {
				errsCh <- err
				return
			}
			if err := m.dcService.PerformInstanceAction(inst.ID, verda.ActionDelete); err != nil {
				errsCh <- err
			}
		}(insRef.Hostname)
	}
	wg.Wait()
	close(errsCh)

	var errs []error
	for err := range errsCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to delete all instances: %w", errors.Join(errs...))
	}

	m.cacheMutex.Lock()
	for _, insRef := range instanceRefs {
		if ref, _ := m.findCachedRefByHostname(insRef.Hostname); ref != nil {
			delete(m.instanceToAsg, *ref)
		}
	}
	delete(m.registeredAsgs, ref)
	delete(m.asgToInstances, ref)
	delete(m.asgNodeGroupSpecs, ref)
	if asg != nil {
		asg.curSize = 0
	}
	m.cacheMutex.Unlock()
	return nil
}

func (m *autoScalingGroups) DeleteInstance(ref InstanceRef) error {
	m.cacheMutex.RLock()
	asg, found := m.instanceToAsg[ref]
	if !found {
		cachedRef, cachedAsg := m.findCachedRefByHostname(ref.Hostname)
		if cachedRef != nil {
			ref = *cachedRef
			asg = cachedAsg
			found = true
		}
	}
	m.cacheMutex.RUnlock()

	if !found {
		return fmt.Errorf("instance %s not found in any ASG", ref.Hostname)
	}

	inst, err := m.dcService.GetInstanceByHostname(ref.Hostname)
	if err != nil {
		return fmt.Errorf("get instance %s failed: %w", ref.Hostname, err)
	}

	if err := m.dcService.PerformInstanceAction(inst.ID, verda.ActionDelete); err != nil {
		return fmt.Errorf("delete instance %s failed: %w", inst.ID, err)
	}

	m.cacheMutex.Lock()
	m.removeInstanceFromCache(ref.Hostname, asg)
	m.cacheMutex.Unlock()

	return nil
}
