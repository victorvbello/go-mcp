package shared

import "github.com/victorvbello/gomcp/mcp/types"

type StreamID string
type EventID string
type ReplayEventsAfterSend func(eventID EventID, message types.JSONRPCMessage)

//Interface for resumability support via event storage
type EventStore interface {
	//Stores an event for later retrieval
	//@param streamId ID of the stream the event belongs to
	//@param message The JSON-RPC message to store
	//@returns The generated event ID for the stored event
	StoreEvent(streamId StreamID, message types.JSONRPCMessage) (EventID, error)
	ReplayEventsAfter(lastEventID EventID, send ReplayEventsAfterSend) (StreamID, error)
}
