package ws

import "encoding/json"

// inbound
const (
	EventPing       = "ping"
	EventTyping     = "typing"
	EventStopTyping = "stop_typing"
	EventMarkRead   = "mark_read"
	EventJoin       = "join"
)

// outbound
const (
	EventPong            = "pong"
	EventNewMessage      = "new_message"
	EventMessageEdited   = "message_edited"
	EventMessageDeleted  = "message_deleted"
	EventReactionAdded   = "reaction_added"
	EventReactionRemoved = "reaction_removed"
	EventRead            = "read"
	EventLeave           = "leave"
)

type InboundEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type OutboundEvent struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

func (e OutboundEvent) encode() ([]byte, error) {
	return json.Marshal(e)
}

func chatChannel(convID string) string {
	return "chat:" + convID
}
