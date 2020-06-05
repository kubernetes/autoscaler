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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"text/template"
	"time"

	"gopkg.in/gcfg.v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

type packetManagerRest struct {
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

// ConfigGlobal options only include the project-id for now
type ConfigGlobal struct {
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
	Global ConfigGlobal `gcfg:"global"`
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
}

// HostnameTemplateData represents the template variables used to construct host names for new nodes
type HostnameTemplateData struct {
	ClusterName string
	NodeGroup   string
	RandString8 string
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

	if opts.ClusterName == "" && cfg.Global.ClusterName == "" {
		klog.Fatalf("The cluster-name parameter must be set")
	} else if opts.ClusterName != "" && cfg.Global.ClusterName == "" {
		cfg.Global.ClusterName = opts.ClusterName
	}

	manager := packetManagerRest{
		baseURL:           "https://api.packet.net",
		clusterName:       cfg.Global.ClusterName,
		projectID:         cfg.Global.ProjectID,
		apiServerEndpoint: cfg.Global.APIServerEndpoint,
		facility:          cfg.Global.Facility,
		plan:              cfg.Global.Plan,
		os:                cfg.Global.OS,
		billing:           cfg.Global.Billing,
		cloudinit:         cfg.Global.CloudInit,
		reservation:       cfg.Global.Reservation,
		hostnamePattern:   cfg.Global.HostnamePattern,
	}
	return &manager, nil
}

func (mgr *packetManagerRest) listPacketDevices() (*Devices, error) {
	var jsonStr = []byte(``)
	packetAuthToken := os.Getenv("PACKET_AUTH_TOKEN")
	url := mgr.baseURL + "/projects/" + mgr.projectID + "/devices"
	req, _ := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Auth-Token", packetAuthToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
		// klog.Fatalf("Error listing nodes: %v", err)
	}
	defer resp.Body.Close()

	klog.Infof("response Status: %s", resp.Status)

	var devices Devices

	if "200 OK" == resp.Status {
		body, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal([]byte(body), &devices)
		return &devices, nil
	}

	return &devices, fmt.Errorf(resp.Status, resp.Body)
}

