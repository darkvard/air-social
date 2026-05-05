package ws

import "encoding/json"

// inbound
const (
	EventPing       = "ping"
	EventTyping     = "typing"
	EventStopTyping = "stop_typing"
	EventJoin       = "join"
	EventLeave      = "leave"
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
	EventNotification    = "notification"
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

type convPayload struct {
	ConversationID string `json:"conversation_id"`
}

func chatChannel(convID string) string {
	return "chat:" + convID
}
