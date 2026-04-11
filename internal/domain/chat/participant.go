package chat

import "time"

type ParticipantState string

const (
	StateActive  ParticipantState = "active"
	StatePending ParticipantState = "pending"
	StateIgnored ParticipantState = "ignored"
	StateMuted   ParticipantState = "muted"
)

type ParticipantRole string

const (
	RoleAdmin  ParticipantRole = "admin"
	RoleMember ParticipantRole = "member"
)

type Participant struct {
	UserID     int64
	Role       ParticipantRole  // admin | member; direct conv: both member
	State      ParticipantState // active | pending | ignored | muted
	JoinedAt   time.Time
	LastReadID string // ULID of last read message; "" = never read
}

func (p Participant) HasReadAnyMessage() bool {
	return p.LastReadID != ""
}
