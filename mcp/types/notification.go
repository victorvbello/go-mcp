package types

import (
	"encoding/json"
	"fmt"

	"github.com/victorvbello/gomcp/mcp/methods"
)

const (
	CANCELLED_NOTIFICATION_CLIENT_NOTIFICATION_TYPE = iota + 1
	PROGRESS_NOTIFICATION_CLIENT_NOTIFICATION_TYPE
	INITIALIZED_NOTIFICATION_CLIENT_NOTIFICATION_TYPE
	ROOTS_LIST_CHANGED_NOTIFICATION_CLIENT_NOTIFICATION_TYPE
)

const (
	CANCELLED_NOTIFICATION_SERVER_NOTIFICATION_TYPE = iota + 11
	PROGRESS_NOTIFICATION_SERVER_NOTIFICATION_TYPE
	LOGGING_MESSAGE_NOTIFICATION_SERVER_NOTIFICATION_TYPE
	RESOURCE_UPDATED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
	RESOURCE_LIST_CHANGED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
	TOOL_LIST_CHANGED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
	PROMPT_LIST_CHANGED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
)

const (
	CANCELLED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE = iota + 300
	INITIALIZED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
	PROGRESS_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
	NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
	LOGGING_MESSAGE_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
	RESOURCE_UPDATED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
	RESOURCE_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
	TOOL_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
	PROMPT_LIST_CHANGED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
)

type Notification struct {
	Method string                  `json:"method"`
	Params *BaseNotificationParams `json:"params,omitempty"`
}

type BaseNotificationParams struct {
	//Attach additional metadata to their notifications.
	Metadata map[string]interface{} `json:"_meta,omitempty"`
	//Attach additional properties, _meta is reserved by MCP
	AdditionalProperties map[string]interface{} `json:"-"`
}

func (np *BaseNotificationParams) MarshalJSON() ([]byte, error) {
	raw := make(map[string]interface{})
	if np.Metadata != nil {
		raw["_meta"] = np.Metadata
	}
	for key, value := range np.AdditionalProperties {
		if key == "_meta" {
			continue //Skip the _meta key is reserved by MCP
		}
		raw[key] = value
	}

	return json.Marshal(raw)
}

func (np *BaseNotificationParams) UnmarshalJSON(data []byte) error {
	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	if _, ok := raw["_meta"]; !ok {
		return nil //No _meta field, nothing to unmarshal
	}
	bm, err := json.Marshal(raw["_meta"])
	if err != nil {
		return fmt.Errorf("error marshaling _meta: %v", err)
	}
	if err := json.Unmarshal(bm, &np.Metadata); err != nil {
		return fmt.Errorf("error unmarshaling into metadata: %v", err)
	}
	delete(raw, "_meta")
	np.AdditionalProperties = raw
	return nil
}

func (n *Notification) TypeOfNotification() int {
	return NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
}
func (n *Notification) GetNotification() Notification {
	return *n
}

type NotificationInterface interface {
	TypeOfNotification() int
	GetNotification() Notification
}

//This notification can be sent by either side to indicate that it is cancelling a previously-issued request.
//
//The request SHOULD still be in-flight, but due to communication latency, it is always possible that this notification MAY arrive after the request has already finished.
//
//This notification indicates that the result will be unused, so any associated processing SHOULD cease.
//
//A client MUST NOT attempt to cancel its `initialize` request.
//
//Only method: METHOD_NOTIFICATION_CANCELLED
type CancelledNotification struct {
	Notification
	Params CancelledNotificationParams `json:"params,omitempty"`
}

func (cn *CancelledNotification) TypeOfClientNotification() int {
	return CANCELLED_NOTIFICATION_CLIENT_NOTIFICATION_TYPE
}
func (cn *CancelledNotification) TypeOfServerNotification() int {
	return CANCELLED_NOTIFICATION_SERVER_NOTIFICATION_TYPE
}
func (cn *CancelledNotification) TypeOfNotification() int {
	return CANCELLED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
}
func (cn *CancelledNotification) GetNotification() Notification {
	return cn.Notification
}
func (cn *CancelledNotification) JSONRPCMessageType() int {
	return JSONRPC_MESSAGE_CANCELLED_NOTIFICATION_TYPE
}

type CancelledNotificationParams struct {
	BaseNotificationParams
	//The ID of the request to cancel.
	//
	//This MUST correspond to the ID of a request previously issued in the same direction.
	RequestID RequestID `json:"requestId"`
	//An optional string describing the reason for the cancellation. This MAY be logged or presented to the user.
	Reason string `json:"reason,omitempty"`
}

