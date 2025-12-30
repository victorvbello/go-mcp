package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"

	MCPServer "github.com/victorvbello/gomcp/mcp/server"
	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
	utils "github.com/victorvbello/gomcp/mcp/utils"
	utilsLogger "github.com/victorvbello/gomcp/mcp/utils/logger"
)

var dataKind = []string{"user", "payment", "products"}
var userList = map[int]map[string]interface{}{
	42: map[string]interface{}{
		"id":    42,
		"name":  "test user 42 name",
		"email": "test.user.42@email.com",
	},
	43: map[string]interface{}{
		"id":    43,
		"name":  "test user 43 name",
		"email": "test.user.43@email.com",
	},
	44: map[string]interface{}{
		"id":    44,
		"name":  "test user 44 name",
		"email": "test.user.44@email.com",
	},
}

func ExampleEverythingWithSTDIOServer() {
	logger := utilsLogger.NewLoggerService("everything-with-stdio-server")
	serverInfo := types.Implementation{}
	serverInfo.Name = "everything-with-stdio-server"
	serverInfo.Title = "Everything whit stdio server"
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

	_, err = mpcServer.RegisterPrompt(MCPServer.RegisterPromptOpts{
		Name:        "describe-json-and-transform-to-table",
		Title:       "Generate a description and a table using JSON data",
		Description: "Create a short description based on the JSON input, and generate a table using the provided data",
		Arguments: map[string]MCPServer.PromptArgsSchemaField{
			"kind": MCPServer.PromptArgsSchemaField{
				Description: "Data kind",
				Complete: func(values string, ctx types.CompleteParamsContext) []string {
					return dataKind
				},
				IsOptional: false,
			},
			"json": MCPServer.PromptArgsSchemaField{
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
							fmt.Sprintf("Please generare a short description using this json data: '%s', after that generate a table in markdown using the same json data. Remember, this data is kind %s",
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
		Uri:  "file://seven-golden-rules.txt",
		Meta: &MCPServer.ResourceMetadata{
			Resource: types.Resource{
				BaseMetadata: types.BaseMetadata{
					Name:  "seven-golden-rules.txt",
					Title: "Summary of the seven golden rules",
				},
				URI:         "file://seven-golden-rules.txt",
				Description: "The financial principles from The Richest Man in Babylon, by George Samuel Clason.",
				MIMEType:    "text/plain",
			},
		},
		Callback: func(uri string, extra *shared.RequestHandlerExtra) (*types.ReadResourceResult, error) {
			realUrl := "https://raw.githubusercontent.com/victorvbello/go-mcp/refs/heads/main/example/server/seven-golden-rules.txt"
			resultRequest, err := utils.HttpRequest(http.MethodGet, realUrl, nil, nil)
			if err != nil {
				return nil, fmt.Errorf("invalid request for seven-golden-rules %s", err)
			}
			result := &types.ReadResourceResult{
				Contents: []types.ResourceContents{
					&types.TextResourceContents{
						BaseResourceContents: types.BaseResourceContents{
							URI:      "file://seven-golden-rules.txt",
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
	_, err = mpcServer.RegisterResource(MCPServer.RegisterResourceOpts{
		Name: "terms-and-conditions",
		Uri:  "file://terms-and-conditions.txt",
		Meta: &MCPServer.ResourceMetadata{
			Resource: types.Resource{
				BaseMetadata: types.BaseMetadata{
					Name:  "terms-and-conditions.txt",
					Title: "Terms And Conditions",
				},
				URI:         "file://terms-and-conditions.txt",
				Description: "A short terms and conditions of this",
				MIMEType:    "text/plain",
			},
		},
		Callback: func(uri string, extra *shared.RequestHandlerExtra) (*types.ReadResourceResult, error) {
			realUrl := "https://raw.githubusercontent.com/victorvbello/go-mcp/refs/heads/main/example/server/terms-and-conditions.txt"
			resultRequest, err := utils.HttpRequest(http.MethodGet, realUrl, nil, nil)
			if err != nil {
				return nil, fmt.Errorf("invalid request for seven-golden-rules %s", err)
			}
			result := &types.ReadResourceResult{
				Contents: []types.ResourceContents{
					&types.TextResourceContents{
						BaseResourceContents: types.BaseResourceContents{
							URI:      "file://seven-golden-rules.txt",
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
	userProfileUrlTemplate, err := utils.NewUriTemplate("app://users/{user_id}")
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf(" utils.NewUriTemplate user-profile %v", err))
	}

	userProfileResourceTemplate := MCPServer.NewResourceTemplate(*userProfileUrlTemplate, MCPServer.ResourceTemplateCallbacks{
		Complete: map[string]MCPServer.CompleteResourceTemplateCallback{
			"user_id": func(value string, context types.CompleteParamsContext) ([]string, error) {
				return []string{value}, nil
			},
		},
	})
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
				Description: "Return all de user profile detail",
				MIMEType:    "application/json",
			},
		},
		Callback: func(uri string, variables utils.UriVariables, extra *shared.RequestHandlerExtra) (*types.ReadResourceResult, error) {
			userIDVars, okUserIDVars := variables["user_id"]
			if !okUserIDVars {
				return nil, fmt.Errorf("resource template user-profile: var user_id not found")
			}
			if len(userIDVars) == 0 {
				return nil, fmt.Errorf("resource template user-profile: user_id is required")
			}
			userID, err := strconv.Atoi(userIDVars[0])
			if err != nil {
				return nil, fmt.Errorf("resource template user-profile: strconv.Atoi user id%s", userIDVars[0])
			}
			userData, okUserData := userList[userID]
			if !okUserData {
				return nil, fmt.Errorf("resource template user-profile: user_id %d not found", userID)
			}
			userDataB, err := json.Marshal(userData)
			if err != nil {
				return nil, fmt.Errorf("resource template user-profile: json.Marshal user id %d data error %v", userID, err)
			}
			result := &types.ReadResourceResult{
				Contents: []types.ResourceContents{
					&types.TextResourceContents{
						BaseResourceContents: types.BaseResourceContents{
							URI:      uri,
							MIMEType: "application/json",
						},
						Text: string(userDataB),
					},
				},
			}
			return result, nil
		},
	})
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mpcServer.RegisterResourceTemplate user-profile %v", err))
	}

	_, err = mpcServer.RegisterTool(MCPServer.RegisterToolOpts{
		Name:        "generate_wealthy_plan",
		Description: "Generate a wealthy plan using the user salary",
		InputSchema: types.ToolInputSchema{
			Type: "object",
			Properties: map[string]types.ToolInputSchemaProperties{
				"salary": types.ToolInputSchemaProperties{
					Type:        "string",
					Description: "Salary",
				},
				"frequency": types.ToolInputSchemaProperties{
					Type:        "string",
					Description: "Salary frequency Weekly, Semi-monthly, Monthly",
				},
			},
			Required: []string{"salary", "frequency"},
		},
		Callback: func(args map[string]interface{}, extra *shared.RequestHandlerExtra) (*types.CallToolResult, error) {
			salary := args["salary"]
			frequency := args["frequency"]
			if salary == "" || frequency == "" {
				return nil, fmt.Errorf("args salary and frequency are required")
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
							fmt.Sprintf("Please generate a wealthy plan for a person that has a salary of %s, and his payment frequency are %s. For this you should use the resource %s", salary, frequency, "seven-golden-rules.txt"),
						),
					},
				},
				MaxTokens: 500,
			}, rExtra)
			if err != nil {
				return nil, fmt.Errorf("mpcServer.GetServer().CreateMessage %v", err)
			}

			resultText := "Unable to generate wealthy plan"
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
					types.NewResourceLink(types.Resource{
						BaseMetadata: types.BaseMetadata{
							Name:  "seven-golden-rules.txt",
							Title: "Summary of the seven golden rules",
						},
						URI:         "file://seven-golden-rules.txt",
						Description: "The financial principles from The Richest Man in Babylon, by George Samuel Clason.",
						MIMEType:    "text/plain",
					}),
				},
			}
			return result, nil
		},
	})
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mpcServer.RegisterTool summarize %v", err))
	}

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
					Description: "First number",
				},
				"b": types.ToolInputSchemaProperties{
					Type:        "number",
					Description: "Second number",
				},
			},
			Required: []string{"a", "b"},
		},
		OutputSchema: &types.ToolOutputSchema{
			Type: "object",
			Properties: map[string]types.ToolOutputSchemaProperties{
				"a": types.ToolOutputSchemaProperties{
					Type:        "number",
					Description: "First number",
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

			extra.SendNotification(extra.Context, types.NewLoggingMessageNotification(&types.LoggingMessageNotificationParams{
				Level: types.LOGGING_LEVEL_INFO,
				Data:  "All params are valid and now starting sum",
			}))

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

	_, err = mpcServer.RegisterTool(MCPServer.RegisterToolOpts{
		Name:        "check-new-resources",
		Description: "Update resource list",
		Callback: func(args map[string]interface{}, extra *shared.RequestHandlerExtra) (*types.CallToolResult, error) {
			_, err = mpcServer.RegisterResource(MCPServer.RegisterResourceOpts{
				Name: "another-text-resource",
				Uri:  "text://another-text-resource",
				Meta: &MCPServer.ResourceMetadata{
					Resource: types.Resource{
						BaseMetadata: types.BaseMetadata{
							Name:  "another-text-resource",
							Title: "Another text resource",
						},
						URI:         "text://another-text-resource",
						Description: "Another text resource to test",
						MIMEType:    "text/plain",
					},
				},
				Callback: func(uri string, extra *shared.RequestHandlerExtra) (*types.ReadResourceResult, error) {
					result := &types.ReadResourceResult{
						Contents: []types.ResourceContents{
							&types.TextResourceContents{
								BaseResourceContents: types.BaseResourceContents{
									URI:      "text://another-text-resource",
									MIMEType: "text/plain",
								},
								Text: "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book.",
							},
						},
					}
					return result, nil
				},
			})
			if err != nil {
				logger.Fatal(nil, fmt.Sprintf("mpcServer.RegisterResource another-text-resource %v", err))
			}
			extra.SendNotification(extra.Context, types.NewResourceListChangedNotification(&types.BaseNotificationParams{}))
			result := &types.CallToolResult{}
			return result, nil
		},
	})
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mpcServer.RegisterTool sum%v", err))
	}

	transport := MCPServer.NewStdioServerTransport(os.Stdin, os.Stdout, os.Stderr)
	logger.Info(nil, "MCP server is running...")
	ctx := context.Background()
	err = mpcServer.Connect(ctx, transport)
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mpcServer.Connect %v", err))
	}
}
