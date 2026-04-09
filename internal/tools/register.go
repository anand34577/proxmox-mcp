package tools

import (
	"github.com/mark3labs/mcp-go/server"
)



func RegisterAll(s *server.MCPServer, h *Handler) {
	RegisterClusterTools(s, h)
	RegisterNodeTools(s, h)
	RegisterVMTools(s, h)
	RegisterContainerTools(s, h)
	RegisterStorageTools(s, h)
}