// nodeGroupSize gets the current size of the nodegroup as reported by packet tags.
func (mgr *packetManagerRest) nodeGroupSize(nodegroup string) (int, error) {
	devices, _ := mgr.listPacketDevices()
	// Get the count of devices tagged as nodegroup members
	count := 0
	for _, d := range devices.Devices {
		if Contains(d.Tags, "k8s-cluster-"+mgr.clusterName) && Contains(d.Tags, "k8s-nodepool-"+nodegroup) {
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

func (mgr *packetManagerRest) createNode(cloudinit, nodegroup string) {
	udvars := CloudInitTemplateData{
		BootstrapTokenID:     os.Getenv("BOOTSTRAP_TOKEN_ID"),
		BootstrapTokenSecret: os.Getenv("BOOTSTRAP_TOKEN_SECRET"),
		APIServerEndpoint:    mgr.apiServerEndpoint,
	}
	ud := renderTemplate(cloudinit, udvars)
	hnvars := HostnameTemplateData{
		ClusterName: mgr.clusterName,
		NodeGroup:   nodegroup,
		RandString8: randString8(),
	}
	hn := renderTemplate(mgr.hostnamePattern, hnvars)

	reservation := ""
	if mgr.reservation == "require" || mgr.reservation == "prefer" {
		reservation = "next-available"
	}

	cr := DeviceCreateRequest{
		Hostname:              hn,
		Facility:              []string{mgr.facility},
		Plan:                  mgr.plan,
		OS:                    mgr.os,
		ProjectID:             mgr.projectID,
		BillingCycle:          mgr.billing,
		UserData:              ud,
		Tags:                  []string{"k8s-cluster-" + mgr.clusterName, "k8s-nodepool-" + nodegroup},
		HardwareReservationID: reservation,
	}

	resp, err := createDevice(&cr, mgr.baseURL)
	if err != nil || resp.StatusCode > 299 {
		// If reservation is preferred but not available, retry provisioning as on-demand
		if reservation != "" && mgr.reservation == "prefer" {
			klog.Infof("Reservation preferred but not available. Provisioning on-demand node.")
			cr.HardwareReservationID = ""
			resp, err = createDevice(&cr, mgr.baseURL)
			if err != nil {
				klog.Errorf("Failed to create device using Packet API: %v", err)
				panic(err)
			} else if resp.StatusCode > 299 {
				klog.Errorf("Failed to create device using Packet API: %v", resp)
				panic(resp)
			}
		} else if err != nil {
			klog.Errorf("Failed to create device using Packet API: %v", err)
			panic(err)
		} else if resp.StatusCode > 299 {
			klog.Errorf("Failed to create device using Packet API: %v", resp)
			panic(resp)
		}
	}
	defer resp.Body.Close()

	rbody, err := ioutil.ReadAll(resp.Body)
	klog.V(3).Infof("Response body: %v", string(rbody))
	if err != nil {
		klog.Errorf("Failed to read response body: %v", err)
	} else {
		klog.Infof("Created new node on Packet.")
	}
	if cr.HardwareReservationID != "" {
		klog.Infof("Reservation %v", cr.HardwareReservationID)
	}
}

// createNodes provisions new nodes on packet and bootstraps them in the cluster.
func (mgr *packetManagerRest) createNodes(nodegroup string, nodes int) error {
	klog.Infof("Updating node count to %d for nodegroup %s", nodes, nodegroup)
	cloudinit, err := base64.StdEncoding.DecodeString(mgr.cloudinit)
	if err != nil {
		log.Fatal(err)
		return fmt.Errorf("Could not decode cloudinit script: %v", err)
	}

	for i := 0; i < nodes; i++ {
		mgr.createNode(string(cloudinit), nodegroup)
	}

	return nil
}

func createDevice(cr *DeviceCreateRequest, baseURL string) (*http.Response, error) {
	packetAuthToken := os.Getenv("PACKET_AUTH_TOKEN")
	url := baseURL + "/projects/" + cr.ProjectID + "/devices"
	jsonValue, _ := json.Marshal(cr)
	klog.Infof("Creating new node")
	klog.V(3).Infof("POST %s \n%v", url, string(jsonValue))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		klog.Errorf("Failed to create device: %v", err)
		panic(err)
	}
	req.Header.Set("X-Auth-Token", packetAuthToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	return resp, err
}

// getNodes should return ProviderIDs for all nodes in the node group,
// used to find any nodes which are unregistered in kubernetes.
func (mgr *packetManagerRest) getNodes(nodegroup string) ([]string, error) {
	// Get node ProviderIDs by getting device IDs from Packet
	devices, err := mgr.listPacketDevices()
	nodes := []string{}
	for _, d := range devices.Devices {
		if Contains(d.Tags, "k8s-cluster-"+mgr.clusterName) && Contains(d.Tags, "k8s-nodepool-"+nodegroup) {
			nodes = append(nodes, d.ID)
		}
	}
	return nodes, err
}

// getNodeNames should return Names for all nodes in the node group,
// used to find any nodes which are unregistered in kubernetes.
func (mgr *packetManagerRest) getNodeNames(nodegroup string) ([]string, error) {
	devices, err := mgr.listPacketDevices()
	nodes := []string{}
	for _, d := range devices.Devices {
		if Contains(d.Tags, "k8s-cluster-"+mgr.clusterName) && Contains(d.Tags, "k8s-nodepool-"+nodegroup) {
			nodes = append(nodes, d.Hostname)
		}
	}
	return nodes, err
}

// deleteNodes deletes nodes by passing a comma separated list of names or IPs
func (mgr *packetManagerRest) deleteNodes(nodegroup string, nodes []NodeRef, updatedNodeCount int) error {
	klog.Infof("Deleting nodes %v", nodes)
	packetAuthToken := os.Getenv("PACKET_AUTH_TOKEN")
	for _, n := range nodes {
		klog.Infof("Node %s - %s - %s", n.Name, n.MachineID, n.IPs)
		dl, _ := mgr.listPacketDevices()
		klog.Infof("%d devices total", len(dl.Devices))
		// Get the count of devices tagged as nodegroup
		for _, d := range dl.Devices {
			klog.Infof("Checking device %v", d)
			if Contains(d.Tags, "k8s-cluster-"+mgr.clusterName) && Contains(d.Tags, "k8s-nodepool-"+nodegroup) {
				klog.Infof("nodegroup match %s %s", d.Hostname, n.Name)
				if d.Hostname == n.Name {
					klog.V(1).Infof("Matching Packet Device %s - %s", d.Hostname, d.ID)
					req, _ := http.NewRequest("DELETE", mgr.baseURL+"/devices/"+d.ID, bytes.NewBuffer([]byte("")))
					req.Header.Set("X-Auth-Token", packetAuthToken)
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						klog.Fatalf("Error deleting node: %v", err)
						panic(err)
					}
					defer resp.Body.Close()
					body, _ := ioutil.ReadAll(resp.Body)
					klog.Infof("Deleted device %s: %v", d.ID, body)
				}
			}
		}
	}
	return nil
}

// templateNodeInfo returns a NodeInfo with a node template based on the VM flavor
// that is used to created minions in a given node group.
func (mgr *packetManagerRest) templateNodeInfo(nodegroup string) (*schedulerframework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func renderTemplate(str string, vars interface{}) string {
	tmpl, err := template.New("tmpl").Parse(str)

	if err != nil {
		panic(err)
	}
	var tmplBytes bytes.Buffer

	err = tmpl.Execute(&tmplBytes, vars)
	if err != nil {
		panic(err)
	}
	return tmplBytes.String()
}
