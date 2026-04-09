

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"

	"github.com/anand34577/proxmox-mcp/internal/config"
	pxclient "github.com/anand34577/proxmox-mcp/internal/proxmox"
)



type Handler struct {
	Client *pxclient.Client
	Config *config.Config
	Logger *zap.Logger
}


func NewHandler(client *pxclient.Client, cfg *config.Config, log *zap.Logger) *Handler {
	return &Handler{Client: client, Config: cfg, Logger: log}
}








func getArgs(req mcp.CallToolRequest) map[string]any {
	if m, ok := req.Params.Arguments.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}


func stringArg(req mcp.CallToolRequest, key string) (string, error) {
	v, ok := getArgs(req)[key]
	if !ok {
		return "", fmt.Errorf("missing required argument %q", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("argument %q must be a string, got %T", key, v)
	}
	return s, nil
}


func optionalStringArg(req mcp.CallToolRequest, key, fallback string) string {
	if v, ok := getArgs(req)[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return fallback
}


func intArg(req mcp.CallToolRequest, key string) (int, error) {
	v, ok := getArgs(req)[key]
	if !ok {
		return 0, fmt.Errorf("missing required argument %q", key)
	}
	switch t := v.(type) {
	case float64:
		return int(t), nil
	case string:
		n, err := strconv.Atoi(t)
		if err != nil {
			return 0, fmt.Errorf("argument %q: %w", key, err)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("argument %q must be a number, got %T", key, v)
	}
}


func optionalIntArg(req mcp.CallToolRequest, key string, fallback int) int {
	v, err := intArg(req, key)
	if err != nil {
		return fallback
	}
	return v
}


func boolArg(req mcp.CallToolRequest, key string, fallback bool) bool {
	v, ok := getArgs(req)[key]
	if !ok {
		return fallback
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return fallback
}





func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("serialise result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}


func textResult(msg string) *mcp.CallToolResult {
	return mcp.NewToolResultText(msg)
}


func errResult(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}





func (h *Handler) guardDestructive(op string) *mcp.CallToolResult {
	if !h.Config.Security.AllowDestructive {
		return mcp.NewToolResultError(
			fmt.Sprintf("operation %q is disabled; set PROXMOX_MCP_SECURITY_ALLOW_DESTRUCTIVE=true to enable", op),
		)
	}
	return nil
}




func ctxFrom(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

