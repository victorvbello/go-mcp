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
		fmt.Println("client SetOnErrorCallBack", err)
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	})

	transport := MCPClient.NewStdioClientTransport(MCPClient.StdioServerParameters{
		Command: "go",
		Args:    []string{"run", ".", "-t", "server"},
	})
	logger.Info(nil, "MCP client is running...")

	mcpClient.SetNotificationHandler(types.NewLoggingMessageNotification(nil), func(ctx context.Context, notification types.NotificationInterface) error {
		notificationCount++
		loggNotification := notification.(*types.LoggingMessageNotification)
		logger.Info(nil, fmt.Sprintf("Notification #%d: %s - %#v", notificationCount, loggNotification.Params.Level, loggNotification.Params.Data))
		return nil
	})
	mcpClient.SetNotificationHandler(types.NewResourceListChangedNotification(nil), func(ctx context.Context, notification types.NotificationInterface) error {
		logger.Info(nil, "Resource list changed notification received!")
		resourcesResult, err := mcpClient.ListResources(types.PaginatedRequestParams{}, nil)
		if err != nil {
			logger.Error(nil, fmt.Sprintf("mcpClient.ListResources in ResourceListChangedNotification %v", err))
		}
		logger.Info(nil, fmt.Sprintf("Available resources count: %d", len(resourcesResult.Resources)))
		return nil
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

	//Call tool sum
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

	time.Sleep(2 * time.Second)
	mcpClient.Close()
}
