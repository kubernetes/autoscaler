package proxmox

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	goproxmox "github.com/luthermonson/go-proxmox"
)

const (
	taskIntervalDuration time.Duration = 5 * time.Second
	taskMaxDuration      time.Duration = 10 * time.Minute
)

type Client struct {
	poolName   string
	httpClient *http.Client
	px         *goproxmox.Client
}

func NewClient(apiEndpoint, username, password, tokenID, tokenSecret string, insecureSkipTLSVerify bool, poolName string) (*Client, error) {
	var client *goproxmox.Client
	var clientOptions []goproxmox.Option
	httpClient := &http.Client{}
	// Configure TLS if needed
	if insecureSkipTLSVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	clientOptions = append(clientOptions, goproxmox.WithHTTPClient(httpClient))
	if tokenID != "" && tokenSecret != "" {
		// Use API token authentication
		clientOptions = append(clientOptions, goproxmox.WithAPIToken(tokenID, tokenSecret))
		client = goproxmox.NewClient(apiEndpoint, clientOptions...)
	} else if username != "" && password != "" {
		// Use username/password authentication
		creds := &goproxmox.Credentials{
			Username: username,
			Password: password,
		}
		clientOptions = append(clientOptions, goproxmox.WithCredentials(creds))
		client = goproxmox.NewClient(apiEndpoint, clientOptions...)
	} else {
		return nil, fmt.Errorf("either API token (tokenID + tokenSecret) or credentials (username + password) must be provided")
	}

	return &Client{
		px:         client,
		httpClient: httpClient,
		poolName:   poolName,
	}, nil
}

func (c *Client) CreatePool(ctx context.Context, name string) error {
	pool, err := c.px.Pool(ctx, name)
	if err != nil {
		return err
	}

	if pool != nil {
		return nil
	}

	return c.px.NewPool(ctx, name, "created by cluster-autoscaler")
}

func (c *Client) GetNextFreeVMID(ctx context.Context) (int, error) {
	cluster, err := c.px.Cluster(ctx)
	if err != nil {
		return 0, err
	}

	return cluster.NextID(ctx)
}

type NodeUtilization struct {
	Name           string
	PercentageFree float32
}

func (c *Client) GetLeastUsedNode(ctx context.Context) (*ProxmoxNode, error) {
	nodes, err := c.px.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	nus := []NodeUtilization{}
	for _, node := range nodes {
		nus = append(nus, NodeUtilization{
			Name:           node.Name,
			PercentageFree: 1 - (float32(node.Mem) / float32(node.MaxMem)),
		})
	}

	if len(nus) == 0 {
		return nil, fmt.Errorf("no nodes exist in the proxmox cluster")
	}

	var max NodeUtilization = nus[0]
	for _, nu := range nus {
		if max.PercentageFree > nu.PercentageFree {
			max = nu
		}
	}

	var pxNode ProxmoxNode
	for _, node := range nodes {
		if node.Name == max.Name {
			pxNode.ID = node.Name
			pxNode.Online = node.Online == 1
			pxNode.Status = node.Status
		}
	}

	return &pxNode, nil
}

