package types

//Represents a root directory or file that the server can operate on.
type Root struct {
	//The URI identifying the root. This *must* start with file://for now.
	//This restriction may be relaxed in future versions of the protocol to allow
	//other URI schemes.
	URI string `json:"uri"`
	//An optional name for the root. This can be used to provide a human-readable
	//identifier for the root, which may be useful for display purposes or for
	//referencing the root in other parts of the application.
	Name string `json:"name,omitempty"`
}

//Sent from the server to request a list of root URIs from the client. Roots allow
//servers to ask for specific directories or files to operate on. A common example
//for roots is providing a set of repositories or directories a server should operate
//on.
//
//This request is typically used when the server needs to understand the file system
//structure or access specific locations that the client has permission to read from.
//
//Only method: METHOD_LIST_ROOTS
type ListRootsRequest struct {
	Request
}

func (lrr *ListRootsRequest) TypeOfServerRequest() int { return LIST_ROOTS_REQUEST_SERVER_REQUEST_TYPE }

//The client's response to a roots/list request from the server.
//This result contains an array of Root objects, each representing a root directory
//or file that the server can operate on.
type ListRootsResult struct {
	Result
	Roots []Root `json:"roots"`
}

func (lrr *ListRootsResult) TypeOfClientResult() int { return LIST_ROOTS_RESULT_CLIENT_RESULT_TYPE }
func (lrr *ListRootsResult) TypeOfResultInterface() int {
	return LIST_ROOTS_RESULT_RESULT_INTERFACE_TYPE
}

//A notification from the client to the server, informing it that the list of roots has changed.
//This notification should be sent whenever the client adds, removes, or modifies any root.
//The server should then request an updated list of roots using the ListRootsRequest.
//
//Only method: METHOD_NOTIFICATION_ROOTS_LIST_CHANGED
type RootsListChangedNotification struct {
	Notification
}

func (rln *RootsListChangedNotification) TypeOfClientNotification() int {
	return ROOTS_LIST_CHANGED_NOTIFICATION_CLIENT_NOTIFICATION_TYPE
}
