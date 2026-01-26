package client

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	MCPClient "github.com/victorvbello/gomcp/mcp/client"
	"github.com/victorvbello/gomcp/mcp/shared"
	"github.com/victorvbello/gomcp/mcp/types"
	"github.com/victorvbello/gomcp/mcp/utils"
	utilsLogger "github.com/victorvbello/gomcp/mcp/utils/logger"
)

func ExampleEverythingWithSTDIOClient() {
	var notificationCount int
	logger := utilsLogger.NewLoggerService("everything-with-stdio-client")
	clientInfo := types.Implementation{}
	clientInfo.Name = "everything-with-stdio-client"
	clientInfo.Title = "Everything whit stdio client"
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
	}, true)
	logger.Info(nil, "MCP client is running...")

	mcpClient.SetNotificationHandler(types.NewLoggingMessageNotification(nil), func(ctx context.Context, notification types.NotificationInterface) error {
		notificationCount++
		loggNotification := notification.(*types.LoggingMessageNotification)
		logger.Info(nil, fmt.Sprintf("Notification #%d: %s - %#v", notificationCount, loggNotification.Params.Level, loggNotification.Params.Data))
		return nil
	})
	mcpClient.SetNotificationHandler(types.NewResourceListChangedNotification(nil), func(ctx context.Context, notification types.NotificationInterface) error {
		logger.Info(nil, "[Resource list changed ] Notification received!")
		resourcesResult, err := mcpClient.ListResources(types.PaginatedRequestParams{}, nil)
		if err != nil {
			logger.Error(nil, fmt.Sprintf("mcpClient.ListResources in ResourceListChangedNotification %v", err))
		}
		if resourcesResult == nil {
			logger.Error(nil, "resourcesResult is nil in ResourceListChangedNotification")
			return nil
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

		result := &types.CreateMessageResult{
			StopReason: "endTurn",
		}
		result.Role = "assistant"

		result.Content = types.NewTextContent("local-testing-response")
		return result, nil
	})
	ctx := context.Background()
	err = mcpClient.Connect(ctx, transport, nil)
	if err != nil {
		logger.Fatal(nil, fmt.Sprintf("mcpClient.Connect %v", err))
	}
	//Ping
	logger.Info(nil, "Ping")
	pingResult, err := mcpClient.Ping(&shared.RequestOptions{})
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.Ping %v", err))
	}
	logger.Info(nil, fmt.Sprintf("Ping result: %v", pingResult))

	//Set Logging Level
	logger.Info(nil, "Set Logging Level")
	logLevelResult, err := mcpClient.SetLoggingLevel(types.LOGGING_LEVEL_DEBUG, &shared.RequestOptions{})
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.SetLoggingLevel %v", err))
	}
	logger.Info(nil, fmt.Sprintf("SetLoggingLevel result: %v", logLevelResult))

	//Get resources list
	logger.Info(nil, "Get resources list")
	resourceListURIMap := make(map[string]string)
	resourceList, err := mcpClient.ListResources(types.PaginatedRequestParams{}, &shared.RequestOptions{})
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.ListResources %v", err))
	}
	logger.Info(utilsLogger.LogFields{"count": len(resourceList.Resources)}, "Get resources list")
	for _, r := range resourceList.Resources {
		logger.Info(utilsLogger.LogFields{"name": r.Name, "uri": r.URI}, "Resource")
		resourceRead, err := mcpClient.ReadResource(types.ReadResourceRequestParams{URI: r.URI}, &shared.RequestOptions{})
		if err != nil {
			logger.Error(nil, fmt.Sprintf("mcpClient.ReadResource %v", err))
		}
		logger.Info(nil, "Contents")
		for _, content := range resourceRead.Contents {
			strContent := content.GetContent()
			crb := content.GetBaseResourceContents()
			logger.Info(utilsLogger.LogFields{"mime": crb.MIMEType, "content": strContent[0:20] + "..."}, fmt.Sprintf("[%s]", crb.URI))
		}
		resourceListURIMap[r.Name] = r.URI
	}

	//Get resources template list
	logger.Info(nil, "Get resources template list")
	resourceListTemplates, err := mcpClient.ListResourceTemplates(types.PaginatedRequestParams{}, &shared.RequestOptions{})
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.ListResourceTemplates %v", err))
	}
	logger.Info(utilsLogger.LogFields{"count": len(resourceListTemplates.ResourceTemplates)}, "Get resources template list")
	for _, r := range resourceListTemplates.ResourceTemplates {
		logger.Info(utilsLogger.LogFields{"name": r.Name, "uri-template": r.URITemplate}, "Resource")
		uriTmp, err := utils.NewUriTemplate(r.URITemplate)
		if err != nil {
			logger.Error(nil, fmt.Sprintf("utils.NewUriTemplate %s error: %v", r.URITemplate, err))
		}
		finalURI, err := uriTmp.Expand(map[string][]string{"user_id": []string{"44"}})
		if err != nil {
			logger.Error(nil, fmt.Sprintf("uriTmp.Expand %s error: %v", r.URITemplate, err))
		}
		resourceRead, err := mcpClient.ReadResource(types.ReadResourceRequestParams{URI: finalURI}, &shared.RequestOptions{})
		if err != nil {
			logger.Error(nil, fmt.Sprintf("mcpClient.ReadResource %v", err))
		}
		logger.Info(nil, "Contents")
		for _, content := range resourceRead.Contents {
			strContent := content.GetContent()
			crb := content.GetBaseResourceContents()
			logger.Info(utilsLogger.LogFields{"mime": crb.MIMEType, "content": strContent[0:20] + "..."}, fmt.Sprintf("[%s]", crb.URI))
		}
	}

	//Get prompt list
	logger.Info(nil, "Get prompt list")
	promptList, err := mcpClient.ListPrompts(types.PaginatedRequestParams{}, &shared.RequestOptions{})
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.ListPrompts %v", err))
	}
	logger.Info(utilsLogger.LogFields{"count": len(promptList.Prompts)}, "Get prompt list")
	for _, p := range promptList.Prompts {
		logger.Info(utilsLogger.LogFields{"name": p.Name}, "Prompt")
		argJson := []map[string]string{
			map[string]string{
				"id":   "1",
				"name": "name 1",
			},
			map[string]string{
				"id":   "2",
				"name": "name 2",
			},
		}
		argJsonB, err := json.Marshal(argJson)
		if err != nil {
			logger.Error(nil, fmt.Sprintf("json.Marshal argJson %v", err))
		}
		getPromptResult, err := mcpClient.GetPrompt(
			types.GetPromptParams{
				Name:      p.Name,
				Arguments: map[string]string{"json": string(argJsonB), "kind": "user"},
			},
			&shared.RequestOptions{})
		if err != nil {
			logger.Error(nil, fmt.Sprintf("mcpClient.GetPrompt %v", err))
		}
		logger.Info(nil, "Messages")
		for _, pmsg := range getPromptResult.Messages {
			switch pmsg.Content.TypeOfContent() {
			case types.TEXT_CONTENT_TYPE:
				content := pmsg.Content.(*types.TextContent)
				logger.Info(utilsLogger.LogFields{"content": content.Text[0:20] + "..."}, fmt.Sprintf("[%s]", pmsg.Role))
			}
		}
	}

	//Get tools list
	logger.Info(nil, "Get tools list")
	toolsList, err := mcpClient.ListTools(types.PaginatedRequestParams{}, &shared.RequestOptions{})
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.ListTools %v", err))
	}
	logger.Info(utilsLogger.LogFields{"count": len(toolsList.Tools)}, "Get tools list")
	for _, tool := range toolsList.Tools {
		logger.Info(utilsLogger.LogFields{"name": tool.Name}, "Tool")
	}

	//Call tool sum
	logger.Info(nil, "Get tool sum")
	sumToolResult, err := mcpClient.CallTool(
		types.CallToolRequestParams{
			Name: "sum",
			Arguments: map[string]interface{}{
				"a": 2,
				"b": 3,
			},
		},
		&shared.RequestOptions{},
	)
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.CallTool sum %v", err))
	}
	logger.Info(nil, fmt.Sprintf("The result of sum [%v] and [%v] is [%v]",
		sumToolResult.StructuredContent["a"],
		sumToolResult.StructuredContent["b"],
		sumToolResult.StructuredContent["result"]))

	//Call tool check-new-resources
	logger.Info(nil, "Get tool check-new-resources")
	_, err = mcpClient.CallTool(
		types.CallToolRequestParams{
			Name: "check-new-resources",
		},
		&shared.RequestOptions{},
	)
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.CallTool check-new-resources %v", err))
	}

	//Call tool check-the-weather-today
	logger.Info(nil, "Get tool check-the-weather-today")
	weatherResult, err := mcpClient.CallTool(
		types.CallToolRequestParams{
			Name: "check-the-weather-today",
			Arguments: map[string]interface{}{
				"city": "Santiago",
			},
		},
		&shared.RequestOptions{},
	)
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.CallTool check-the-weather-today %v", err))
	}
	for _, content := range weatherResult.Content {
		contentType := content.TypeOfContent()
		if contentType != types.TEXT_CONTENT_TYPE {
			continue
		}
		logger.Info(utilsLogger.LogFields{"content": content.(*types.TextContent).Text}, "The result of check-the-weather-today tool is:")
	}

	//Call tool generate-wealthy-plan
	logger.Info(nil, "Get tool generate-wealthy-plan")
	wealthyPlanResult, err := mcpClient.CallTool(
		types.CallToolRequestParams{
			Name: "generate-wealthy-plan",
			Arguments: map[string]interface{}{
				"salary":    "3000$",
				"frequency": "Monthly",
			},
		},
		&shared.RequestOptions{},
	)
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.CallTool generate-wealthy-plan %v", err))
	}
	for _, content := range wealthyPlanResult.Content {
		contentType := content.TypeOfContent()
		if contentType != types.TEXT_CONTENT_TYPE {
			continue
		}
		logger.Info(utilsLogger.LogFields{"content": content.(*types.TextContent).Text}, "The result of generate-wealthy-plan tool is:")
	}

	// Notification handler for the resource terms-and-conditions.txt when it has been updated
	mcpClient.SetNotificationHandler(types.NewResourceUpdatedNotification(&types.ResourceUpdatedNotificationParams{URI: resourceListURIMap["terms-and-conditions"]}), func(ctx context.Context, notification types.NotificationInterface) error {
		logger.Info(nil, "[Resource terms-and-conditions updated] Notification received!")
		logger.Info(nil, "[Resource terms-and-conditions updated] Method:"+notification.GetNotification().Method)

		_, err = mcpClient.UnsubscribeResource(types.UnsubscribeRequestParams{URI: resourceListURIMap["terms-and-conditions"]}, nil)
		if err != nil {
			logger.Error(nil, fmt.Sprintf("mcpClient.UnsubscribeResource terms-and-conditions %v", err))
		}
		return nil
	})

	// Resource subscribe terms-and-conditions
	logger.Info(nil, "Resource subscribe terms-and-conditions")
	_, err = mcpClient.SubscribeResource(types.SubscribeRequestParams{URI: resourceListURIMap["terms-and-conditions"]}, &shared.RequestOptions{})
	if err != nil {
		logger.Error(nil, fmt.Sprintf("mcpClient.SubscribeResource terms-and-conditions %v", err))
	}
	time.Sleep(4 * time.Second)
	mcpClient.Close()
}