func (c *Client) CreateVM(ctx context.Context, nodeID string, templateID int, vmID int, config VMConfig, nodeGroup string) error {
	node, err := c.px.Node(ctx, nodeID)
	if err != nil {
		return err
	}

	template, err := node.VirtualMachine(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to get template VM %d: %v", templateID, err)
	}

	if !bool(template.Template) {
		return fmt.Errorf("VM %d is not marked as a template", templateID)
	}
	name := fmt.Sprintf("talos-worker-%d", vmID)
	// Clone the template to create a new VM
	cloneOptions := &goproxmox.VirtualMachineCloneOptions{
		NewID: vmID,
		Full:  1,
		Name:  name,
	}

	_, task, err := template.Clone(ctx, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone template %d to VM %d: %v", templateID, vmID, err)
	}

	if err := task.Wait(ctx, taskIntervalDuration, taskMaxDuration); err != nil {
		return fmt.Errorf("failed to wait for clone task: %v", err)
	}

	// Get the newly created VM to configure it
	newVM, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return fmt.Errorf("failed to get newly created VM %d: %v", vmID, err)
	}

	// Add node group tag
	allOpts := []goproxmox.VirtualMachineOption{}
	_nodeGroup := fmt.Sprintf("node-group_%s", nodeGroup)
	tagsOpts := goproxmox.VirtualMachineOption{
		Name:  "tags",
		Value: strings.Join(append(config.Tags, _nodeGroup), ","),
	}
	allOpts = append(allOpts, tagsOpts)

	task, err = newVM.Config(ctx, allOpts...)
	if err != nil {
		return err
	}
	if err := task.Wait(ctx, taskIntervalDuration, taskMaxDuration); err != nil {
		return fmt.Errorf("failed to set tags for vm %d: %v", vmID, err)
	}

	task, err = newVM.Start(ctx)
	if err != nil {
		return err
	}

	if err := task.Wait(ctx, taskIntervalDuration, taskMaxDuration); err != nil {
		return fmt.Errorf("failed to start vm %d: %v", vmID, err)
	}

	return nil
}

func (c *Client) DeleteVM(ctx context.Context, nodeID string, vmID int) error {
	node, err := c.px.Node(ctx, nodeID)
	if err != nil {
		return err
	}

	vm, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return err
	}

	task, err := vm.Delete(ctx)
	if err != nil {
		return err
	}

	if err := task.Wait(ctx, taskIntervalDuration, taskMaxDuration); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetVM(ctx context.Context, nodeName string, vmID int) (*VM, error) {
	node, err := c.px.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}

	vm, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return nil, err
	}

	uuid := extractUUID(vm.VirtualMachineConfig)

	return &VM{
		ID:     int(vm.VMID),
		Name:   vm.Name,
		Status: vm.Status,
		Node:   vm.Node,
		Tags:   vm.Tags,
		UUID:   uuid,
	}, nil
}

func (c *Client) GetVMs(ctx context.Context, nodeName string) ([]VM, error) {
	node, err := c.px.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}

	vms, err := node.VirtualMachines(ctx)
	if err != nil {
		return nil, err
	}

	_vms := []VM{}
	for _, _vm := range vms {
		vm, err := node.VirtualMachine(ctx, int(_vm.VMID))
		if err != nil {
			return nil, err
		}
		tags := parseVMTags(vm.Tags)
		_, ok := tags["node-group"]

		if bool(vm.Template) || !ok {
			continue
		}

		uuid := extractUUID(vm.VirtualMachineConfig)
		_vms = append(_vms, VM{
			ID:     int(vm.VMID),
			Name:   vm.Name,
			Status: vm.Status,
			Node:   vm.Node,
			Tags:   vm.Tags,
			UUID:   uuid,
		})
	}

	return _vms, nil
}

func (c *Client) GetNodes(ctx context.Context) ([]ProxmoxNode, error) {
	nodes, err := c.px.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	pxNodes := []ProxmoxNode{}
	for _, node := range nodes {
		pxNodes = append(pxNodes, ProxmoxNode{
			ID:     node.ID,
			Status: node.Status,
			Online: node.Online == 1,
		})
	}

	return pxNodes, nil
}

func (c *Client) StartVM(ctx context.Context, nodeID string, vmID int) error {
	node, err := c.px.Node(ctx, nodeID)
	if err != nil {
		return err
	}

	vm, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return err
	}

	// TODO: Check if already running if so no-op
	task, err := vm.Start(ctx)
	if err != nil {
		return err
	}

	return task.Wait(ctx, taskIntervalDuration, taskMaxDuration)
}

func (c *Client) StopVM(ctx context.Context, nodeID string, vmID int) error {
	node, err := c.px.Node(ctx, nodeID)
	if err != nil {
		return err
	}

	vm, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return err
	}

	// TODO: Check if already shutdown
	task, err := vm.Shutdown(ctx)
	if err != nil {
		return err
	}

	return task.Wait(ctx, taskIntervalDuration, taskMaxDuration)
}
