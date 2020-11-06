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

package packet

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
	"os"
	"path"
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
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	userAgent                    = "kubernetes/cluster-autoscaler/" + version.ClusterAutoscalerVersion
	expectedAPIContentTypePrefix = "application/json"
)

type instanceType struct {
	InstanceName string
	CPU          int64
	MemoryMb     int64
	GPU          int64
}

// InstanceTypes is a map of packet resources
var InstanceTypes = map[string]*instanceType{
	"c1.large.arm": {
		InstanceName: "c1.large.arm",
		CPU:          96,
		MemoryMb:     131072,
		GPU:          0,
	},
	"c1.small.x86": {
		InstanceName: "c1.small.x86",
		CPU:          4,
		MemoryMb:     32768,
		GPU:          0,
	},
	"c1.xlarge.x86": {
		InstanceName: "c1.xlarge.x86",
		CPU:          16,
		MemoryMb:     131072,
		GPU:          0,
	},
	"c2.large.arm": {
		InstanceName: "c2.large.arm",
		CPU:          32,
		MemoryMb:     131072,
		GPU:          0,
	},
	"c2.medium.x86": {
		InstanceName: "c2.medium.x86",
		CPU:          24,
		MemoryMb:     65536,
		GPU:          0,
	},
	"c3.medium.x86": {
		InstanceName: "c3.medium.x86",
		CPU:          24,
		MemoryMb:     65536,
		GPU:          0,
	},
	"c3.small.x86": {
		InstanceName: "c3.small.x86",
		CPU:          8,
		MemoryMb:     32768,
		GPU:          1,
	},
	"g2.large.x86": {
		InstanceName: "g2.large.x86",
		CPU:          24,
		MemoryMb:     196608,
		GPU:          0,
	},
	"m1.xlarge.x86": {
		InstanceName: "m1.xlarge.x86",
		CPU:          24,
		MemoryMb:     262144,
		GPU:          0,
	},
	"m2.xlarge.x86": {
		InstanceName: "m2.xlarge.x86",
		CPU:          28,
		MemoryMb:     393216,
		GPU:          0,
	},
	"n2.xlarge.x86": {
		InstanceName: "n2.xlarge.x86",
		CPU:          28,
		MemoryMb:     393216,
		GPU:          0,
	},
	"s1.large.x86": {
		InstanceName: "s1.large.x86",
		CPU:          8,
		MemoryMb:     65536,
		GPU:          0,
	},
	"s3.xlarge.x86": {
		InstanceName: "s3.xlarge.x86",
		CPU:          24,
		MemoryMb:     196608,
		GPU:          0,
	},
	"t1.small.x86": {
		InstanceName: "t1.small.x86",
		CPU:          4,
		MemoryMb:     8192,
		GPU:          0,
	},
	"t3.small.x86": {
		InstanceName: "t3.small.x86",
		CPU:          4,
		MemoryMb:     16384,
		GPU:          0,
	},
	"x1.small.x86": {
		InstanceName: "x1.small.x86",
		CPU:          4,
		MemoryMb:     32768,
		GPU:          0,
	},
	"x2.xlarge.x86": {
		InstanceName: "x2.xlarge.x86",
		CPU:          28,
		MemoryMb:     393216,
		GPU:          1,
	},
}

type packetManagerNodePool struct {
	baseURL           string
	clusterName       string
	projectID         string
	apiServerEndpoint string
	facility          string
	plan              string
	os                string
	billing           string
	cloudinit         string
	reservation       string
	hostnamePattern   string
	waitTimeStep      time.Duration
}

type packetManagerRest struct {
	authToken              string
	packetManagerNodePools map[string]*packetManagerNodePool
}

// ConfigNodepool options only include the project-id for now
type ConfigNodepool struct {
	ClusterName       string `gcfg:"cluster-name"`
	ProjectID         string `gcfg:"project-id"`
	APIServerEndpoint string `gcfg:"api-server-endpoint"`
	Facility          string `gcfg:"facility"`
	Plan              string `gcfg:"plan"`
	OS                string `gcfg:"os"`
	Billing           string `gcfg:"billing"`
	CloudInit         string `gcfg:"cloudinit"`
	Reservation       string `gcfg:"reservation"`
	HostnamePattern   string `gcfg:"hostname-pattern"`
}

// ConfigFile is used to read and store information from the cloud configuration file
type ConfigFile struct {
	DefaultNodegroupdef ConfigNodepool             `gcfg:"global"`
	Nodegroupdef        map[string]*ConfigNodepool `gcfg:"nodegroupdef"`
}

// Device represents a Packet device
type Device struct {
	ID          string   `json:"id"`
	ShortID     string   `json:"short_id"`
	Hostname    string   `json:"hostname"`
	Description string   `json:"description"`
	State       string   `json:"state"`
	Tags        []string `json:"tags"`
}

