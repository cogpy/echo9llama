package llm

// DoneReason indicates why a completion finished
type DoneReason int

const (
	// DoneReasonStop indicates the completion stopped naturally
	DoneReasonStop DoneReason = iota
	// DoneReasonLength indicates the completion stopped due to length limits
	DoneReasonLength
	// DoneReasonConnectionClosed indicates the completion stopped due to the connection being closed
	DoneReasonConnectionClosed
)

// String returns the string representation of DoneReason
func (d DoneReason) String() string {
	switch d {
	case DoneReasonStop:
		return "stop"
	case DoneReasonLength:
		return "length"
	case DoneReasonConnectionClosed:
		return "connection_closed"
	default:
		return "unknown"
	}
}

// ServerStatus represents the status of a model server
type ServerStatus int

const (
	ServerStatusReady ServerStatus = iota
	ServerStatusLoading
	ServerStatusError
)

// String returns the string representation of ServerStatus
func (s ServerStatus) String() string {
	switch s {
	case ServerStatusReady:
		return "ready"
	case ServerStatusLoading:
		return "loading"
	case ServerStatusError:
		return "error"
	default:
		return "unknown"
	}
}
