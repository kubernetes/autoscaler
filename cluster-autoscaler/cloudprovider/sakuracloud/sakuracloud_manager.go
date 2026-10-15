/*
Copyright The Kubernetes Authors.

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

package sakuracloud

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	apiBaseFormat = "https://secure.sakura.ad.jp/cloud/zone/%s/api/cloud/1.1"
	// groupTagPrefix marks a server as a member of a cluster-autoscaler
	// node group: "ca-group-<nodeGroupName>".
	groupTagPrefix = "ca-group-"

	diskAvailabilityTimeout = 10 * time.Minute
	serverDeleteTimeout     = 5 * time.Minute
	pollInterval            = 5 * time.Second
)

// nodeGroupConfig is the per-group section of SAKURACLOUD_CLUSTER_CONFIG.
type nodeGroupConfig struct {
	MinSize         int               `json:"minSize"`
	MaxSize         int               `json:"maxSize"`
	Core            int               `json:"core"`
	MemoryGB        int               `json:"memoryGB"`
	DiskGB          int               `json:"diskGB"`
	SourceArchiveID string            `json:"sourceArchiveID"`
	StartupNoteID   string            `json:"startupNoteID"`
	Labels          map[string]string `json:"labels"`
	Taints          []apiv1.Taint     `json:"taints"`
}

// clusterConfig is the SAKURACLOUD_CLUSTER_CONFIG JSON document.
type clusterConfig struct {
	Zone       string                     `json:"zone"`
	NodeGroups map[string]nodeGroupConfig `json:"nodeGroups"`
}

// sakuraServer is the subset of the IaaS API server representation we need.
type sakuraServer struct {
	ID             json.Number `json:"ID"`
	Name           string      `json:"Name"`
	Tags           []string    `json:"Tags"`
	InstanceStatus string      `json:"-"`
	Instance       *struct {
		Status string `json:"Status"`
	} `json:"Instance,omitempty"`
	Disks []struct {
		ID json.Number `json:"ID"`
	} `json:"Disks"`
}

func (s *sakuraServer) status() string {
	if s.Instance != nil {
		return s.Instance.Status
	}
	return ""
}

func (s *sakuraServer) groupName() (string, bool) {
	for _, t := range s.Tags {
		if strings.HasPrefix(t, groupTagPrefix) {
			return strings.TrimPrefix(t, groupTagPrefix), true
		}
	}
	return "", false
}

// sakuracloudManager talks to the SAKURA cloud IaaS API and caches state.
type sakuracloudManager struct {
	token      string
	secret     string
	zone       string
	apiBase    string
	httpClient *http.Client

	clusterConfig *clusterConfig
	nodeGroups    map[string]*sakuracloudNodeGroup

	mu      sync.Mutex
	servers []sakuraServer
}

func newManager() (*sakuracloudManager, error) {
	token := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	secret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")
	if token == "" || secret == "" {
		return nil, fmt.Errorf("SAKURACLOUD_ACCESS_TOKEN and SAKURACLOUD_ACCESS_TOKEN_SECRET are required")
	}
	rawConfig := os.Getenv("SAKURACLOUD_CLUSTER_CONFIG")
	if rawConfig == "" {
		return nil, fmt.Errorf("SAKURACLOUD_CLUSTER_CONFIG is required")
	}
	var cfg clusterConfig
	if err := json.Unmarshal([]byte(rawConfig), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse SAKURACLOUD_CLUSTER_CONFIG: %w", err)
	}
	if cfg.Zone == "" {
		return nil, fmt.Errorf("SAKURACLOUD_CLUSTER_CONFIG: zone is required")
	}
	if len(cfg.NodeGroups) == 0 {
		return nil, fmt.Errorf("SAKURACLOUD_CLUSTER_CONFIG: at least one node group is required")
	}
	for name, ng := range cfg.NodeGroups {
		if ng.MinSize < 0 || ng.MaxSize <= 0 || ng.MinSize > ng.MaxSize {
			return nil, fmt.Errorf("node group %q: require 0 <= minSize <= maxSize and maxSize > 0, have min %d max %d", name, ng.MinSize, ng.MaxSize)
		}
		if ng.Core <= 0 || ng.MemoryGB <= 0 || ng.DiskGB <= 0 {
			return nil, fmt.Errorf("node group %q: core, memoryGB and diskGB must be positive", name)
		}
		if ng.SourceArchiveID == "" || ng.StartupNoteID == "" {
			return nil, fmt.Errorf("node group %q: sourceArchiveID and startupNoteID are required", name)
		}
	}

	m := &sakuracloudManager{
		token:         token,
		secret:        secret,
		zone:          cfg.Zone,
		apiBase:       fmt.Sprintf(apiBaseFormat, cfg.Zone),
		httpClient:    &http.Client{Timeout: 60 * time.Second},
		clusterConfig: &cfg,
		nodeGroups:    map[string]*sakuracloudNodeGroup{},
	}

	if err := m.refreshServers(); err != nil {
		return nil, fmt.Errorf("initial server list failed: %w", err)
	}
	for name, ngCfg := range cfg.NodeGroups {
		cfgCopy := ngCfg
		ng := &sakuracloudNodeGroup{
			id:      name,
			manager: m,
			config:  &cfgCopy,
		}
		ng.targetSize = len(m.serversInGroup(name))
		m.nodeGroups[name] = ng
	}
	return m, nil
}

// doRequest performs one API call. GET parameters are passed as a raw JSON
// document in the query string, which is the convention of the IaaS API.
func (m *sakuracloudManager) doRequest(method, path string, params interface{}, out interface{}) error {
	reqURL := m.apiBase + path
	var body io.Reader
	if params != nil {
		encoded, err := json.Marshal(params)
		if err != nil {
			return err
		}
		if method == http.MethodGet {
			reqURL += "?" + url.QueryEscape(string(encoded))
		} else {
			body = bytes.NewReader(encoded)
		}
	}
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return err
	}
	req.SetBasicAuth(m.token, m.secret)
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sakuracloud API %s %s: status %d: %s", method, path, resp.StatusCode, truncate(string(data), 300))
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("sakuracloud API %s %s: failed to decode response: %w", method, path, err)
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// refreshServers reloads the full server list for the zone. Membership is
// derived from the ca-group-* tag, so a plain list is sufficient.
func (m *sakuracloudManager) refreshServers() error {
	var result struct {
		Servers []sakuraServer `json:"Servers"`
	}
	if err := m.doRequest(http.MethodGet, "/server", map[string]interface{}{"Count": 500}, &result); err != nil {
		return err
	}
	m.mu.Lock()
	m.servers = result.Servers
	m.mu.Unlock()
	return nil
}

func (m *sakuracloudManager) serversInGroup(group string) []sakuraServer {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []sakuraServer
	for _, s := range m.servers {
		if g, ok := s.groupName(); ok && g == group {
			out = append(out, s)
		}
	}
	return out
}

func (m *sakuracloudManager) serverByName(name string) *sakuraServer {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.servers {
		if m.servers[i].Name == name {
			s := m.servers[i]
			return &s
		}
	}
	return nil
}

// findServerPlanID resolves a standard-commitment server plan for the
// requested cpu/memory combination.
func (m *sakuracloudManager) findServerPlanID(core, memoryGB int) (string, error) {
	var result struct {
		ServerPlans []struct {
			ID           json.Number `json:"ID"`
			CPU          int         `json:"CPU"`
			MemoryMB     int         `json:"MemoryMB"`
			Commitment   string      `json:"Commitment"`
			Availability string      `json:"Availability"`
			Generation   int         `json:"Generation"`
		} `json:"ServerPlans"`
	}
	if err := m.doRequest(http.MethodGet, "/product/server", map[string]interface{}{"Count": 1000}, &result); err != nil {
		return "", err
	}
	bestID := ""
	bestGen := -1
	for _, p := range result.ServerPlans {
		if p.CPU == core && p.MemoryMB == memoryGB*1024 &&
			p.Commitment == "standard" && p.Availability == "available" && p.Generation > bestGen {
			bestID = p.ID.String()
			bestGen = p.Generation
		}
	}
	if bestID == "" {
		return "", fmt.Errorf("no available standard server plan for %d core / %d GB in zone %s", core, memoryGB, m.zone)
	}
	return bestID, nil
}

func randomSuffix() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 5)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return string(b)
}

func randomPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 24)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return string(b)
}

// createServer provisions one node group member end to end: disk copy from
// the source archive, server creation on the shared segment, disk attach,
// disk config (hostname + startup note) and power on. SAKURA cloud has no
// server-group primitive, so this is the equivalent of an ASG launch.
func (m *sakuracloudManager) createServer(group string, cfg *nodeGroupConfig) error {
	name := fmt.Sprintf("%s-%s", group, randomSuffix())
	klog.V(2).Infof("sakuracloud: provisioning server %s for node group %s", name, group)

	// 1. disk (copy from archive; the copy is what takes time)
	var diskResp struct {
		Disk struct {
			ID json.Number `json:"ID"`
		} `json:"Disk"`
	}
	err := m.doRequest(http.MethodPost, "/disk", map[string]interface{}{
		"Disk": map[string]interface{}{
			"Name":          name,
			"Plan":          map[string]interface{}{"ID": 4}, // SSD
			"SizeMB":        cfg.DiskGB * 1024,
			"Connection":    "virtio",
			"SourceArchive": map[string]interface{}{"ID": cfg.SourceArchiveID},
		},
	}, &diskResp)
	if err != nil {
		return fmt.Errorf("disk create failed: %w", err)
	}
	diskID := diskResp.Disk.ID.String()

	cleanupDisk := func() {
		if derr := m.doRequest(http.MethodDelete, "/disk/"+diskID, nil, nil); derr != nil {
			klog.Errorf("sakuracloud: failed to clean up disk %s: %v", diskID, derr)
		}
	}

	if err := m.waitDiskAvailable(diskID); err != nil {
		cleanupDisk()
		return err
	}

	// 2. server on the shared segment
	var serverResp struct {
		Server struct {
			ID json.Number `json:"ID"`
		} `json:"Server"`
	}
	// The server plan is specified by CPU/memory spec, not by plan ID:
	// plan-ID form is rejected with 400 by the current API (observed).
	err = m.doRequest(http.MethodPost, "/server", map[string]interface{}{
		"Server": map[string]interface{}{
			"Name":              name,
			"ServerPlan":        map[string]interface{}{"CPU": cfg.Core, "MemoryMB": cfg.MemoryGB * 1024},
			"Description":       "managed by cluster-autoscaler",
			"Tags":              []string{groupTagPrefix + group},
			"ConnectedSwitches": []map[string]interface{}{{"Scope": "shared"}},
		},
	}, &serverResp)
	if err != nil {
		cleanupDisk()
		return fmt.Errorf("server create failed: %w", err)
	}
	serverID := serverResp.Server.ID.String()

	// cleanupServer tears the half-provisioned server down again so a
	// post-create failure does not leave a billable, group-tagged server
	// behind (it would be counted by later refreshes). withDisk deletes the
	// attached disk in the same call; before attachment the disk is removed
	// separately via cleanupDisk.
	cleanupServer := func(withDisk bool) {
		if derr := m.doRequest(http.MethodDelete, "/server/"+serverID+"/power", map[string]interface{}{"Force": true}, nil); derr != nil && !strings.Contains(derr.Error(), "status 409") {
			klog.Warningf("sakuracloud: cleanup power off of server %s failed: %v", serverID, derr)
		}
		payload := map[string]interface{}{}
		if withDisk {
			payload["WithDisk"] = []string{diskID}
		}
		if derr := m.doRequest(http.MethodDelete, "/server/"+serverID, payload, nil); derr != nil {
			klog.Errorf("sakuracloud: failed to clean up server %s after provisioning failure: %v", serverID, derr)
			return
		}
		if !withDisk {
			cleanupDisk()
		}
	}

	// 3. attach disk and inject hostname + startup note
	if err := m.doRequest(http.MethodPut, "/disk/"+diskID+"/to/server/"+serverID, nil, nil); err != nil {
		cleanupServer(false)
		return fmt.Errorf("disk attach failed: %w", err)
	}
	err = m.doRequest(http.MethodPut, "/disk/"+diskID+"/config", map[string]interface{}{
		"HostName":      name,
		"Password":      randomPassword(),
		"DisablePWAuth": true,
		"Notes":         []map[string]interface{}{{"ID": cfg.StartupNoteID}},
	}, nil)
	if err != nil {
		cleanupServer(true)
		return fmt.Errorf("disk config failed: %w", err)
	}

	// The config write (hostname/startup note injection) puts the disk into a
	// modifying state; powering on before it settles fails with 409
	// disk_is_not_available (observed). Wait for it to become available again.
	if err := m.waitDiskAvailable(diskID); err != nil {
		cleanupServer(true)
		return err
	}

	// 4. power on
	if err := m.doRequest(http.MethodPut, "/server/"+serverID+"/power", nil, nil); err != nil {
		cleanupServer(true)
		return fmt.Errorf("power on failed: %w", err)
	}
	klog.V(2).Infof("sakuracloud: server %s (id %s) powered on", name, serverID)
	return nil
}

func (m *sakuracloudManager) waitDiskAvailable(diskID string) error {
	deadline := time.Now().Add(diskAvailabilityTimeout)
	for time.Now().Before(deadline) {
		var resp struct {
			Disk struct {
				Availability string `json:"Availability"`
			} `json:"Disk"`
		}
		if err := m.doRequest(http.MethodGet, "/disk/"+diskID, nil, &resp); err != nil {
			return err
		}
		switch resp.Disk.Availability {
		case "available":
			return nil
		case "failed":
			return fmt.Errorf("disk %s entered failed state", diskID)
		}
		time.Sleep(pollInterval)
	}
	return fmt.Errorf("disk %s did not become available within %s", diskID, diskAvailabilityTimeout)
}

// deleteServer force-stops a server and deletes it together with its disks.
func (m *sakuracloudManager) deleteServer(s *sakuraServer) error {
	serverID := s.ID.String()
	klog.V(2).Infof("sakuracloud: deleting server %s (id %s)", s.Name, serverID)

	// The /server list response does not reliably include the instance power
	// state, so always attempt a force power-off and treat a 409 (already
	// down) as success, then poll until the server reports non-up.
	if err := m.doRequest(http.MethodDelete, "/server/"+serverID+"/power", map[string]interface{}{"Force": true}, nil); err != nil {
		if !strings.Contains(err.Error(), "status 409") {
			return fmt.Errorf("power off failed: %w", err)
		}
	}
	deadline := time.Now().Add(serverDeleteTimeout)
	for time.Now().Before(deadline) {
		var resp struct {
			Server sakuraServer `json:"Server"`
		}
		if err := m.doRequest(http.MethodGet, "/server/"+serverID, nil, &resp); err != nil {
			return err
		}
		if resp.Server.status() != "up" {
			break
		}
		time.Sleep(pollInterval)
	}

	diskIDs := make([]json.Number, 0, len(s.Disks))
	for _, d := range s.Disks {
		diskIDs = append(diskIDs, d.ID)
	}
	if err := m.doRequest(http.MethodDelete, "/server/"+serverID, map[string]interface{}{"WithDisk": diskIDs}, nil); err != nil {
		return fmt.Errorf("server delete failed: %w", err)
	}
	return nil
}
