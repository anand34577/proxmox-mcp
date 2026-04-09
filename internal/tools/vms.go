package tools

import (
	"context"
	"fmt"

	pve "github.com/luthermonson/go-proxmox"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)


func RegisterVMTools(s *server.MCPServer, h *Handler) {
	
	s.AddTool(
		mcp.NewTool("list_vms",
			mcp.WithDescription("Lists all QEMU virtual machines on the specified node with status, CPU, and memory usage."),
			mcp.WithString("node",
				mcp.Required(),
				mcp.Description("Proxmox node name (e.g. pve01)"),
			),
		),
		h.handleListVMs,
	)

	
	s.AddTool(
		mcp.NewTool("get_vm_status",
			mcp.WithDescription("Returns the current runtime status of a specific VM (running, stopped, paused, etc.)."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID (integer)")),
		),
		h.handleGetVMStatus,
	)

	
	s.AddTool(
		mcp.NewTool("start_vm",
			mcp.WithDescription("Powers on a stopped or paused VM. Returns a task UPID to track progress."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID")),
		),
		h.handleStartVM,
	)

	
	s.AddTool(
		mcp.NewTool("stop_vm",
			mcp.WithDescription("Force-stops a VM immediately (equivalent to pulling the power cord). Destructive — data loss possible."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID")),
		),
		h.handleStopVM,
	)

	
	s.AddTool(
		mcp.NewTool("shutdown_vm",
			mcp.WithDescription("Sends an ACPI shutdown signal to gracefully stop the guest OS."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID")),
		),
		h.handleShutdownVM,
	)

	
	s.AddTool(
		mcp.NewTool("reboot_vm",
			mcp.WithDescription("Sends an ACPI reboot signal to the VM guest OS."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID")),
		),
		h.handleRebootVM,
	)

	
	s.AddTool(
		mcp.NewTool("suspend_vm",
			mcp.WithDescription("Suspends a running VM (saves RAM to disk, QEMU freeze)."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID")),
		),
		h.handleSuspendVM,
	)

	
	s.AddTool(
		mcp.NewTool("resume_vm",
			mcp.WithDescription("Resumes a previously suspended VM."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID")),
		),
		h.handleResumeVM,
	)

	
	s.AddTool(
		mcp.NewTool("delete_vm",
			mcp.WithDescription("Permanently deletes a VM and its associated disk images. VM must be stopped. DESTRUCTIVE."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID")),
		),
		h.handleDeleteVM,
	)

	
	s.AddTool(
		mcp.NewTool("clone_vm",
			mcp.WithDescription("Clones an existing VM (full or linked clone). Returns new VMID and task UPID."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Source node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("Source VM ID")),
			mcp.WithNumber("newid", mcp.Required(), mcp.Description("New VM ID for the clone")),
			mcp.WithString("name", mcp.Description("Hostname for the clone")),
			mcp.WithString("description", mcp.Description("Description for the clone")),
			mcp.WithBoolean("full", mcp.Description("Full clone (true) or linked clone (false, default)")),
		),
		h.handleCloneVM,
	)

	
	s.AddTool(
		mcp.NewTool("list_vm_snapshots",
			mcp.WithDescription("Lists all snapshots for a VM."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID")),
		),
		h.handleListVMSnapshots,
	)

	
	s.AddTool(
		mcp.NewTool("create_vm_snapshot",
			mcp.WithDescription("Creates a new snapshot for a VM."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Snapshot name (no spaces)")),
			mcp.WithString("description", mcp.Description("Optional snapshot description")),
		),
		h.handleCreateVMSnapshot,
	)

	
	s.AddTool(
		mcp.NewTool("rollback_vm_snapshot",
			mcp.WithDescription("Rolls back a VM to a previously created snapshot. DESTRUCTIVE — current VM state is lost."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID")),
			mcp.WithString("snapshot", mcp.Required(), mcp.Description("Snapshot name to roll back to")),
		),
		h.handleRollbackVMSnapshot,
	)

	
	s.AddTool(
		mcp.NewTool("delete_vm_snapshot",
			mcp.WithDescription("Deletes a snapshot from a VM. DESTRUCTIVE."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("vmid", mcp.Required(), mcp.Description("VM ID")),
			mcp.WithString("snapshot", mcp.Required(), mcp.Description("Snapshot name to delete")),
		),
		h.handleDeleteVMSnapshot,
	)
}



func (h *Handler) handleListVMs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Debug("tool: list_vms", zap.String("node", nodeName))
	vms, err := h.Client.VirtualMachines(ctxFrom(ctx), nodeName)
	if err != nil {
		h.Logger.Error("list_vms failed", zap.String("node", nodeName), zap.Error(err))
		return errResult(fmt.Errorf("list VMs on %q: %w", nodeName, err)), nil
	}
	return jsonResult(vms)
}

func (h *Handler) handleGetVMStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Debug("tool: get_vm_status", zap.String("node", nodeName), zap.Int("vmid", vmid))
	vm, err := h.Client.VirtualMachine(ctxFrom(ctx), nodeName, vmid)
	if err != nil {
		return errResult(fmt.Errorf("get VM %d on %q: %w", vmid, nodeName, err)), nil
	}
	return jsonResult(vm)
}

func (h *Handler) handleStartVM(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Info("tool: start_vm", zap.String("node", nodeName), zap.Int("vmid", vmid))
	task, err := h.Client.StartVM(ctxFrom(ctx), nodeName, vmid)
	if err != nil {
		return errResult(fmt.Errorf("start VM %d on %q: %w", vmid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{"upid": task.UPID, "vmid": vmid, "node": nodeName, "status": "started"})
}

func (h *Handler) handleStopVM(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if guard := h.guardDestructive("stop_vm"); guard != nil {
		return guard, nil
	}
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Warn("tool: stop_vm (force)", zap.String("node", nodeName), zap.Int("vmid", vmid))
	task, err := h.Client.StopVM(ctxFrom(ctx), nodeName, vmid)
	if err != nil {
		return errResult(fmt.Errorf("stop VM %d on %q: %w", vmid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{"upid": task.UPID, "vmid": vmid, "node": nodeName, "status": "stopped"})
}

func (h *Handler) handleShutdownVM(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Info("tool: shutdown_vm", zap.String("node", nodeName), zap.Int("vmid", vmid))
	task, err := h.Client.ShutdownVM(ctxFrom(ctx), nodeName, vmid)
	if err != nil {
		return errResult(fmt.Errorf("shutdown VM %d on %q: %w", vmid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{"upid": task.UPID, "vmid": vmid, "node": nodeName, "status": "shutting_down"})
}

func (h *Handler) handleRebootVM(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Info("tool: reboot_vm", zap.String("node", nodeName), zap.Int("vmid", vmid))
	task, err := h.Client.RebootVM(ctxFrom(ctx), nodeName, vmid)
	if err != nil {
		return errResult(fmt.Errorf("reboot VM %d on %q: %w", vmid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{"upid": task.UPID, "vmid": vmid, "node": nodeName, "status": "rebooting"})
}

func (h *Handler) handleSuspendVM(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Info("tool: suspend_vm", zap.String("node", nodeName), zap.Int("vmid", vmid))
	task, err := h.Client.SuspendVM(ctxFrom(ctx), nodeName, vmid)
	if err != nil {
		return errResult(fmt.Errorf("suspend VM %d on %q: %w", vmid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{"upid": task.UPID, "vmid": vmid, "node": nodeName, "status": "suspending"})
}

func (h *Handler) handleResumeVM(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Info("tool: resume_vm", zap.String("node", nodeName), zap.Int("vmid", vmid))
	task, err := h.Client.ResumeVM(ctxFrom(ctx), nodeName, vmid)
	if err != nil {
		return errResult(fmt.Errorf("resume VM %d on %q: %w", vmid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{"upid": task.UPID, "vmid": vmid, "node": nodeName, "status": "resuming"})
}

func (h *Handler) handleDeleteVM(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if guard := h.guardDestructive("delete_vm"); guard != nil {
		return guard, nil
	}
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Warn("tool: delete_vm", zap.String("node", nodeName), zap.Int("vmid", vmid))
	task, err := h.Client.DeleteVM(ctxFrom(ctx), nodeName, vmid)
	if err != nil {
		return errResult(fmt.Errorf("delete VM %d on %q: %w", vmid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{"upid": task.UPID, "vmid": vmid, "node": nodeName, "status": "deleting"})
}

func (h *Handler) handleCloneVM(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	newid, err := intArg(req, "newid")
	if err != nil {
		return errResult(err), nil
	}
	name := optionalStringArg(req, "name", "")
	description := optionalStringArg(req, "description", "")
	full := boolArg(req, "full", false)

	h.Logger.Info("tool: clone_vm",
		zap.String("node", nodeName), zap.Int("vmid", vmid), zap.Int("newid", newid))

	opts := pve.VirtualMachineCloneOptions{
		NewID:       newid,
		Name:        name,
		Description: description,
		Full:        uint8(boolToInt(full)), 
	}

	task, clonedID, err := h.Client.CloneVM(ctxFrom(ctx), nodeName, vmid, opts)
	if err != nil {
		return errResult(fmt.Errorf("clone VM %d → %d on %q: %w", vmid, newid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{
		"upid":        task.UPID,
		"source_vmid": vmid,
		"new_vmid":    clonedID,
		"node":        nodeName,
		"status":      "cloning",
	})
}

func (h *Handler) handleListVMSnapshots(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Debug("tool: list_vm_snapshots", zap.String("node", nodeName), zap.Int("vmid", vmid))
	snaps, err := h.Client.VMSnapshots(ctxFrom(ctx), nodeName, vmid)
	if err != nil {
		return errResult(fmt.Errorf("list snapshots for VM %d on %q: %w", vmid, nodeName, err)), nil
	}
	return jsonResult(snaps)
}

func (h *Handler) handleCreateVMSnapshot(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	snapName, err := stringArg(req, "name")
	if err != nil {
		return errResult(err), nil
	}
	description := optionalStringArg(req, "description", "")

	h.Logger.Info("tool: create_vm_snapshot",
		zap.String("node", nodeName), zap.Int("vmid", vmid), zap.String("snapshot", snapName))

	task, err := h.Client.CreateVMSnapshot(ctxFrom(ctx), nodeName, vmid, snapName, description)
	if err != nil {
		return errResult(fmt.Errorf("create snapshot %q for VM %d: %w", snapName, vmid, err)), nil
	}
	return jsonResult(map[string]any{
		"upid":     task.UPID,
		"vmid":     vmid,
		"node":     nodeName,
		"snapshot": snapName,
		"status":   "creating",
	})
}

func (h *Handler) handleRollbackVMSnapshot(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if guard := h.guardDestructive("rollback_vm_snapshot"); guard != nil {
		return guard, nil
	}
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	snapName, err := stringArg(req, "snapshot")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Warn("tool: rollback_vm_snapshot",
		zap.String("node", nodeName), zap.Int("vmid", vmid), zap.String("snapshot", snapName))

	task, err := h.Client.RollbackVMSnapshot(ctxFrom(ctx), nodeName, vmid, snapName)
	if err != nil {
		return errResult(fmt.Errorf("rollback VM %d to snapshot %q: %w", vmid, snapName, err)), nil
	}
	return jsonResult(map[string]any{
		"upid":     task.UPID,
		"vmid":     vmid,
		"node":     nodeName,
		"snapshot": snapName,
		"status":   "rolling_back",
	})
}

func (h *Handler) handleDeleteVMSnapshot(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if guard := h.guardDestructive("delete_vm_snapshot"); guard != nil {
		return guard, nil
	}
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	vmid, err := intArg(req, "vmid")
	if err != nil {
		return errResult(err), nil
	}
	snapName, err := stringArg(req, "snapshot")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Warn("tool: delete_vm_snapshot",
		zap.String("node", nodeName), zap.Int("vmid", vmid), zap.String("snapshot", snapName))

	task, err := h.Client.DeleteVMSnapshot(ctxFrom(ctx), nodeName, vmid, snapName)
	if err != nil {
		return errResult(fmt.Errorf("delete snapshot %q from VM %d: %w", snapName, vmid, err)), nil
	}
	return jsonResult(map[string]any{
		"upid":     task.UPID,
		"vmid":     vmid,
		"node":     nodeName,
		"snapshot": snapName,
		"status":   "deleting",
	})
}


func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

