package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	MCPClient "github.com/victorvbello/gomcp/mcp/client"
	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
	utilsLogger "github.com/victorvbello/gomcp/mcp/utils/logger"
	"google.golang.org/genai"
)

func ExampleWithSTDIOClientChat() {
	commands := `Commands:
  resources                - List of all available resources
  tools                	   - List of all available tools
  help                     - Show help
  quit                     - Exit`

	logger := utilsLogger.NewLoggerService("stdio-client-chat")
	clientInfo := types.Implementation{}
	clientInfo.Name = "stdio-client-chat"
	clientInfo.Title = "STDIO client chat"
	clientInfo.Version = "1.0.0"
	mcpClient, err := MCPClient.NewClient(clientInfo, MCPClient.ClientOptions{
		Capabilities: types.ClientCapabilities{
			Sampling: struct{}{},
		},
	})
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("MCPClient.NewClient %v", err))
	}
	mcpClient.SetOnErrorCallBack(func(err error) {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	})
	transport := MCPClient.NewStdioClientTransport(MCPClient.StdioServerParameters{
		Command: "go",
		Args:    []string{"run", ".", "-t", "server"},
	}, false)

	ctx := context.Background()
	err = mcpClient.Connect(ctx, transport, nil)
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mcpClient.Connect %v", err))
	}

	_, err = mcpClient.Ping(&shared.RequestOptions{})
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.Ping %v", err))
	}

	// Create GenAI client
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		logger.Fatal(nil, fmt.Sprintf("GenAI Api key is empty use the env <GOOGLE_API_KEY>%v", err))
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		logger.Error(nil, fmt.Sprintf("genai.NewClient %v", err))
	}

	// Create a new Chat.
	modelName := "gemini-2.5-flash"
	genaiTools, err := mcpToolsToGenAITools(mcpClient, logger)
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpToolsToGenAITools %v", err))
		return
	}

	genAIResources, err := mcpResourcesToGenAIResource(mcpClient, logger)
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpResourcesToGenAIResource %v", err))
		return
	}

	chat, err := client.Chats.Create(ctx, modelName, &genai.GenerateContentConfig{
		Temperature: genai.Ptr[float32](0.5),
		Tools:       genaiTools,
		SystemInstruction: &genai.Content{
			Role: "system",
			Parts: []*genai.Part{
				&genai.Part{
					Text: "You are a helpful assistant that helps the user to find information and answer questions using the available tools and resources.",
				},
				&genai.Part{
					Text: genAIResources,
				},
			},
		},
	}, nil)
	if err != nil {
		logger.Error(nil, fmt.Sprintf("client.Chats.Create  %v", err))
	}

	mcpClient.SetNotificationHandler(types.NewResourceListChangedNotification(nil), func(ctx context.Context, notification types.NotificationInterface) error {
		logger.Info(nil, "[Resource list changed ] Notification received!")
		resourcesResult, err := mcpClient.ListResources(types.PaginatedRequestParams{}, nil)
		if err != nil {
			logger.Error(nil, fmt.Sprintf("mcpClient.ListResources in ResourceListChangedNotification %v", err))
		}
		logger.Info(nil, fmt.Sprintf("[Resource list changed] Available resources count: %d", len(resourcesResult.Resources)))
		return nil
	})

	mcpClient.SetRequestHandler(types.NewCreateMessageRequest(nil), func(request types.RequestInterface, extra *shared.RequestHandlerExtra) (types.ResultInterface, error) {
		logger.Info(nil, "[Create message] Request received!")
		req, okType := request.(*types.CreateMessageRequest)
		if !okType {
			err := fmt.Errorf("invalid request type CreateMessageRequest")
			return nil, err
		}
		var finalMsg string
		for _, ms := range req.Params.Messages {
			switch ms.Content.TypeOfContent() {
			case types.TEXT_CONTENT_TYPE:
				realMsg := ms.Content.(*types.TextContent)
				finalMsg += realMsg.Text
			default:
				realMsg := ms.Content.(*types.TextContent)
				finalMsg += realMsg.Text
			}
		}

		chatResult, err := chat.SendMessage(ctx, genai.Part{Text: finalMsg})
		if err != nil {
			return nil, fmt.Errorf("chat.SendMessage %v", err)
		}

		result := &types.CreateMessageResult{
			Model:      chatResult.ModelVersion,
			StopReason: "endTurn",
		}
		result.Role = "assistant"

		var resultContent string
		for _, rc := range chatResult.Candidates {
			if rc.Content == nil {
				continue
			}
			for _, cp := range rc.Content.Parts {
				resultContent += cp.Text
			}
		}

		result.Content = types.NewTextContent(resultContent)
		promptTokens, candidatesTokens, totalTokens := getTokenUsage(chatResult)
		logger.Info(utilsLogger.LogFields{
			"prompt":     promptTokens,
			"candidates": candidatesTokens,
			"total":      totalTokens,
		}, "[Create message] tokens used")
		return result, nil
	})

	fmt.Println("-----------------------MENU----------------------------------")
	fmt.Println("Simple MCP STDio Client Started")
	fmt.Println("Connecting to local MCP server")
	fmt.Println(commands)
	fmt.Println("Or ask your questions naturally!")
	fmt.Println("------------------------MENU---------------------------------")

	for {
		var query string
		fmt.Printf("\nQuery: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			logger.Fatal(nil, "Failed to read input")
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			// Handle potential errors during scanning
			logger.Fatal(nil, fmt.Sprintf("Error reading from stdin %v", err))
		}
		query = scanner.Text()
		if query == "" {
			continue
		}
		switch strings.ToLower(query) {
		case "resources":
			fmt.Println("--------------------------Resource List--------------------------------")
			err := resourceList(mcpClient, logger)
			if err != nil {
				logger.Error(nil, fmt.Sprintf("resourcesList %v", err))
			}
			fmt.Println("--------------------------Resource List--------------------------------")
		case "tools":
			fmt.Println("--- Tool List ---")
			fmt.Println("--------------------------Tool List--------------------------------")
			err := toolList(mcpClient, logger)
			if err != nil {
				logger.Error(nil, fmt.Sprintf("toolList %v", err))
			}
			fmt.Println("--------------------------Tool List--------------------------------")
		case "help":
			fmt.Println("--------------------------Help--------------------------------")
			fmt.Println(commands)
			fmt.Println("--------------------------Help--------------------------------")
		case "quit":
			fmt.Println("--- Quit ---")
			fmt.Println("Goodbye")
			os.Exit(1)
		default:
			fmt.Println("--------------------------Natural questionHelp--------------------------------")
			err := processSimpleChatQuery(ctx, mcpClient, chat, query, logger)
			if err != nil {
				logger.Error(nil, fmt.Sprintf("processSimpleChatQuery %v", err))
			}
			fmt.Println("--------------------------Natural questionHelp--------------------------------")
		}
	}
}

