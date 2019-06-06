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
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/packethost/packngo"
	"gopkg.in/gcfg.v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

type packetManagerPackngo struct {
	clusterName       string
	projectID         string
	apiServerEndpoint string
	facility          string
	plan              string
	os                string
	billing           string
	cloudinit         string
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
}

// ConfigFile is used to read and store information from the cloud configuration file
type ConfigFile struct {
	Global ConfigGlobal `gcfg:"global"`
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

// createPacketManagerPackngo sets up the client and returns
// an packetManagerPackngo.
func createPacketManagerPackngo(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (*packetManagerPackngo, error) {
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

	manager := packetManagerPackngo{
		clusterName:       cfg.Global.ClusterName,
		projectID:         cfg.Global.ProjectID,
		apiServerEndpoint: cfg.Global.APIServerEndpoint,
		facility:          cfg.Global.Facility,
		plan:              cfg.Global.Plan,
		os:                cfg.Global.OS,
		billing:           cfg.Global.Billing,
		cloudinit:         cfg.Global.CloudInit,
	}
	return &manager, nil
}

// nodeGroupSize gets the current size of the nodegroup as reported by packet tags.
func (mgr *packetManagerPackngo) nodeGroupSize(nodegroup string) (int, error) {
	c, err := packngo.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	dl, _, err := c.Devices.List(mgr.projectID, nil)
	if err != nil {
		klog.Fatalf("Error listing nodes: %v", err)
	}

	// Get the count of devices tagged as nodegroup
	count := 0
	for _, d := range dl {
		if Contains(d.Tags, "k8s-cluster-"+mgr.clusterName) && Contains(d.Tags, "k8s-nodepool-"+nodegroup) {
			count++
		}
	}
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

// createNodes provisions new nodes on packet and bootstraps them in the cluster.
func (mgr *packetManagerPackngo) createNodes(nodegroup string, nodes int) error {
	klog.Infof("Updating node count to %d for nodegroup %s", nodes, nodegroup)
	c, err := packngo.NewClient()
	if err != nil {
		log.Fatal(err)
		return fmt.Errorf("Could not authenticate with Packet: %v", err)
	}
	bootstrapTokenID := os.Getenv("BOOTSTRAP_TOKEN_ID")
	bootstrapTokenSecret := os.Getenv("BOOTSTRAP_TOKEN_SECRET")

	cloudinit, err := base64.StdEncoding.DecodeString(mgr.cloudinit)
	if err != nil {
		log.Fatal(err)
		return fmt.Errorf("Could not decode cloudinit script: %v", err)
	}

	ud := string(cloudinit) + "\nkubeadm join --discovery-token-unsafe-skip-ca-verification --token " + bootstrapTokenID + "." + bootstrapTokenSecret + " " + mgr.apiServerEndpoint
	hn := "k8s-" + mgr.clusterName + "-" + nodegroup + "-" + randString8()

	cr := packngo.DeviceCreateRequest{
		Hostname:     hn,
		Facility:     []string{mgr.facility},
		Plan:         mgr.plan,
		OS:           mgr.os,
		ProjectID:    mgr.projectID,
		BillingCycle: mgr.billing,
		UserData:     ud,
		Tags:         []string{"k8s-cluster-" + mgr.clusterName, "k8s-nodepool-" + nodegroup},
	}

	d, _, err := c.Devices.Create(&cr)
	if err != nil {
		klog.Fatalf("Error creating node: %v", err)
	}
	klog.Infof("Created device %v", d)

	return nil
}

// getNodes should return ProviderIDs for all nodes in the node group,
// used to find any nodes which are unregistered in kubernetes.
// This can not be done with packngo currently but a change has been merged upstream
// that will allow this.
func (mgr *packetManagerPackngo) getNodes(nodegroup string) ([]string, error) {
	// TODO: get node ProviderIDs by getting device IDs from packngo
	// This works fine being empty for now anyway.
	return []string{}, nil
}

// deleteNodes deletes nodes by passing a comma separated list of names or IPs
// of minions to remove to packngo, and simultaneously sets the new number of minions on the stack.
// The packet node_count is then set to the new value (does not cause any more nodes to be removed).
//
// TODO: The two step process is required until https://storyboard.openstack.org/#!/story/2005052
// is complete, which will allow resizing with specific nodes to be deleted as a single Packet operation.
func (mgr *packetManagerPackngo) deleteNodes(nodegroup string, nodes []NodeRef, updatedNodeCount int) error {
	klog.Infof("Deleting nodes")
	c, err := packngo.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range nodes {
		klog.Infof("Node %s - %s - %s", n.Name, n.MachineID, n.IPs)
		dl, _, err := c.Devices.List(mgr.projectID, nil)
		if err != nil {
			klog.Fatalf("Error listing nodes: %v", err)
		}
		// Get the count of devices tagged as nodegroup
		for _, d := range dl {
			if Contains(d.Tags, "k8s-cluster-"+mgr.clusterName) && Contains(d.Tags, "k8s-nodepool-"+nodegroup) {
				if d.Hostname == n.Name {
					klog.Infof("Matching Packet Device %s - %s - %v", d.Hostname, d.ID, d.Network)
					/*newTags := []string{}
					ur := packngo.DeviceUpdateRequest{Tags: &newTags}
					_, _, err := c.Devices.Update(d.ID, &ur) //Delete(d.ID)
					if err != nil {
						klog.Fatalf("Error deleting node: %v", err)
					}
					klog.Infof("Deleted Packet Device %s - %s", d.Hostname, d.ID)*/
				}
			}
		}
	}
	return nil
}

// templateNodeInfo returns a NodeInfo with a node template based on the VM flavor
// that is used to created minions in a given node group.
func (mgr *packetManagerPackngo) templateNodeInfo(nodegroup string) (*schedulernodeinfo.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}
