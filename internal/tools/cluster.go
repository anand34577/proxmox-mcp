package tools

import (
"context"
"fmt"

"github.com/mark3labs/mcp-go/mcp"
"github.com/mark3labs/mcp-go/server"
"go.uber.org/zap"
)


func RegisterClusterTools(s *server.MCPServer, h *Handler) {
	
	s.AddTool(
mcp.NewTool("get_cluster_status",
mcp.WithDescription("Returns the overall Proxmox cluster health status including quorum, node count, and resource usage."),
),
h.handleGetClusterStatus,
)

	
	s.AddTool(
mcp.NewTool("get_cluster_resources",
mcp.WithDescription("Lists all cluster resources: VMs, LXC containers, storage, and nodes. Optionally filter by resource type."),
mcp.WithString("type",
mcp.Description("Resource type filter: vm | lxc | storage | node | sdn (leave empty for all)"),
),
),
h.handleGetClusterResources,
)

	
	s.AddTool(
mcp.NewTool("get_proxmox_version",
mcp.WithDescription("Returns the Proxmox VE server version."),
),
h.handleGetVersion,
)
}

func (h *Handler) handleGetClusterStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.Logger.Debug("tool: get_cluster_status")
	statuses, err := h.Client.ClusterStatus(ctxFrom(ctx))
	if err != nil {
		h.Logger.Error("get_cluster_status failed", zap.Error(err))
		return errResult(fmt.Errorf("get cluster status: %w", err)), nil
	}
	return jsonResult(statuses)
}

func (h *Handler) handleGetClusterResources(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType := optionalStringArg(req, "type", "")
	h.Logger.Debug("tool: get_cluster_resources", zap.String("type", resourceType))
	resources, err := h.Client.ClusterResources(ctxFrom(ctx), resourceType)
	if err != nil {
		h.Logger.Error("get_cluster_resources failed", zap.Error(err))
		return errResult(fmt.Errorf("get cluster resources: %w", err)), nil
	}
	return jsonResult(resources)
}

func (h *Handler) handleGetVersion(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.Logger.Debug("tool: get_proxmox_version")
	v, err := h.Client.Version(ctxFrom(ctx))
	if err != nil {
		return errResult(fmt.Errorf("get version: %w", err)), nil
	}
	return jsonResult(v)
}