func getTokenUsage(chatResult *genai.GenerateContentResponse) (prompt, candidates, total int32) {
	if chatResult.UsageMetadata == nil {
		return
	}
	prompt = chatResult.UsageMetadata.PromptTokenCount
	candidates = chatResult.UsageMetadata.CandidatesTokenCount
	total = chatResult.UsageMetadata.TotalTokenCount
	return
}

func processSimpleChatQuery(ctx context.Context, mcpClient *MCPClient.Client, chat *genai.Chat, query string, logger utilsLogger.LogService) error {
	result, err := chat.SendMessage(ctx, genai.Part{Text: query})
	if err != nil {
		return fmt.Errorf("chat.SendMessage %v", err)
	}
	maxIterations := 2 // prevent infinite loops
	promptTokens, candidatesTokens, totalTokens := getTokenUsage(result)
	for i := 0; i < maxIterations; i++ {
		if len(result.Candidates) == 0 {
			logger.Error(nil, fmt.Sprintf("no candidates in response  %v", err))
			continue
		}

		candidate := result.Candidates[0]
		if candidate.Content == nil {
			break
		}

		// Check for function calls
		hasFunctionCall := false
		var functionResponses []genai.Part

		for _, part := range candidate.Content.Parts {
			if part == nil {
				continue
			}
			partFunCall := part.FunctionCall
			switch true {
			case partFunCall != nil: // If part is a function call
				hasFunctionCall = true
				functionResponses = append(functionResponses, processFunctionCallGeanAI(mcpClient, partFunCall))
				continue
			case strings.Contains(part.Text, "[RESOURCE:"):
				responseText := part.Text
				start := strings.Index(responseText, "[RESOURCE:")
				end := strings.Index(responseText[start:], "]")
				if end == -1 {
					continue
				}
				uri := strings.TrimSpace(responseText[start+10 : start+end])
				fmt.Printf("\n🤖 AI: Fetching resource: %s\n", uri)
				rrContent, err := readResource(mcpClient, uri, logger)
				if err != nil {
					fmt.Printf("\n❌ Error: [readResource]%v\n", err)
					continue
				}
				followUp := fmt.Sprintf("Here's the content from %s:\n\n%s\n", uri, rrContent)
				result, err = chat.SendMessage(ctx, genai.Part{Text: followUp})
				if err != nil {
					return fmt.Errorf("send function response error: %w", err)
				}
				promptTokensL, candidatesTokensL, totalTokensL := getTokenUsage(result)
				promptTokens += promptTokensL
				candidatesTokens += candidatesTokensL
				totalTokens += totalTokensL

			}
			fmt.Printf("\n🤖 Assistant: %s\n", part.Text)
		}

		if hasFunctionCall {
			// Send function responses back
			result, err = chat.SendMessage(ctx, functionResponses...)
			if err != nil {
				return fmt.Errorf("send function response error: %w", err)
			}
			promptTokensL, candidatesTokensL, totalTokensL := getTokenUsage(result)
			promptTokens += promptTokensL
			candidatesTokens += candidatesTokensL
			totalTokens += totalTokensL
		}
	}
	logger.Info(utilsLogger.LogFields{
		"prompt":     promptTokens,
		"candidates": candidatesTokens,
		"total":      totalTokens,
	}, "[ProcessSimpleChatQuery] tokens used")
	return nil
}

