package enums

// EventType represents the type of an event that can be sent over ZMQ.
type EventType string

const (
	EventTypeSessionStart   EventType = "session_start"
	EventTypeSessionRunning EventType = "session_running"
	EventTypeSessionEnd     EventType = "session_end"
	EventTypeSessionFailed  EventType = "session_failed"
)

// String returns the string representation of the EventType.
func (e EventType) String() string {
	return string(e)
}
