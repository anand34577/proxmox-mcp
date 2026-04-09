


package proxmox

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	pve "github.com/luthermonson/go-proxmox"
	"go.uber.org/zap"

	"github.com/anand34577/proxmox-mcp/internal/config"
)




type Client struct {
	pve    *pve.Client
	cfg    *config.ProxmoxConfig
	logger *zap.Logger
}



func New(ctx context.Context, cfg *config.ProxmoxConfig, log *zap.Logger) (*Client, error) {
	httpClient := &http.Client{
		Timeout: cfg.RequestTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.InsecureSkipVerify, 
			},
		},
	}

	opts := []pve.Option{
		pve.WithHTTPClient(httpClient),
		pve.WithAPIToken(cfg.TokenID, cfg.Secret),
	}

	c := pve.NewClient(cfg.BaseURL, opts...)

	
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	version, err := c.Version(ctx)
	if err != nil {
		return nil, fmt.Errorf("proxmox: connecting to %s: %w", cfg.BaseURL, err)
	}

	log.Info("connected to Proxmox VE",
		zap.String("url", cfg.BaseURL),
		zap.String("release", version.Release),
	)

	return &Client{pve: c, cfg: cfg, logger: log}, nil
}







func (c *Client) ClusterStatus(ctx context.Context) (*pve.Cluster, error) {
	cluster, err := c.pve.Cluster(ctx)
	if err != nil {
		return nil, fmt.Errorf("cluster status: %w", err)
	}
	return cluster, nil
}



func (c *Client) ClusterResources(ctx context.Context, resourceType string) ([]*pve.ClusterResource, error) {
	cluster, err := c.pve.Cluster(ctx)
	if err != nil {
		return nil, fmt.Errorf("cluster resources: %w", err)
	}
	return cluster.Resources(ctx, resourceType)
}


func (c *Client) Version(ctx context.Context) (*pve.Version, error) {
	v, err := c.pve.Version(ctx)
	if err != nil {
		return nil, fmt.Errorf("version: %w", err)
	}
	return v, nil
}




func (c *Client) Nodes(ctx context.Context) (pve.NodeStatuses, error) {
	nodes, err := c.pve.Nodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("nodes: %w", err)
	}
	return nodes, nil
}


func (c *Client) Node(ctx context.Context, name string) (*pve.Node, error) {
	node, err := c.pve.Node(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("node %q: %w", name, err)
	}
	return node, nil
}




func (c *Client) VirtualMachines(ctx context.Context, nodeName string) (pve.VirtualMachines, error) {
	node, err := c.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	vms, err := node.VirtualMachines(ctx)
	if err != nil {
		return nil, fmt.Errorf("vms on node %q: %w", nodeName, err)
	}
	return vms, nil
}


func (c *Client) VirtualMachine(ctx context.Context, nodeName string, vmid int) (*pve.VirtualMachine, error) {
	node, err := c.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	vm, err := node.VirtualMachine(ctx, vmid)
	if err != nil {
		return nil, fmt.Errorf("vm %d on node %q: %w", vmid, nodeName, err)
	}
	return vm, nil
}


func (c *Client) StartVM(ctx context.Context, nodeName string, vmid int) (*pve.Task, error) {
	vm, err := c.VirtualMachine(ctx, nodeName, vmid)
	if err != nil {
		return nil, err
	}
	task, err := vm.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("start vm %d: %w", vmid, err)
	}
	return task, nil
}


func (c *Client) StopVM(ctx context.Context, nodeName string, vmid int) (*pve.Task, error) {
	vm, err := c.VirtualMachine(ctx, nodeName, vmid)
	if err != nil {
		return nil, err
	}
	task, err := vm.Stop(ctx)
	if err != nil {
		return nil, fmt.Errorf("stop vm %d: %w", vmid, err)
	}
	return task, nil
}


func (c *Client) ShutdownVM(ctx context.Context, nodeName string, vmid int) (*pve.Task, error) {
	vm, err := c.VirtualMachine(ctx, nodeName, vmid)
	if err != nil {
		return nil, err
	}
	task, err := vm.Shutdown(ctx)
	if err != nil {
		return nil, fmt.Errorf("shutdown vm %d: %w", vmid, err)
	}
	return task, nil
}


func (c *Client) RebootVM(ctx context.Context, nodeName string, vmid int) (*pve.Task, error) {
	vm, err := c.VirtualMachine(ctx, nodeName, vmid)
	if err != nil {
		return nil, err
	}
	task, err := vm.Reboot(ctx)
	if err != nil {
		return nil, fmt.Errorf("reboot vm %d: %w", vmid, err)
	}
	return task, nil
}




func (c *Client) SuspendVM(ctx context.Context, nodeName string, vmid int) (*pve.Task, error) {
	vm, err := c.VirtualMachine(ctx, nodeName, vmid)
	if err != nil {
		return nil, err
	}
	task, err := vm.Hibernate(ctx)
	if err != nil {
		return nil, fmt.Errorf("suspend vm %d: %w", vmid, err)
	}
	return task, nil
}


func (c *Client) ResumeVM(ctx context.Context, nodeName string, vmid int) (*pve.Task, error) {
	vm, err := c.VirtualMachine(ctx, nodeName, vmid)
	if err != nil {
		return nil, err
	}
	task, err := vm.Resume(ctx)
	if err != nil {
		return nil, fmt.Errorf("resume vm %d: %w", vmid, err)
	}
	return task, nil
}


func (c *Client) DeleteVM(ctx context.Context, nodeName string, vmid int) (*pve.Task, error) {
	vm, err := c.VirtualMachine(ctx, nodeName, vmid)
	if err != nil {
		return nil, err
	}
	task, err := vm.Delete(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete vm %d: %w", vmid, err)
	}
	return task, nil
}





