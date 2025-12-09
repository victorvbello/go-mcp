package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/victorvbello/gomcp/mcp/server"
	MCPServer "github.com/victorvbello/gomcp/mcp/server"
	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
	utils "github.com/victorvbello/gomcp/mcp/utils"
	utilsLogger "github.com/victorvbello/gomcp/mcp/utils/logger"
)

var dataKind = []string{"user", "payment", "products"}

func ExampleToolWithSTDIOServer() {
	logger := utilsLogger.NewLoggerService()
	serverInfo := types.Implementation{}
	serverInfo.Name = "everything-with-stdio-server"
	serverInfo.Title = "Everything whit stdio"
	serverInfo.Version = "1.0.0"
	mpcServer, err := MCPServer.NewMcpServer(serverInfo,
		MCPServer.ServerOptions{
			Capabilities: types.ServerCapabilities{
				Logging:     struct{}{},
				Sampling:    struct{}{},
				Completions: struct{}{},
				Tools: &types.ServerCapabilitiesListChanged{
					ListChanged: true,
				},
				Prompts: &types.ServerCapabilitiesListChanged{
					ListChanged: true,
				},
				Resources: &types.ServerCapabilitiesResources{
					Subscribe: true,
					ServerCapabilitiesListChanged: types.ServerCapabilitiesListChanged{
						ListChanged: true,
					},
				},
			},
		})
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mcpServer.NewMcpServer %v", err))
	}
	mpcServer.SetOnInitialized(func() error {
		logger.Info(nil, "notification initialized")
		return nil
	})
	mpcServer.GetServer().SetOnErrorCallBack(func(err error) {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
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
			rExtra := &shared.RequestOptions{
				TransportSendOptions: shared.TransportSendOptions{
					RelatedRequestID: extra.RequestID,
				},
			}

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
			}, rExtra)
			if err != nil {
				return nil, fmt.Errorf("mpcServer.GetServer().CreateMessage %v", err)
			}

			resultText := "Unable to generate summary"
			safeTypeResponse, okType := response.(*types.CreateMessageResult)
			logger.Info(utilsLogger.LogFields{
				"is_create_message_result": okType,
			}, "Callback response")
			if okType && safeTypeResponse != nil && safeTypeResponse.Content != nil {
				if safeContentText, okType := safeTypeResponse.Content.(*types.TextContent); okType && safeContentText != nil {
					resultText = safeContentText.Text
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
		logger.Fatal(nil, fmt.Sprintf("mpcServer.RegisterTool summarize %v", err))
	}

	_, err = mpcServer.RegisterTool(MCPServer.RegisterToolOpts{
		Name:        "sum",
		Description: "Sum two numbers",
		InputSchema: types.ToolInputSchema{
			Type: "object",
			Properties: map[string]types.ToolInputSchemaProperties{
				"a": types.ToolInputSchemaProperties{
					Type:        "number",
					Description: "Frist number",
				},
				"b": types.ToolInputSchemaProperties{
					Type:        "number",
					Description: "Second number",
				},
			},
			Required: []string{"a", "b"},
		},
		OutputSchema: types.ToolOutputSchema{
			Type: "object",
			Properties: map[string]types.ToolOutputSchemaProperties{
				"a": types.ToolOutputSchemaProperties{
					Type:        "number",
					Description: "Frist number",
				},
				"b": types.ToolOutputSchemaProperties{
					Type:        "number",
					Description: "Second number",
				},
				"result": types.ToolOutputSchemaProperties{
					Type:        "number",
					Description: "Sum result a + b",
				},
			},
		},
		Callback: func(args map[string]interface{}, extra *shared.RequestHandlerExtra) (*types.CallToolResult, error) {
			a, ok := args["a"]
			if !ok {
				return nil, fmt.Errorf("a value not found")
			}
			b, ok := args["b"]
			if !ok {
				return nil, fmt.Errorf("b value not found")
			}

			aInt, ok := a.(float64)
			if !ok {
				return nil, fmt.Errorf("a not number")
			}
			bInt, ok := b.(float64)
			if !ok {
				return nil, fmt.Errorf("b not number")
			}

			content := map[string]interface{}{
				"a":      aInt,
				"b":      bInt,
				"result": aInt + bInt,
			}

			contentB, err := json.Marshal(content)
			if err != nil {
				return nil, fmt.Errorf("json.Marshal content %v", err)
			}

			result := &types.CallToolResult{
				Content: []types.Content{
					types.NewTextContent(string(contentB)),
				},
				StructuredContent: content,
			}
			return result, nil
		},
	})
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mpcServer.RegisterTool sum%v", err))
	}

	_, err = mpcServer.RegisterPrompt(MCPServer.RegisterPromptOpts{
		Name:        "describe-json-and-transform-to-table",
		Title:       "Generate a description and a table using JSON data",
		Description: "Create a short description based on the JSON input, and generate a table using the provided data",
		Arguments: map[string]server.PromptArgsSchemaField{
			"kind": server.PromptArgsSchemaField{
				Description: "Data kind",
				Complete: func(values string, ctx types.CompleteParamsContext) []string {
					return dataKind
				},
				IsOptional: false,
			},
			"json": server.PromptArgsSchemaField{
				Description: "JSON data",
				IsOptional:  false,
			},
		},
		Callback: func(args map[string]string, extra *shared.RequestHandlerExtra) (*types.GetPromptResult, error) {
			if args["json"] == "" || args["kind"] == "" {
				return nil, fmt.Errorf("json and kind are requerid")
			}
			if !json.Valid([]byte(args["json"])) {
				return nil, fmt.Errorf("json is not a valid json data")
			}
			jsonDataB, err := json.Marshal(args["json"])
			if err != nil {
				return nil, fmt.Errorf("json data json.Marshal err: %v", err)
			}
			result := types.GetPromptResult{
				Messages: []types.PromptMessage{
					types.PromptMessage{
						Role: "user",
						Content: types.NewTextContent(
							fmt.Sprintf("Please generare a short descripion using this json data: '%s', after that generate a table in markdown using the same json data. Remenber, this data is kind %s",
								string(jsonDataB), args["kind"])),
					}},
			}
			return &result, nil
		},
	})
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mpcServer.RegisterPrompt describe-json-and-transform-to-table %v", err))
	}

	_, err = mpcServer.RegisterResource(MCPServer.RegisterResourceOpts{
		Name: "seven-golden-rules",
		Uri:  "file:/seven-golden-rules.txt",
		Meta: &MCPServer.ResourceMetadata{
			Resource: types.Resource{
				BaseMetadata: types.BaseMetadata{
					Name:  "Seven golden rules",
					Title: "Sumary of the seven golden rules",
				},
				URI:         "file:/seven-golden-rules.txt",
				Description: "The financial principles from The Richest Man in Babylon, by George Samuel Clason.",
				MIMEType:    "text/plain",
			},
		},
		Callback: func(uri string, extra *shared.RequestHandlerExtra) (*types.ReadResourceResult, error) {
			realUrl := "https://github.com/victorvbello/go-mcp/blob/main/example/server/seven_golden_rules.txt"
			resultRequest, err := utils.HttpRequest(http.MethodGet, realUrl, nil, nil)
			if err != nil {
				return nil, fmt.Errorf("invalid request for seven-golden-rules %s", err)
			}
			result := &types.ReadResourceResult{
				Contents: []types.ResourceContents{
					types.TextResourceContents{
						BaseResourceContents: types.BaseResourceContents{
							URI:      "file:/seven-golden-rules.txt",
							MIMEType: "text/plain",
						},
						Text: resultRequest.RawBody,
					},
				},
			}
			return result, nil
		},
	})
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mpcServer.RegisterResource seven-golden-rules %v", err))
	}
	userProfileUrlTemplate, err := utils.NewUriTemplate("https://app/users/{user_id}")
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf(" utils.NewUriTemplate user-profile %v", err))
	}
	userProfileResourceTemplate := server.NewResourceTemplate(*userProfileUrlTemplate, MCPServer.ResourceTemplateCallbacks{})
	_, err = mpcServer.RegisterResourceTemplate(MCPServer.RegisterResourceTemplateOpts{
		Name:     "user-profile",
		Title:    "Return the user profile using his ID",
		Template: *userProfileResourceTemplate,
		Meta: &MCPServer.ResourceMetadata{
			Resource: types.Resource{
				BaseMetadata: types.BaseMetadata{
					Name:  "User profile",
					Title: "Return the user profile using his ID",
				},
				Description: "Return all de user prfile detail",
				MIMEType:    "application/json",
			},
		},
	})
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mpcServer.RegisterResourceTemplate user-profile %v", err))
	}

	transport := MCPServer.NewStdioServerTransport(os.Stdin, os.Stdout)
	logger.Info(nil, "MCP server is running...")
	ctx := context.Background()
	err = mpcServer.Connect(ctx, transport)
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mpcServer.Connect %v", err))
	}
}
