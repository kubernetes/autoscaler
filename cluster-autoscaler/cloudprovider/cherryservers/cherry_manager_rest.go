/*
Copyright 2022 The Kubernetes Authors.

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

package cherryservers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"gopkg.in/gcfg.v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	klog "k8s.io/klog/v2"
)

const (
	userAgent                    = "kubernetes/cluster-autoscaler/" + version.ClusterAutoscalerVersion
	expectedAPIContentTypePrefix = "application/json"
	cherryPrefix                 = "cherryservers://"
	baseURL                      = "https://api.cherryservers.com/v1/"
)

type instanceType struct {
	InstanceName string
	CPU          int64
	MemoryMb     int64
	GPU          int64
}

type cherryManagerNodePool struct {
	clusterName       string
	projectID         int
	apiServerEndpoint string
	region            string
	plan              string
	os                string
	cloudinit         string
	hostnamePattern   string
	sshKeyIDs         []int
	osPartitionSize   int
	waitTimeStep      time.Duration
}

type cherryManagerRest struct {
	authToken  string
	baseURL    *url.URL
	nodePools  map[string]*cherryManagerNodePool
	plans      map[string]*Plan
	planUpdate time.Time
}

// ConfigNodepool options only include the project-id for now
type ConfigNodepool struct {
	ClusterName       string   `gcfg:"cluster-name"`
	ProjectID         int      `gcfg:"project-id"`
	APIServerEndpoint string   `gcfg:"api-server-endpoint"`
	Region            string   `gcfg:"region"`
	Plan              string   `gcfg:"plan"`
	OS                string   `gcfg:"os"`
	SSHKeys           []string `gcfg:"ssh-key-ids"`
	CloudInit         string   `gcfg:"cloudinit"`
	HostnamePattern   string   `gcfg:"hostname-pattern"`
	OsPartitionSize   int      `gcfg:"os-partition-size"`
}

// IsEmpty determine if this is an empty config
func (c ConfigNodepool) IsEmpty() bool {
	return c.ClusterName == "" && c.CloudInit == "" && c.Region == "" && c.Plan == "" && c.ProjectID == 0
}

// ConfigFile is used to read and store information from the cloud configuration file
type ConfigFile struct {
	DefaultNodegroupdef ConfigNodepool             `gcfg:"global"`
	Nodegroupdef        map[string]*ConfigNodepool `gcfg:"nodegroupdef"`
}

// CloudInitTemplateData represents the variables that can be used in cloudinit templates
type CloudInitTemplateData struct {
	BootstrapTokenID     string
	BootstrapTokenSecret string
	APIServerEndpoint    string
	NodeGroup            string
}

// HostnameTemplateData represents the template variables used to construct host names for new nodes
type HostnameTemplateData struct {
	ClusterName string
	NodeGroup   string
	RandString8 string
}

// ErrorResponse is the http response used on errors
type ErrorResponse struct {
	Response    *http.Response
	Errors      []string `json:"errors"`
	SingleError string   `json:"error"`
}

var multipliers = map[string]int64{
	"KB": 1024,
	"MB": 1024 * 1024,
	"GB": 1024 * 1024 * 1024,
	"TB": 1024 * 1024 * 1024 * 1024,
}

// Error implements the error interface
func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, strings.Join(r.Errors, ", "), r.SingleError)
}

// Find returns the smallest index i at which x == a[i],
// or len(a) if there is no such index.
func Find(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// createCherryManagerRest sets up the client and returns
// an cherryManagerRest.
func createCherryManagerRest(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (*cherryManagerRest, error) {
	// Initialize ConfigFile instance
	cfg := ConfigFile{
		DefaultNodegroupdef: ConfigNodepool{},
		Nodegroupdef:        map[string]*ConfigNodepool{},
	}

	if configReader != nil {
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			klog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	}

	var manager cherryManagerRest
	manager.nodePools = make(map[string]*cherryManagerNodePool)

	if _, ok := cfg.Nodegroupdef["default"]; !ok {
		cfg.Nodegroupdef["default"] = &cfg.DefaultNodegroupdef
	}

	if cfg.Nodegroupdef["default"].IsEmpty() {
		klog.Fatalf("No \"default\" or [Global] nodepool definition was found")
	}

	cherryAuthToken := os.Getenv("CHERRY_AUTH_TOKEN")
	if len(cherryAuthToken) == 0 {
		klog.Fatalf("CHERRY_AUTH_TOKEN is required and missing")
	}

	manager.authToken = cherryAuthToken
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid baseURL %s: %v", baseURL, err)
	}

	manager.baseURL = base

	projectID := cfg.Nodegroupdef["default"].ProjectID
	apiServerEndpoint := cfg.Nodegroupdef["default"].APIServerEndpoint

	for key, nodepool := range cfg.Nodegroupdef {
		if opts.ClusterName == "" && nodepool.ClusterName == "" {
			klog.Fatalf("The cluster-name parameter must be set")
		} else if opts.ClusterName != "" && nodepool.ClusterName == "" {
			nodepool.ClusterName = opts.ClusterName
		}

		var sshKeyIDs []int
		for i, keyIDString := range nodepool.SSHKeys {
			keyID, err := strconv.ParseInt(keyIDString, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid ssh-key ID at position %d: %s; it must be an integer", i, keyIDString)
			}
			sshKeyIDs = append(sshKeyIDs, int(keyID))
		}
		manager.nodePools[key] = &cherryManagerNodePool{
			projectID:         projectID,
			apiServerEndpoint: apiServerEndpoint,
			clusterName:       nodepool.ClusterName,
			region:            nodepool.Region,
			plan:              nodepool.Plan,
			os:                nodepool.OS,
			cloudinit:         nodepool.CloudInit,
			sshKeyIDs:         sshKeyIDs,
			hostnamePattern:   nodepool.HostnamePattern,
			osPartitionSize:   nodepool.OsPartitionSize,
		}
	}

	return &manager, nil
}

func (mgr *cherryManagerRest) request(ctx context.Context, method, pathUrl string, jsonData []byte) ([]byte, error) {
	u, err := url.Parse(pathUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid request path %s: %v", pathUrl, err)
	}
	reqUrl := mgr.baseURL.ResolveReference(u)

	req, err := http.NewRequestWithContext(ctx, method, reqUrl.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mgr.authToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	dump, _ := httputil.DumpRequestOut(req, true)
	klog.V(2).Infof("%s", string(dump))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			klog.Errorf("failed to close response body: %v", err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, expectedAPIContentTypePrefix) {
		errorResponse := &ErrorResponse{Response: resp}
		errorResponse.SingleError = fmt.Sprintf("Unexpected Content-Type: %s with status: %s", ct, resp.Status)
		return nil, errorResponse
	}

	// If the response is good return early
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, nil
	}

	errorResponse := &ErrorResponse{Response: resp}

	if len(body) > 0 {
		if err := json.Unmarshal(body, errorResponse); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return nil, errorResponse
}

func (mgr *cherryManagerRest) listCherryPlans(ctx context.Context) (Plans, error) {
	req := "plans"

	result, err := mgr.request(ctx, "GET", req, []byte(``))
	if err != nil {
		return nil, err
	}

	var plans Plans
	if err := json.Unmarshal(result, &plans); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return plans, nil
}

func (mgr *cherryManagerRest) listCherryServers(ctx context.Context) ([]Server, error) {
	pool := mgr.getNodePoolDefinition("default")
	req := path.Join("projects", fmt.Sprintf("%d", pool.projectID), "servers")

	result, err := mgr.request(ctx, "GET", req, []byte(``))
	if err != nil {
		return nil, err
	}

	var servers []Server
	if err := json.Unmarshal(result, &servers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return servers, nil
}

func (mgr *cherryManagerRest) getCherryServer(ctx context.Context, id string) (*Server, error) {
	req := path.Join("servers", id)

	result, err := mgr.request(ctx, "GET", req, []byte(``))
	if err != nil {
		return nil, err
	}

	var server Server
	if err := json.Unmarshal(result, &server); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &server, nil
}

func (mgr *cherryManagerRest) NodeGroupForNode(labels map[string]string, nodeId string) (string, error) {
	if nodegroup, ok := labels["pool"]; ok {
		return nodegroup, nil
	}

	trimmedNodeId := strings.TrimPrefix(nodeId, cherryPrefix)

	server, err := mgr.getCherryServer(context.TODO(), trimmedNodeId)
	if err != nil {
		return "", fmt.Errorf("Could not find group for node: %s %s", nodeId, err)
	}
	for k, v := range server.Tags {
		if k == "k8s-nodepool" {
			return v, nil
		}
	}
	return "", nil
}

// nodeGroupSize gets the current size of the nodegroup as reported by Cherry Servers tags.
func (mgr *cherryManagerRest) nodeGroupSize(nodegroup string) (int, error) {
	servers, err := mgr.listCherryServers(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("failed to list servers: %w", err)
	}

	// Get the count of servers tagged as nodegroup members
	count := 0
	for _, s := range servers {
		clusterName, ok := s.Tags["k8s-cluster"]
		if !ok || clusterName != mgr.getNodePoolDefinition(nodegroup).clusterName {
			continue
		}
		nodepoolName, ok := s.Tags["k8s-nodepool"]
		if !ok || nodegroup != nodepoolName {
			continue
		}
		count++
	}
	klog.V(3).Infof("Nodegroup %s: %d/%d", nodegroup, count, len(servers))
	return count, nil
}

func randString8() string {
	n := 8
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("acdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// createNode creates a cluster node by creating a server with the appropriate userdata to add it to the cluster.
func (mgr *cherryManagerRest) createNode(ctx context.Context, cloudinit, nodegroup string) error {
	udvars := CloudInitTemplateData{
		BootstrapTokenID:     os.Getenv("BOOTSTRAP_TOKEN_ID"),
		BootstrapTokenSecret: os.Getenv("BOOTSTRAP_TOKEN_SECRET"),
		APIServerEndpoint:    mgr.getNodePoolDefinition(nodegroup).apiServerEndpoint,
		NodeGroup:            nodegroup,
	}

	ud, err := renderTemplate(cloudinit, udvars)
	if err != nil {
		return fmt.Errorf("failed to create userdata from template: %w", err)
	}

	hnvars := HostnameTemplateData{
		ClusterName: mgr.getNodePoolDefinition(nodegroup).clusterName,
		NodeGroup:   nodegroup,
		RandString8: randString8(),
	}
	hn, err := renderTemplate(mgr.getNodePoolDefinition(nodegroup).hostnamePattern, hnvars)
	if err != nil {
		return fmt.Errorf("failed to create hostname from template: %w", err)
	}
	cr := &CreateServer{
		Hostname:        hn,
		Region:          mgr.getNodePoolDefinition(nodegroup).region,
		Plan:            mgr.getNodePoolDefinition(nodegroup).plan,
		Image:           mgr.getNodePoolDefinition(nodegroup).os,
		ProjectID:       mgr.getNodePoolDefinition(nodegroup).projectID,
		UserData:        base64.StdEncoding.EncodeToString([]byte(ud)),
		SSHKeys:         mgr.getNodePoolDefinition(nodegroup).sshKeyIDs,
		Tags:            &map[string]string{"k8s-cluster": mgr.getNodePoolDefinition(nodegroup).clusterName, "k8s-nodepool": nodegroup},
		OSPartitionSize: mgr.getNodePoolDefinition(nodegroup).osPartitionSize,
	}

	if err := mgr.createServerRequest(ctx, cr, nodegroup); err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	klog.Infof("Created new node on Cherry Servers.")

	return nil
}

// createNodes provisions new nodes at Cherry Servers and bootstraps them in the cluster.
func (mgr *cherryManagerRest) createNodes(nodegroup string, nodes int) error {
	klog.Infof("Updating node count to %d for nodegroup %s", nodes, nodegroup)

	cloudinit, err := base64.StdEncoding.DecodeString(mgr.getNodePoolDefinition(nodegroup).cloudinit)
	if err != nil {
		err = fmt.Errorf("could not decode cloudinit script: %w", err)
		klog.Fatal(err)
		return err
	}

	errList := make([]error, 0, nodes)
	for i := 0; i < nodes; i++ {
		errList = append(errList, mgr.createNode(context.TODO(), string(cloudinit), nodegroup))
	}

	return utilerrors.NewAggregate(errList)
}

func (mgr *cherryManagerRest) createServerRequest(ctx context.Context, cr *CreateServer, nodegroup string) error {
	req := path.Join("projects", fmt.Sprintf("%d", cr.ProjectID), "servers")

	jsonValue, err := json.Marshal(cr)
	if err != nil {
		return fmt.Errorf("failed to marshal create request: %w", err)
	}

	klog.Infof("Creating new node")
	if _, err := mgr.request(ctx, "POST", req, jsonValue); err != nil {
		return err
	}

	return nil
}

// getNodes should return ProviderIDs for all nodes in the node group,
// used to find any nodes which are unregistered in kubernetes.
func (mgr *cherryManagerRest) getNodes(nodegroup string) ([]string, error) {
	// Get node ProviderIDs by getting server IDs from Cherry Servers
	servers, err := mgr.listCherryServers(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	nodes := []string{}

	for _, s := range servers {
		clusterName, ok := s.Tags["k8s-cluster"]
		if !ok || clusterName != mgr.getNodePoolDefinition(nodegroup).clusterName {
			continue
		}
		nodepoolName, ok := s.Tags["k8s-nodepool"]
		if !ok || nodegroup != nodepoolName {
			continue
		}
		nodes = append(nodes, fmt.Sprintf("%s%d", cherryPrefix, s.ID))
	}

	return nodes, nil
}

// getNodeNames should return Names for all nodes in the node group,
// used to find any nodes which are unregistered in kubernetes.
func (mgr *cherryManagerRest) getNodeNames(nodegroup string) ([]string, error) {
	servers, err := mgr.listCherryServers(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	nodes := []string{}

	for _, s := range servers {
		clusterName, ok := s.Tags["k8s-cluster"]
		if !ok || clusterName != mgr.getNodePoolDefinition(nodegroup).clusterName {
			continue
		}
		nodepoolName, ok := s.Tags["k8s-nodepool"]
		if !ok || nodegroup != nodepoolName {
			continue
		}
		nodes = append(nodes, s.Hostname)
	}

	return nodes, nil
}

func (mgr *cherryManagerRest) deleteServer(ctx context.Context, nodegroup string, id int) error {
	req := path.Join("servers", fmt.Sprintf("%d", id))

	result, err := mgr.request(context.TODO(), "DELETE", req, []byte(""))
	if err != nil {
		return err
	}

	klog.Infof("Deleted server %d: %v", id, result)
	return nil

}

// deleteNodes deletes nodes by passing a comma separated list of names or IPs
func (mgr *cherryManagerRest) deleteNodes(nodegroup string, nodes []NodeRef, updatedNodeCount int) error {
	klog.Infof("Deleting %d nodes from nodegroup %s", len(nodes), nodegroup)
	klog.V(2).Infof("Deleting nodes %v", nodes)

	ctx := context.TODO()

	errList := make([]error, 0, len(nodes))

	servers, err := mgr.listCherryServers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}
	klog.V(2).Infof("total servers found: %d", len(servers))

	for _, n := range nodes {
		fakeNode := false

		if n.Name == n.ProviderID {
			klog.Infof("Fake Node: %s", n.Name)
			fakeNode = true
		} else {
			klog.Infof("Node %s - %s - %s", n.Name, n.MachineID, n.IPs)
		}

		// Get the count of servers tagged as nodegroup
		for _, s := range servers {
			klog.V(2).Infof("Checking server %v", s)
			clusterName, ok := s.Tags["k8s-cluster"]
			if !ok || clusterName != mgr.getNodePoolDefinition(nodegroup).clusterName {
				continue
			}
			nodepoolName, ok := s.Tags["k8s-nodepool"]
			if !ok || nodegroup != nodepoolName {
				continue
			}
			klog.V(2).Infof("nodegroup match %s %s", s.Hostname, n.Name)

			trimmedProviderID := strings.TrimPrefix(n.ProviderID, cherryPrefix)
			nodeID, err := strconv.ParseInt(trimmedProviderID, 10, 32)
			if err != nil {
				errList = append(errList, fmt.Errorf("invalid node ID is not integer for %s", n.Name))
			}

			switch {
			case s.Hostname == n.Name:
				klog.V(1).Infof("Matching Cherry Server %s - %d", s.Hostname, s.ID)
				errList = append(errList, mgr.deleteServer(ctx, nodegroup, s.ID))
			case fakeNode && int(nodeID) == s.ID:
				klog.V(1).Infof("Fake Node %d", s.ID)
				errList = append(errList, mgr.deleteServer(ctx, nodegroup, s.ID))
			}
		}
	}

	return utilerrors.NewAggregate(errList)
}

// BuildGenericLabels builds basic labels for Cherry Servers nodes
func BuildGenericLabels(nodegroup string, plan *Plan) map[string]string {
	result := make(map[string]string)

	result[apiv1.LabelInstanceType] = plan.Name
	//result[apiv1.LabelZoneRegion] = ""
	//result[apiv1.LabelZoneFailureDomain] = "0"
	//result[apiv1.LabelHostname] = ""
	result["pool"] = nodegroup

	return result
}

// templateNodeInfo returns a NodeInfo with a node template based on the Cherry Servers plan
// that is used to create nodes in a given node group.
func (mgr *cherryManagerRest) templateNodeInfo(nodegroup string) (*framework.NodeInfo, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-asg-%d", nodegroup, rand.Int63())
	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}
	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	// check if we need to update our plans
	if time.Since(mgr.planUpdate) > time.Hour*1 {
		plans, err := mgr.listCherryPlans(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("unable to update cherry plans: %v", err)
		}
		mgr.plans = map[string]*Plan{}
		for _, plan := range plans {
			mgr.plans[plan.Slug] = &plan
		}
	}
	planSlug := mgr.getNodePoolDefinition(nodegroup).plan
	cherryPlan, ok := mgr.plans[planSlug]
	if !ok {
		klog.V(5).Infof("no plan found for planSlug %s", planSlug)
		return nil, fmt.Errorf("cherry plan %q not supported", mgr.getNodePoolDefinition(nodegroup).plan)
	}
	var (
		memoryMultiplier int64
	)
	if memoryMultiplier, ok = multipliers[cherryPlan.Specs.Memory.Unit]; !ok {
		memoryMultiplier = 1
	}
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(int64(cherryPlan.Specs.Cpus.Cores), resource.DecimalSI)
	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(0, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(int64(cherryPlan.Specs.Memory.Total)*memoryMultiplier, resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity
	node.Status.Conditions = cloudprovider.BuildReadyConditions()

	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, BuildGenericLabels(nodegroup, cherryPlan))

	nodeInfo := framework.NewNodeInfo(&node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(nodegroup)})
	return nodeInfo, nil
}

func (mgr *cherryManagerRest) getNodePoolDefinition(nodegroup string) *cherryManagerNodePool {
	NodePoolDefinition, ok := mgr.nodePools[nodegroup]
	if !ok {
		NodePoolDefinition, ok = mgr.nodePools["default"]
		if !ok {
			klog.Fatalf("No default cloud-config was found")
		}
		klog.V(1).Infof("No cloud-config was found for %s, using default", nodegroup)
	}

	return NodePoolDefinition
}

func renderTemplate(str string, vars interface{}) (string, error) {
	tmpl, err := template.New("tmpl").Parse(str)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %q, %w", str, err)
	}

	var tmplBytes bytes.Buffer

	if err := tmpl.Execute(&tmplBytes, vars); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return tmplBytes.String(), nil
}
