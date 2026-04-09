package tools

import (
"context"
"fmt"

"github.com/mark3labs/mcp-go/mcp"
"github.com/mark3labs/mcp-go/server"
"go.uber.org/zap"
)


func RegisterNodeTools(s *server.MCPServer, h *Handler) {
	
	s.AddTool(
mcp.NewTool("list_nodes",
mcp.WithDescription("Lists all Proxmox VE nodes in the cluster, including status, CPU, memory, and disk metrics."),
),
h.handleListNodes,
)

	
	s.AddTool(
mcp.NewTool("get_node_status",
mcp.WithDescription("Returns detailed runtime status for a single Proxmox VE node."),
mcp.WithString("node",
mcp.Required(),
				mcp.Description("Name of the Proxmox node (e.g. pve01)"),
			),
		),
		h.handleGetNodeStatus,
	)

	
	s.AddTool(
mcp.NewTool("get_node_storage",
mcp.WithDescription("Lists all storage pools available on a node with usage statistics."),
mcp.WithString("node",
mcp.Required(),
				mcp.Description("Name of the Proxmox node"),
			),
		),
		h.handleGetNodeStorage,
	)

	
	s.AddTool(
mcp.NewTool("get_node_tasks",
mcp.WithDescription("Returns the task history log for a specific node, including running tasks."),
mcp.WithString("node",
mcp.Required(),
				mcp.Description("Name of the Proxmox node"),
			),
		),
		h.handleGetNodeTasks,
	)
}

func (h *Handler) handleListNodes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.Logger.Debug("tool: list_nodes")
	nodes, err := h.Client.Nodes(ctxFrom(ctx))
	if err != nil {
		h.Logger.Error("list_nodes failed", zap.Error(err))
		return errResult(fmt.Errorf("list nodes: %w", err)), nil
	}
	return jsonResult(nodes)
}

func (h *Handler) handleGetNodeStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Debug("tool: get_node_status", zap.String("node", nodeName))

	node, err := h.Client.Node(ctxFrom(ctx), nodeName)
	if err != nil {
		h.Logger.Error("get_node_status failed", zap.String("node", nodeName), zap.Error(err))
		return errResult(fmt.Errorf("get node status for %q: %w", nodeName, err)), nil
	}
	return jsonResult(node)
}

func (h *Handler) handleGetNodeStorage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Debug("tool: get_node_storage", zap.String("node", nodeName))

	storages, err := h.Client.NodeStorages(ctxFrom(ctx), nodeName)
	if err != nil {
		h.Logger.Error("get_node_storage failed", zap.String("node", nodeName), zap.Error(err))
		return errResult(fmt.Errorf("get storage for node %q: %w", nodeName, err)), nil
	}
	return jsonResult(storages)
}

func (h *Handler) handleGetNodeTasks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Debug("tool: get_node_tasks", zap.String("node", nodeName))

	tasks, err := h.Client.NodeTasks(ctxFrom(ctx), nodeName)
	if err != nil {
		h.Logger.Error("get_node_tasks failed", zap.String("node", nodeName), zap.Error(err))
		return errResult(fmt.Errorf("get tasks for node %q: %w", nodeName, err)), nil
	}
	return jsonResult(tasks)
}