func mcpResourcesToGenAIResource(mcpClient *MCPClient.Client, logger utilsLogger.LogService) (string, error) {
	systemPrompt := "You are a helpful assistant with access to the following resources:\n\n"

	resourceList, err := mcpClient.ListResources(types.PaginatedRequestParams{}, &shared.RequestOptions{})
	if err != nil {
		return "", fmt.Errorf("mcpClient.ListResources %v", err)
	}

	for _, r := range resourceList.Resources {
		systemPrompt += fmt.Sprintf(`- {"name":"%s","uri":"%s","mime":"%s","descrip":"%s"}\n`, r.Name, r.URI, r.MIMEType, r.Description)
	}
	systemPrompt += "\nWhen you need information from these resources, clearly state which resource you want to access using the format: [RESOURCE:uri]"
	return systemPrompt, nil
}

func mcpToolsToGenAITools(mcpClient *MCPClient.Client, logger utilsLogger.LogService) ([]*genai.Tool, error) {
	tools := []*genai.Tool{}
	toolList, err := mcpClient.ListTools(types.PaginatedRequestParams{}, &shared.RequestOptions{})
	if err != nil {
		return tools, fmt.Errorf("mcpClient.ListTools %v", err)
	}

	getGeminiType := func(mcpType string) genai.Type {
		var schemaType genai.Type
		switch mcpType {
		case "string":
			schemaType = genai.TypeString
		case "integer", "number":
			schemaType = genai.TypeNumber
		case "boolean":
			schemaType = genai.TypeBoolean
		case "array":
			schemaType = genai.TypeArray
		case "object":
			schemaType = genai.TypeObject
		default:
			schemaType = genai.TypeString // fallback
		}
		return schemaType
	}

	for _, t := range toolList.Tools {
		genaiToolProperties := map[string]*genai.Schema{}
		for key, property := range t.InputSchema.Properties {
			genaiToolProperties[key] = &genai.Schema{
				Type:        getGeminiType(property.Type),
				Description: property.Description,
			}
		}
		//Default response schema
		responseShema := &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"content": {
					Type:        genai.TypeString,
					Description: "The tool execution result content",
				},
			},
		}
		if t.OutputSchema != nil {
			genaiToolResponse := map[string]*genai.Schema{}
			for key, property := range t.OutputSchema.Properties {
				genaiToolResponse[key] = &genai.Schema{
					Type:        getGeminiType(property.Type),
					Description: property.Description,
				}
			}
			responseShema = &genai.Schema{
				Type:       getGeminiType(t.OutputSchema.Type),
				Properties: genaiToolResponse,
				Required:   t.OutputSchema.Required,
			}
		}

		tool := genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				&genai.FunctionDeclaration{
					Name:        t.Name,
					Description: t.Description,
					Parameters: &genai.Schema{
						Type:       getGeminiType(t.InputSchema.Type),
						Properties: genaiToolProperties,
						Required:   t.InputSchema.Required,
					},
					Response: responseShema,
				},
			},
		}
		tools = append(tools, &tool)
	}
	return tools, nil
}