func (c *Client) CloneVM(ctx context.Context, nodeName string, vmid int, opts pve.VirtualMachineCloneOptions) (*pve.Task, int64, error) {
	vm, err := c.VirtualMachine(ctx, nodeName, vmid)
	if err != nil {
		return nil, 0, err
	}
	newID, task, err := vm.Clone(ctx, &opts)
	if err != nil {
		return nil, 0, fmt.Errorf("clone vm %d: %w", vmid, err)
	}
	return task, int64(newID), nil
}


func (c *Client) VMSnapshots(ctx context.Context, nodeName string, vmid int) ([]*pve.Snapshot, error) {
	vm, err := c.VirtualMachine(ctx, nodeName, vmid)
	if err != nil {
		return nil, err
	}
	snaps, err := vm.Snapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("snapshots for vm %d: %w", vmid, err)
	}
	return snaps, nil
}




func (c *Client) CreateVMSnapshot(ctx context.Context, nodeName string, vmid int, name, description string) (*pve.Task, error) {
	vm, err := c.VirtualMachine(ctx, nodeName, vmid)
	if err != nil {
		return nil, err
	}
	task, err := vm.NewSnapshot(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("create snapshot %q for vm %d: %w", name, vmid, err)
	}
	return task, nil
}



func (c *Client) RollbackVMSnapshot(ctx context.Context, nodeName string, vmid int, snapName string) (*pve.Task, error) {
	vm, err := c.VirtualMachine(ctx, nodeName, vmid)
	if err != nil {
		return nil, err
	}
	task, err := vm.SnapshotRollback(ctx, snapName)
	if err != nil {
		return nil, fmt.Errorf("rollback vm %d to snapshot %q: %w", vmid, snapName, err)
	}
	return task, nil
}




func (c *Client) DeleteVMSnapshot(ctx context.Context, nodeName string, vmid int, snapName string) (*pve.Task, error) {
	var upid pve.UPID
	path := fmt.Sprintf("/nodes/%s/qemu/%d/snapshot/%s", nodeName, vmid, snapName)
	if err := c.pve.Delete(ctx, path, &upid); err != nil {
		return nil, fmt.Errorf("delete snapshot %q from vm %d: %w", snapName, vmid, err)
	}
	return pve.NewTask(upid, c.pve), nil
}




func (c *Client) Containers(ctx context.Context, nodeName string) (pve.Containers, error) {
	node, err := c.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	cts, err := node.Containers(ctx)
	if err != nil {
		return nil, fmt.Errorf("containers on node %q: %w", nodeName, err)
	}
	return cts, nil
}


func (c *Client) Container(ctx context.Context, nodeName string, ctid int) (*pve.Container, error) {
	node, err := c.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	ct, err := node.Container(ctx, ctid)
	if err != nil {
		return nil, fmt.Errorf("container %d on node %q: %w", ctid, nodeName, err)
	}
	return ct, nil
}


func (c *Client) StartContainer(ctx context.Context, nodeName string, ctid int) (*pve.Task, error) {
	ct, err := c.Container(ctx, nodeName, ctid)
	if err != nil {
		return nil, err
	}
	task, err := ct.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("start container %d: %w", ctid, err)
	}
	return task, nil
}


func (c *Client) StopContainer(ctx context.Context, nodeName string, ctid int) (*pve.Task, error) {
	ct, err := c.Container(ctx, nodeName, ctid)
	if err != nil {
		return nil, err
	}
	task, err := ct.Stop(ctx)
	if err != nil {
		return nil, fmt.Errorf("stop container %d: %w", ctid, err)
	}
	return task, nil
}


func (c *Client) ShutdownContainer(ctx context.Context, nodeName string, ctid int) (*pve.Task, error) {
	ct, err := c.Container(ctx, nodeName, ctid)
	if err != nil {
		return nil, err
	}
	task, err := ct.Shutdown(ctx, true, 60)
	if err != nil {
		return nil, fmt.Errorf("shutdown container %d: %w", ctid, err)
	}
	return task, nil
}


func (c *Client) DeleteContainer(ctx context.Context, nodeName string, ctid int) (*pve.Task, error) {
	ct, err := c.Container(ctx, nodeName, ctid)
	if err != nil {
		return nil, err
	}
	task, err := ct.Delete(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete container %d: %w", ctid, err)
	}
	return task, nil
}




func (c *Client) NodeStorages(ctx context.Context, nodeName string) (pve.Storages, error) {
	node, err := c.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	storages, err := node.Storages(ctx)
	if err != nil {
		return nil, fmt.Errorf("storages on node %q: %w", nodeName, err)
	}
	return storages, nil
}


func (c *Client) NodeStorage(ctx context.Context, nodeName, storageName string) (*pve.Storage, error) {
	node, err := c.Node(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	storage, err := node.Storage(ctx, storageName)
	if err != nil {
		return nil, fmt.Errorf("storage %q on node %q: %w", storageName, nodeName, err)
	}
	return storage, nil
}






func (c *Client) NodeTasks(ctx context.Context, nodeName string) (pve.Tasks, error) {
	var tasks pve.Tasks
	if err := c.pve.Get(ctx, fmt.Sprintf("/nodes/%s/tasks", nodeName), &tasks); err != nil {
		return nil, fmt.Errorf("tasks on node %q: %w", nodeName, err)
	}
	return tasks, nil
}




func (c *Client) WaitForTask(ctx context.Context, task *pve.Task) error {
	if err := task.Wait(ctx, 1*time.Second, 300*time.Second); err != nil {
		return fmt.Errorf("waiting for task %s: %w", task.UPID, err)
	}
	return nil
}

