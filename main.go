package main

import (
	"encoding/json"
	"fmt"

	mcptypes "github.com/victorvbello/gomcp/mcp/types"
)

func main() {
	var r mcptypes.Request
	var rs string = `{"method":"otherMethod","params":{"_meta":{"progressToken":"otherToken"},"exampleKey2":"exampleValue2"}} `
	r.Method = "exampleMethod"
	r.Params = &mcptypes.RequestParams{
		Metadata: &mcptypes.MetadataRequest{
			ProgressToken: mcptypes.ProgressToken("exampleToken"),
		},
		AdditionalProperties: map[string]interface{}{
			"_meta":      "notAllowed",
			"exampleKey": "exampleValue",
		},
	}

	b, err := json.Marshal(r)
	fmt.Println(string(b), err)

	err = json.Unmarshal([]byte(rs), &r)
	fmt.Println(r.Method, r.Params.Metadata, r.Params.AdditionalProperties, err)

	fmt.Println("-- Notification --")

	var notify mcptypes.Notification
	var ns string = `{"method":"otherNotificationMethod","params":{"_meta":{"progressToken":"otherToken"},"exampleKey2":"exampleValue2"}} `
	notify.Method = "exampleNotificationMethod"
	notify.Params = &mcptypes.NotificationParams{
		Metadata: map[string]interface{}{
			"exampleKey": "exampleValue",
		},
		AdditionalProperties: map[string]interface{}{
			"_meta":       "notAllowed",
			"exampleKey2": "exampleValue2",
		},
	}

	b, err = json.Marshal(notify)
	fmt.Println(string(b), err)

	err = json.Unmarshal([]byte(ns), &notify)
	fmt.Println(notify.Method, notify.Params.Metadata, notify.Params.AdditionalProperties, err)

	fmt.Println("-- Result --")
	var result mcptypes.Result
	var rs2 string = `{"_meta":{"progressToken":"otherToken"},"exampleKey2":"exampleValue2"}`
	result.Metadata = map[string]interface{}{
		"exampleKey": "exampleValue",
	}
	result.AdditionalProperties = map[string]interface{}{
		"_meta":       "XXX",
		"exampleKey2": "exampleValue2",
	}

	b, err = json.Marshal(&result)
	fmt.Println(string(b), err)

	err = json.Unmarshal([]byte(rs2), &result)
	fmt.Println(result.Metadata, result.AdditionalProperties, err)
}