func processFunctionCallGeanAI(mcpClient *MCPClient.Client, partFunCall *genai.FunctionCall) genai.Part {
	var functionResponse genai.Part
	fmt.Printf("🔧 Tool Call: %s(%v)\n", partFunCall.Name, partFunCall.Args)

	// Execute MCP tool
	result, err := callTool(mcpClient, partFunCall.Name, partFunCall.Args)
	if err != nil {
		fmt.Printf("❌ Error: [callTool]%v\n", err)
		functionResponse = genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				Name: partFunCall.Name,
				Response: map[string]interface{}{
					"error": err.Error(),
				},
			},
		}
	} else {
		fmt.Printf("✅ Result received\n\n")
		functionResponse = genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				Name:     partFunCall.Name,
				Response: result,
			},
		}
	}
	return functionResponse
}

func resourceList(mcpClient *MCPClient.Client, logger utilsLogger.LogService) error {
	resourceList, err := mcpClient.ListResources(types.PaginatedRequestParams{}, &shared.RequestOptions{})
	if err != nil {
		return fmt.Errorf("mcpClient.ListResources %v", err)
	}
	for _, r := range resourceList.Resources {
		fmt.Printf("- [%s] %s\n", r.URI, r.Name)
	}
	return nil
}

func readResource(mcpClient *MCPClient.Client, URI string, logger utilsLogger.LogService) (string, error) {
	var content string
	rrResult, err := mcpClient.ReadResource(types.ReadResourceRequestParams{URI: URI}, &shared.RequestOptions{})
	if err != nil {
		return content, fmt.Errorf("mcpClient.ReadResource %v", err)
	}
	if rrResult == nil {
		return content, fmt.Errorf("content of resource URI: %s is empty", URI)
	}
	for _, rc := range rrResult.Contents {
		strContent := rc.GetContent()
		content += strContent
	}
	return content, nil
}

func toolList(mcpClient *MCPClient.Client, logger utilsLogger.LogService) error {
	toolList, err := mcpClient.ListTools(types.PaginatedRequestParams{}, &shared.RequestOptions{})
	if err != nil {
		return fmt.Errorf("mcpClient.ListTools %v", err)
	}
	for _, r := range toolList.Tools {
		fmt.Printf("- [%s] %s\n", r.Name, r.Description)
	}
	return nil
}

func callTool(mcpClient *MCPClient.Client, toolName string, input map[string]interface{}) (map[string]interface{}, error) {
	result, err := mcpClient.CallTool(types.CallToolRequestParams{
		Name:      toolName,
		Arguments: input,
	}, &shared.RequestOptions{})

	if err != nil {
		return nil, fmt.Errorf("mcpClient.CallTool %v", err)
	}

	dataResult := result.StructuredContent
	if result.StructuredContent == nil {
		dataResult = make(map[string]interface{})
		for _, content := range result.Content {
			switch content.TypeOfContent() {
			case types.TEXT_CONTENT_TYPE:
				dataResult["content"] = content.(*types.TextContent).Text
			case types.IMAGE_CONTENT_TYPE:
				dataResult["content"] = content.(*types.ImageContent).Data
			case types.AUDIO_CONTENT_TYPE:
				dataResult["content"] = content.(*types.AudioContent).Data
			}
		}
	}
	return dataResult, nil
}