func (cnp *CancelledNotificationParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		RequestID RequestID `json:"requestId"`
		Reason    string    `json:"reason,omitempty"`
	}{
		RequestID: cnp.RequestID,
		Reason:    cnp.Reason,
	}
	knownFields, err := json.Marshal(&aux)
	if err != nil {
		return nil, fmt.Errorf("marshal known fields: %w", err)
	}
	//Marshal knownFields to map
	baseMap := make(map[string]interface{})
	if err := json.Unmarshal(knownFields, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal known fields to map: %w", err)
	}
	//Marshal base.BaseNotificationParams
	baseExtra, err := cnp.BaseNotificationParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (cnp *CancelledNotificationParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		RequestID RequestID `json:"requestId"`
		Reason    string    `json:"reason,omitempty"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	cnp.RequestID = aux.RequestID
	cnp.Reason = aux.Reason

	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	delete(raw, "requestId")
	delete(raw, "reason")
	bm, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("error marshaling rest of data: %v", err)
	}
	if err := cnp.BaseNotificationParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("BaseNotificationParams.UnmarshalJSON: %w", err)
	}
	return nil
}

func NewCancelledNotification(params *CancelledNotificationParams) *CancelledNotification {
	newCN := CancelledNotification{
		Notification: Notification{Method: methods.METHOD_NOTIFICATION_CANCELLED},
	}
	if params != nil {
		newCN.Params = *params
	}
	return &newCN
}

//This notification is sent from the client to the server after initialization has finished.
//
//Only method: METHOD_NOTIFICATION_INITIALIZED
type InitializedNotification struct {
	Notification
}

func (in *InitializedNotification) TypeOfClientNotification() int {
	return INITIALIZED_NOTIFICATION_CLIENT_NOTIFICATION_TYPE
}
func (in *InitializedNotification) TypeOfNotification() int {
	return INITIALIZED_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
}
func (in *InitializedNotification) GetNotification() Notification {
	return in.Notification
}

func NewInitializedNotification(params *BaseNotificationParams) *InitializedNotification {
	newIn := new(InitializedNotification)
	newIn.Method = methods.METHOD_NOTIFICATION_INITIALIZED
	newIn.Params = params
	return newIn
}

type Progress struct {
	//The progress thus far. This should increase every time progress is made, even if the total is unknown.
	Progress int `json:"progress"`
	//Total number of items to process (or total progress required), if known.
	Total int `json:"total,omitempty"`
	//An optional message describing the current progress.
	Message string `json:"message,omitempty"`
}

//An out-of-band notification used to inform the receiver of a progress update for a long-running request.
//
//Only method: METHOD_NOTIFICATION_PROGRESS
type ProgressNotification struct {
	Notification
	Params ProgressNotificationParams `json:"params"`
}

func (pn *ProgressNotification) TypeOfClientNotification() int {
	return PROGRESS_NOTIFICATION_CLIENT_NOTIFICATION_TYPE
}
func (pn *ProgressNotification) TypeOfServerNotification() int {
	return PROGRESS_NOTIFICATION_SERVER_NOTIFICATION_TYPE
}
func (pn *ProgressNotification) TypeOfNotification() int {
	return PROGRESS_NOTIFICATION_NOTIFICATION_INTERFACE_TYPE
}
func (pn *ProgressNotification) GetNotification() Notification {
	return pn.Notification
}

type ProgressNotificationParams struct {
	BaseNotificationParams
	Progress
	//The progress token which was given in the initial request, used to associate this notification with the request that is proceeding.
	ProgressToken ProgressToken `json:"progressToken"`
}

func (pnp *ProgressNotificationParams) MarshalJSON() ([]byte, error) {
	//bridge struct to marshal known fields
	aux := struct {
		Progress
		ProgressToken ProgressToken `json:"progressToken"`
	}{
		Progress:      pnp.Progress,
		ProgressToken: pnp.ProgressToken,
	}
	knownFields, err := json.Marshal(&aux)
	if err != nil {
		return nil, fmt.Errorf("marshal known fields: %w", err)
	}
	//Marshal knownFields to map
	baseMap := make(map[string]interface{})
	if err := json.Unmarshal(knownFields, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal known fields to map: %w", err)
	}
	//Marshal base.BaseNotificationParams
	baseExtra, err := pnp.BaseNotificationParams.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal base fields: %w", err)
	}
	if err := json.Unmarshal(baseExtra, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}
	return json.Marshal(baseMap)
}

func (pnp *ProgressNotificationParams) UnmarshalJSON(data []byte) error {
	aux := struct {
		Progress
		ProgressToken ProgressToken `json:"progressToken"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
	}
	pnp.Progress = aux.Progress
	pnp.ProgressToken = aux.ProgressToken

	raw := make(map[string]interface{})
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling global data: %v", err)
	}
	delete(raw, "progress")
	delete(raw, "total")
	delete(raw, "message")
	delete(raw, "progressToken")
	bm, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("error marshaling rest of data: %v", err)
	}
	if err := pnp.BaseNotificationParams.UnmarshalJSON(bm); err != nil {
		return fmt.Errorf("BaseNotificationParams.UnmarshalJSON: %w", err)
	}
	return nil
}

func NewProgressNotification(params *ProgressNotificationParams) *ProgressNotification {
	newPN := ProgressNotification{
		Notification: Notification{Method: methods.METHOD_NOTIFICATION_PROGRESS},
	}
	if params != nil {
		newPN.Params = *params
	}
	return &newPN
}

type ClientNotification interface {
	TypeOfClientNotification() int
}

type ServerNotification interface {
	TypeOfServerNotification() int
}
