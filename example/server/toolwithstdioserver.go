package server

import (
	"fmt"
	"os"

	MCPServer "github.com/victorvbello/gomcp/mcp/server"
	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
	utils "github.com/victorvbello/gomcp/mcp/utils/logger"
)

func ExampleToolWithSTDIOServer() {
	logger := utils.NewLoggerService()
	serverInfo := types.Implementation{}
	serverInfo.Name = "tools-with-stdio-server"
	serverInfo.Title = "Tools whit stdio"
	serverInfo.Version = "1.0.0"
	mpcServer, err := MCPServer.NewMcpServer(serverInfo, MCPServer.ServerOptions{})
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mcpServer.NewMcpServer %v", err))
	}
	mpcServer.SetOnInitialized(func() error {
		logger.Info(nil, "notification initialized")
		return nil
	})
	_, err = mpcServer.RegisterTool(MCPServer.RegisterToolOpts{
		Name:        "summarize",
		Description: "Summarize any text using an LLM",
		InputSchema: types.ToolInputSchema{
			Type: "object",
			Properties: map[string]types.ToolInputSchemaProperties{
				"text": types.ToolInputSchemaProperties{
					Type:        "string",
					Description: "Some text input",
				},
			},
			Required: []string{"text"},
		},
		Callback: func(args map[string]interface{}, extra *shared.RequestHandlerExtra) (*types.CallToolResult, error) {
			text, ok := args["text"]
			if !ok {
				return nil, fmt.Errorf("arg text not found")
			}
			// Call the LLM through MCP sampling
			response, err := mpcServer.GetServer().CreateMessage(types.CreateMessageParams{
				Messages: []types.SamplingMessage{
					{
						Role: "user",
						Content: types.NewTextContent(
							fmt.Sprintf("Please summarize the following text concisely:\n\n%s", text),
						),
					},
				},
				MaxTokens: 500,
			}, nil)
			if err != nil {
				return nil, fmt.Errorf("mpcServer.GetServer().CreateMessage %v", err)
			}

			resultText := "Unable to generate summary"
			safeTypeResponse, okType := response.(*types.CallToolResult)
			if okType && safeTypeResponse != nil && len(safeTypeResponse.Content) > 0 {
				for _, c := range safeTypeResponse.Content {
					safeTxtContent, okContentType := c.(*types.TextContent)
					if !okContentType {
						continue
					}
					resultText = safeTxtContent.Text
					break
				}
			}

			result := &types.CallToolResult{
				Content: []types.Content{
					types.NewTextContent(resultText),
				},
			}
			return result, nil
		},
	})
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mpcServer.RegisterTool %v", err))
	}

	transport := MCPServer.NewStdioServerTransport(os.Stdin, os.Stdout)
	logger.Info(nil, "MCP server is running...")
	err = mpcServer.Connect(transport)
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mpcServer.Connect %v", err))
	}
}
