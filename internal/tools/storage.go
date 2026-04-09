package tools

import (
"context"
"fmt"

"github.com/mark3labs/mcp-go/mcp"
"github.com/mark3labs/mcp-go/server"
"go.uber.org/zap"
)


func RegisterStorageTools(s *server.MCPServer, h *Handler) {
	
	s.AddTool(
mcp.NewTool("list_node_storage",
mcp.WithDescription("Lists all storage pools attached to a node with total/used/available space."),
mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
		),
		h.handleListNodeStorage,
	)

	
	s.AddTool(
mcp.NewTool("get_storage_detail",
mcp.WithDescription("Returns detailed information about a specific storage pool on a node."),
mcp.WithString("node", mcp.Required(), mcp.Description("Proxmox node name")),
			mcp.WithString("storage", mcp.Required(), mcp.Description("Storage pool name (e.g. local, local-lvm)")),
		),
		h.handleGetStorageDetail,
	)
}

func (h *Handler) handleListNodeStorage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Debug("tool: list_node_storage", zap.String("node", nodeName))
	storages, err := h.Client.NodeStorages(ctxFrom(ctx), nodeName)
	if err != nil {
		h.Logger.Error("list_node_storage failed", zap.String("node", nodeName), zap.Error(err))
		return errResult(fmt.Errorf("list storage on %q: %w", nodeName, err)), nil
	}
	return jsonResult(storages)
}

func (h *Handler) handleGetStorageDetail(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName, err := stringArg(req, "node")
	if err != nil {
		return errResult(err), nil
	}
	storageName, err := stringArg(req, "storage")
	if err != nil {
		return errResult(err), nil
	}
	h.Logger.Debug("tool: get_storage_detail",
zap.String("node", nodeName), zap.String("storage", storageName))

	storage, err := h.Client.NodeStorage(ctxFrom(ctx), nodeName, storageName)
	if err != nil {
		h.Logger.Error("get_storage_detail failed",
zap.String("node", nodeName), zap.String("storage", storageName), zap.Error(err))
		return errResult(fmt.Errorf("get storage %q on node %q: %w", storageName, nodeName, err)), nil
	}
	return jsonResult(storage)
}

