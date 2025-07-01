package server

import (
	"context"

	"github.com/victorvbello/gomcp/mcp/types"
)

type ToolHandlerFunc func(ctx context.Context, request types.CallToolRequest) (*types.CallToolResult, error)

type MCPServer struct {
}