// Devices represents a list of Packet devices
type Devices struct {
	Devices []Device `json:"devices"`
}

// IPAddressCreateRequest represents a request to create a new IP address within a DeviceCreateRequest
type IPAddressCreateRequest struct {
	AddressFamily int  `json:"address_family"`
	Public        bool `json:"public"`
}

// DeviceCreateRequest represents a request to create a new Packet device. Used by createNodes
type DeviceCreateRequest struct {
	Hostname              string                   `json:"hostname"`
	Plan                  string                   `json:"plan"`
	Facility              []string                 `json:"facility"`
	OS                    string                   `json:"operating_system"`
	BillingCycle          string                   `json:"billing_cycle"`
	ProjectID             string                   `json:"project_id"`
	UserData              string                   `json:"userdata"`
	Storage               string                   `json:"storage,omitempty"`
	Tags                  []string                 `json:"tags"`
	CustomData            string                   `json:"customdata,omitempty"`
	IPAddresses           []IPAddressCreateRequest `json:"ip_addresses,omitempty"`
	HardwareReservationID string                   `json:"hardware_reservation_id,omitempty"`
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

// createPacketManagerRest sets up the client and returns
// an packetManagerRest.
func createPacketManagerRest(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (*packetManagerRest, error) {
	var cfg ConfigFile
	if configReader != nil {
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			klog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	}

	var manager packetManagerRest
	manager.packetManagerNodePools = make(map[string]*packetManagerNodePool)

	if _, ok := cfg.Nodegroupdef["default"]; !ok {
		cfg.Nodegroupdef["default"] = &cfg.DefaultNodegroupdef
	}

	if *cfg.Nodegroupdef["default"] == (ConfigNodepool{}) {
		klog.Fatalf("No \"default\" or [Global] nodepool definition was found")
	}

	packetAuthToken := os.Getenv("PACKET_AUTH_TOKEN")
	if len(packetAuthToken) == 0 {
		klog.Fatalf("PACKET_AUTH_TOKEN is required and missing")
	}

	manager.authToken = packetAuthToken

	for nodepool := range cfg.Nodegroupdef {
		if opts.ClusterName == "" && cfg.Nodegroupdef[nodepool].ClusterName == "" {
			klog.Fatalf("The cluster-name parameter must be set")
		} else if opts.ClusterName != "" && cfg.Nodegroupdef[nodepool].ClusterName == "" {
			cfg.Nodegroupdef[nodepool].ClusterName = opts.ClusterName
		}

		manager.packetManagerNodePools[nodepool] = &packetManagerNodePool{
			baseURL:           "https://api.packet.net",
			clusterName:       cfg.Nodegroupdef[nodepool].ClusterName,
			projectID:         cfg.Nodegroupdef["default"].ProjectID,
			apiServerEndpoint: cfg.Nodegroupdef["default"].APIServerEndpoint,
			facility:          cfg.Nodegroupdef[nodepool].Facility,
			plan:              cfg.Nodegroupdef[nodepool].Plan,
			os:                cfg.Nodegroupdef[nodepool].OS,
			billing:           cfg.Nodegroupdef[nodepool].Billing,
			cloudinit:         cfg.Nodegroupdef[nodepool].CloudInit,
			reservation:       cfg.Nodegroupdef[nodepool].Reservation,
			hostnamePattern:   cfg.Nodegroupdef[nodepool].HostnamePattern,
		}
	}

	return &manager, nil
}

func (mgr *packetManagerRest) request(ctx context.Context, method, url string, jsonData []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Auth-Token", mgr.authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

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

func (mgr *packetManagerRest) listPacketDevices(ctx context.Context) (*Devices, error) {
	url := mgr.getNodePoolDefinition("default").baseURL + "/" + path.Join("projects", mgr.getNodePoolDefinition("default").projectID, "devices")
	klog.Infof("url: %v", url)

	result, err := mgr.request(ctx, "GET", url, []byte(``))
	if err != nil {
		return nil, err
	}

	var devices Devices
	if err := json.Unmarshal(result, &devices); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &devices, nil
}

func (mgr *packetManagerRest) getPacketDevice(ctx context.Context, id string) (*Device, error) {
	url := mgr.getNodePoolDefinition("default").baseURL + "/" + path.Join("devices", id)

	result, err := mgr.request(ctx, "GET", url, []byte(``))
	if err != nil {
		return nil, err
	}

	var device Device
	if err := json.Unmarshal(result, &device); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &device, nil
}

func (mgr *packetManagerRest) NodeGroupForNode(labels map[string]string, nodeId string) (string, error) {
	if nodegroup, ok := labels["pool"]; ok {
		return nodegroup, nil
	}
	device, err := mgr.getPacketDevice(context.TODO(), strings.TrimPrefix(nodeId, "packet://"))
	if err != nil {
		return "", fmt.Errorf("Could not find group for node: %s %s", nodeId, err)
	}
	for _, t := range device.Tags {
		if strings.HasPrefix(t, "k8s-nodepool-") {
			return strings.TrimPrefix(t, "k8s-nodepool-"), nil
		}
	}
	return "", fmt.Errorf("Could not find group for node: %s", nodeId)
}

// nodeGroupSize gets the current size of the nodegroup as reported by packet tags.
func (mgr *packetManagerRest) nodeGroupSize(nodegroup string) (int, error) {
	devices, err := mgr.listPacketDevices(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("failed to list devices: %w", err)
	}

	// Get the count of devices tagged as nodegroup members
	count := 0
	for _, d := range devices.Devices {
		if Contains(d.Tags, "k8s-cluster-"+mgr.getNodePoolDefinition(nodegroup).clusterName) && Contains(d.Tags, "k8s-nodepool-"+nodegroup) {
			count++
		}
	}
	klog.V(3).Infof("Nodegroup %s: %d/%d", nodegroup, count, len(devices.Devices))
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

func (mgr *packetManagerRest) createNode(ctx context.Context, cloudinit, nodegroup string) error {
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

	if err := mgr.createDevice(ctx, hn, ud, nodegroup); err != nil {
		return fmt.Errorf("failed to create device %q in node group %q: %w", hn, nodegroup, err)
	}

	klog.Infof("Created new node on Packet.")

	return nil
}

// createNodes provisions new nodes on packet and bootstraps them in the cluster.
func (mgr *packetManagerRest) createNodes(nodegroup string, nodes int) error {
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

func (mgr *packetManagerRest) createDevice(ctx context.Context, hostname, userData, nodegroup string) error {
	reservation := ""
	if mgr.getNodePoolDefinition(nodegroup).reservation == "require" || mgr.getNodePoolDefinition(nodegroup).reservation == "prefer" {
		reservation = "next-available"
	}

	cr := &DeviceCreateRequest{
		Hostname:              hostname,
		Facility:              []string{mgr.getNodePoolDefinition(nodegroup).facility},
		Plan:                  mgr.getNodePoolDefinition(nodegroup).plan,
		OS:                    mgr.getNodePoolDefinition(nodegroup).os,
		ProjectID:             mgr.getNodePoolDefinition(nodegroup).projectID,
		BillingCycle:          mgr.getNodePoolDefinition(nodegroup).billing,
		UserData:              userData,
		Tags:                  []string{"k8s-cluster-" + mgr.getNodePoolDefinition(nodegroup).clusterName, "k8s-nodepool-" + nodegroup},
		HardwareReservationID: reservation,
	}

	if err := mgr.createDeviceRequest(ctx, cr, nodegroup); err != nil {
		// If reservation is preferred but not available, retry provisioning as on-demand
		if reservation != "" && mgr.getNodePoolDefinition(nodegroup).reservation == "prefer" && isNoAvailableReservationsError(err) {
			klog.Infof("Reservation preferred but not available. Provisioning on-demand node.")

			cr.HardwareReservationID = ""
			return mgr.createDeviceRequest(ctx, cr, nodegroup)
		}

		return fmt.Errorf("failed to create device: %w", err)
	}

	return nil
}

// TODO: find a better way than parsing the error messages for this.
func isNoAvailableReservationsError(err error) bool {
	return strings.Contains(err.Error(), " no available hardware reservations ")
}

func (mgr *packetManagerRest) createDeviceRequest(ctx context.Context, cr *DeviceCreateRequest, nodegroup string) error {
	url := mgr.getNodePoolDefinition("default").baseURL + "/" + path.Join("projects", cr.ProjectID, "devices")

	jsonValue, err := json.Marshal(cr)
	if err != nil {
		return fmt.Errorf("failed to marshal create request: %w", err)
	}

	klog.Infof("Creating new node")
	klog.V(3).Infof("POST %s \n%v", url, string(jsonValue))

	if _, err := mgr.request(ctx, "POST", url, jsonValue); err != nil {
		return err
	}

	return nil
}

// getNodes should return ProviderIDs for all nodes in the node group,
// used to find any nodes which are unregistered in kubernetes.
func (mgr *packetManagerRest) getNodes(nodegroup string) ([]string, error) {
	// Get node ProviderIDs by getting device IDs from Packet
	devices, err := mgr.listPacketDevices(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	nodes := []string{}

	for _, d := range devices.Devices {
		if Contains(d.Tags, "k8s-cluster-"+mgr.getNodePoolDefinition(nodegroup).clusterName) && Contains(d.Tags, "k8s-nodepool-"+nodegroup) {
			nodes = append(nodes, fmt.Sprintf("packet://%s", d.ID))
		}
	}

	return nodes, nil
}

// getNodeNames should return Names for all nodes in the node group,
// used to find any nodes which are unregistered in kubernetes.
func (mgr *packetManagerRest) getNodeNames(nodegroup string) ([]string, error) {
	devices, err := mgr.listPacketDevices(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	nodes := []string{}

	for _, d := range devices.Devices {
		if Contains(d.Tags, "k8s-cluster-"+mgr.getNodePoolDefinition(nodegroup).clusterName) && Contains(d.Tags, "k8s-nodepool-"+nodegroup) {
			nodes = append(nodes, d.Hostname)
		}
	}

	return nodes, nil
}

func (mgr *packetManagerRest) deleteDevice(ctx context.Context, nodegroup, id string) error {
	url := mgr.getNodePoolDefinition("default").baseURL + "/" + path.Join("devices", id)

	result, err := mgr.request(context.TODO(), "DELETE", url, []byte(""))
	if err != nil {
		return err
	}

	klog.Infof("Deleted device %s: %v", id, result)
	return nil

}

// deleteNodes deletes nodes by passing a comma separated list of names or IPs
func (mgr *packetManagerRest) deleteNodes(nodegroup string, nodes []NodeRef, updatedNodeCount int) error {
	klog.Infof("Deleting nodes %v", nodes)

	ctx := context.TODO()

	errList := make([]error, 0, len(nodes))

	devices, err := mgr.listPacketDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}
	klog.Infof("%d devices total", len(devices.Devices))

	for _, n := range nodes {
		fakeNode := false

		if n.Name == n.ProviderID {
			klog.Infof("Fake Node: %s", n.Name)
			fakeNode = true
		} else {
			klog.Infof("Node %s - %s - %s", n.Name, n.MachineID, n.IPs)
		}

		// Get the count of devices tagged as nodegroup
		for _, d := range devices.Devices {
			klog.Infof("Checking device %v", d)
			if Contains(d.Tags, "k8s-cluster-"+mgr.getNodePoolDefinition(nodegroup).clusterName) && Contains(d.Tags, "k8s-nodepool-"+nodegroup) {
				klog.Infof("nodegroup match %s %s", d.Hostname, n.Name)
				switch {
				case d.Hostname == n.Name:
					klog.V(1).Infof("Matching Packet Device %s - %s", d.Hostname, d.ID)
					errList = append(errList, mgr.deleteDevice(ctx, nodegroup, d.ID))
				case fakeNode && strings.TrimPrefix(n.Name, "packet://") == d.ID:
					klog.V(1).Infof("Fake Node %s", d.ID)
					errList = append(errList, mgr.deleteDevice(ctx, nodegroup, d.ID))
				}
			}
		}
	}

	return utilerrors.NewAggregate(errList)
}

// BuildGenericLabels builds basic labels for Packet nodes
func BuildGenericLabels(nodegroup string, instanceType string) map[string]string {
	result := make(map[string]string)

	//result[kubeletapis.LabelArch] = "amd64"
	//result[kubeletapis.LabelOS] = "linux"
	result[apiv1.LabelInstanceType] = instanceType
	//result[apiv1.LabelZoneRegion] = ""
	//result[apiv1.LabelZoneFailureDomain] = "0"
	//result[apiv1.LabelHostname] = ""
	result["pool"] = nodegroup

	return result
}

// templateNodeInfo returns a NodeInfo with a node template based on the packet plan
// that is used to create nodes in a given node group.
func (mgr *packetManagerRest) templateNodeInfo(nodegroup string) (*schedulerframework.NodeInfo, error) {
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

	packetPlan := InstanceTypes[mgr.getNodePoolDefinition(nodegroup).plan]
	if packetPlan == nil {
		return nil, fmt.Errorf("packet plan %q not supported", mgr.getNodePoolDefinition(nodegroup).plan)
	}
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(packetPlan.CPU, resource.DecimalSI)
	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(packetPlan.GPU, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(packetPlan.MemoryMb*1024*1024, resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity
	node.Status.Conditions = cloudprovider.BuildReadyConditions()

	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, BuildGenericLabels(nodegroup, mgr.getNodePoolDefinition(nodegroup).plan))

	nodeInfo := schedulerframework.NewNodeInfo(cloudprovider.BuildKubeProxy(nodegroup))
	nodeInfo.SetNode(&node)
	return nodeInfo, nil
}

func (mgr *packetManagerRest) getNodePoolDefinition(nodegroup string) *packetManagerNodePool {
	NodePoolDefinition, ok := mgr.packetManagerNodePools[nodegroup]
	if !ok {
		NodePoolDefinition, ok = mgr.packetManagerNodePools["default"]
		if !ok {
			klog.Fatalf("No default cloud-config was found")
		}
		klog.Infof("No cloud-config was found for %s, using default", nodegroup)
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
