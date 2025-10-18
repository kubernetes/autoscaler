package proxmox

import (
	"context"
	"fmt"
	"strings"
	"time"

	goproxmox "github.com/luthermonson/go-proxmox"
)

const (
	taskIntervalDuration time.Duration = 5 * time.Second
	taskMaxDuration      time.Duration = 10 * time.Minute
)

type Client struct {
	pmoxClient *goproxmox.Client
}

func NewClient(apiEndpoint, username, password, tokenID, tokenSecret string) *Client {
	return nil
}

func (c *Client) CreateVM(ctx context.Context, nodeID string, templateID int, vmID int, config VMConfig) error {
	node, err := c.pmoxClient.Node(ctx, nodeID)
	if err != nil {
		return err
	}

	templ, err := node.VirtualMachine(ctx, templateID)
	if !bool(templ.Template) {
		return fmt.Errorf("template VM %d is not marked as a template", templateID)
	}

	opts := []goproxmox.VirtualMachineOption{
		{
			Name:  "cores",
			Value: config.Cores,
		},
		{
			Name:  "memory",
			Value: config.Memory,
		},
		{
			Name:  "tags",
			Value: strings.Join(config.Tags, ","),
		},
	}

	task, err := node.NewVirtualMachine(ctx, vmID, opts...)
	if err != nil {
		return err
	}

	if err := task.Wait(ctx, time.Second*5, time.Minute*10); err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteVM(ctx context.Context, nodeID string, vmID int) error {
	node, err := c.pmoxClient.Node(ctx, nodeID)
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

func (c *Client) GetVMs(ctx context.Context, nodeID string) ([]VM, error) {
	node, err := c.pmoxClient.Node(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	vms, err := node.VirtualMachines(ctx)
	if err != nil {
		return nil, err
	}

	_vms := []VM{}
	for _, vm := range vms {
		_vms = append(_vms, VM{
			ID:     int(vm.VMID),
			Name:   vm.Name,
			Status: vm.Status,
			Node:   vm.Node,
			Tags:   vm.Tags,
		})
	}

	return _vms, nil
}

func (c *Client) GetNodes(ctx context.Context) ([]ProxmoxNode, error) {
	nodes, err := c.pmoxClient.Nodes(ctx)
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
	node, err := c.pmoxClient.Node(ctx, nodeID)
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
	node, err := c.pmoxClient.Node(ctx, nodeID)
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
