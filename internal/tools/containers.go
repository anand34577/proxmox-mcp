package tools

import (
"context"
"fmt"

"github.com/mark3labs/mcp-go/mcp"
"github.com/mark3labs/mcp-go/server"
"go.uber.org/zap"
)


func RegisterContainerTools(s *server.MCPServer, h *Handler) {
	
	s.AddTool(
mcp.NewTool("list_containers",
mcp.WithDescription("Lists all LXC containers on the specified node with status and resource usage."),
mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
		),
		h.handleListContainers,
	)

	
	s.AddTool(
mcp.NewTool("get_container_status",
mcp.WithDescription("Returns the current runtime status of a specific LXC container."),
mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("ctid", mcp.Required(), mcp.Description("Container ID (integer)")),
		),
		h.handleGetContainerStatus,
	)

	
	s.AddTool(
mcp.NewTool("start_container",
mcp.WithDescription("Starts a stopped LXC container. Returns a task UPID."),
mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("ctid", mcp.Required(), mcp.Description("Container ID")),
		),
		h.handleStartContainer,
	)

	
	s.AddTool(
mcp.NewTool("stop_container",
mcp.WithDescription("Force-stops a running LXC container. DESTRUCTIVE."),
mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("ctid", mcp.Required(), mcp.Description("Container ID")),
		),
		h.handleStopContainer,
	)

	
	s.AddTool(
mcp.NewTool("shutdown_container",
mcp.WithDescription("Gracefully shuts down an LXC container via init system."),
mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("ctid", mcp.Required(), mcp.Description("Container ID")),
		),
		h.handleShutdownContainer,
	)

	
	s.AddTool(
mcp.NewTool("delete_container",
mcp.WithDescription("Permanently deletes an LXC container and its storage. Container must be stopped. DESTRUCTIVE."),
mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithNumber("ctid", mcp.Required(), mcp.Description("Container ID")),
		),
		h.handleDeleteContainer,
	)
}

func (h *Handler) handleListContainers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Debug("tool: list_containers", zap.String("node", nodeName))
	cts, err := h.Client.Containers(ctxFrom(ctx), nodeName)
	if err != nil {
		h.Logger.Error("list_containers failed", zap.String("node", nodeName), zap.Error(err))
		return errResult(fmt.Errorf("list containers on %q: %w", nodeName, err)), nil
	}
	return jsonResult(cts)
}

func (h *Handler) handleGetContainerStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	ctid, err := intArg(req, "ctid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Debug("tool: get_container_status", zap.String("node", nodeName), zap.Int("ctid", ctid))
	ct, err := h.Client.Container(ctxFrom(ctx), nodeName, ctid)
	if err != nil {
		return errResult(fmt.Errorf("get container %d on %q: %w", ctid, nodeName, err)), nil
	}
	return jsonResult(ct)
}

func (h *Handler) handleStartContainer(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	ctid, err := intArg(req, "ctid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Info("tool: start_container", zap.String("node", nodeName), zap.Int("ctid", ctid))
	task, err := h.Client.StartContainer(ctxFrom(ctx), nodeName, ctid)
	if err != nil {
		return errResult(fmt.Errorf("start container %d on %q: %w", ctid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{"upid": task.UPID, "ctid": ctid, "node": nodeName, "status": "started"})
}

func (h *Handler) handleStopContainer(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if guard := h.guardDestructive("stop_container"); guard != nil {
		return guard, nil
	}
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	ctid, err := intArg(req, "ctid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Warn("tool: stop_container (force)", zap.String("node", nodeName), zap.Int("ctid", ctid))
	task, err := h.Client.StopContainer(ctxFrom(ctx), nodeName, ctid)
	if err != nil {
		return errResult(fmt.Errorf("stop container %d on %q: %w", ctid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{"upid": task.UPID, "ctid": ctid, "node": nodeName, "status": "stopped"})
}

func (h *Handler) handleShutdownContainer(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	ctid, err := intArg(req, "ctid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Info("tool: shutdown_container", zap.String("node", nodeName), zap.Int("ctid", ctid))
	task, err := h.Client.ShutdownContainer(ctxFrom(ctx), nodeName, ctid)
	if err != nil {
		return errResult(fmt.Errorf("shutdown container %d on %q: %w", ctid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{"upid": task.UPID, "ctid": ctid, "node": nodeName, "status": "shutting_down"})
}

func (h *Handler) handleDeleteContainer(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if guard := h.guardDestructive("delete_container"); guard != nil {
		return guard, nil
	}
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	ctid, err := intArg(req, "ctid")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Warn("tool: delete_container", zap.String("node", nodeName), zap.Int("ctid", ctid))
	task, err := h.Client.DeleteContainer(ctxFrom(ctx), nodeName, ctid)
	if err != nil {
		return errResult(fmt.Errorf("delete container %d on %q: %w", ctid, nodeName, err)), nil
	}
	return jsonResult(map[string]any{"upid": task.UPID, "ctid": ctid, "node": nodeName, "status": "deleting"})
}

