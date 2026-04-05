package conversation

import (
	"time"

	"air-social/internal/domain/chat"
)

const (
	fieldID           = "_id"
	fieldType         = "type"
	fieldParticipants = "participants"
	fieldName         = "name"
	fieldAvatarKey    = "avatar_key"
	fieldClientConvID = "client_conv_id"
	fieldCreatedBy    = "created_by"
	fieldLastMsgID    = "last_msg_id"
	fieldCreatedAt    = "created_at"
	fieldUpdatedAt    = "updated_at"
)

const (
	fieldParticipantUserID_Find = "participants.user_id"
	fieldParticipantRole_Find   = "participants.role"
	fieldParticipantState_Find  = "participants.state"
)

const (
	fieldParticipantRole_Update       = "participants.$.role"
	fieldParticipantState_Update      = "participants.$.state"
	fieldParticipantLastReadID_Update = "participants.$.last_read_id"
)

type participantDoc struct {
	UserID     int64     `bson:"user_id"`
	Role       string    `bson:"role"`
	State      string    `bson:"state"`
	JoinedAt   time.Time `bson:"joined_at"`
	LastReadID string    `bson:"last_read_id,omitempty"`
}

func (d *participantDoc) toDomain() chat.Participant {
	return chat.Participant{
		UserID:     d.UserID,
		Role:       chat.ParticipantRole(d.Role),
		State:      chat.ParticipantState(d.State),
		JoinedAt:   d.JoinedAt,
		LastReadID: d.LastReadID,
	}
}

func participantFromDomain(p chat.Participant) participantDoc {
	return participantDoc{
		UserID:     p.UserID,
		Role:       string(p.Role),
		State:      string(p.State),
		JoinedAt:   p.JoinedAt,
		LastReadID: p.LastReadID,
	}
}

type conversationDoc struct {
	ID           string           `bson:"_id"` // ULID
	Type         string           `bson:"type"`
	Participants []participantDoc `bson:"participants"`
	Name         string           `bson:"name,omitempty"`
	AvatarKey    string           `bson:"avatar_key,omitempty"`
	ClientConvID string           `bson:"client_conv_id,omitempty"`
	CreatedBy    int64            `bson:"created_by"`
	LastMsgID    string           `bson:"last_msg_id,omitempty"`
	CreatedAt    time.Time        `bson:"created_at"`
	UpdatedAt    time.Time        `bson:"updated_at"`
}

func (d *conversationDoc) toDomain() *chat.Conversation {
	participants := make([]chat.Participant, len(d.Participants))
	for i, p := range d.Participants {
		participants[i] = p.toDomain()
	}
	return &chat.Conversation{
		ID:           d.ID,
		Type:         chat.ConversationType(d.Type),
		Participants: participants,
		Name:         d.Name,
		AvatarKey:    d.AvatarKey,
		ClientConvID: d.ClientConvID,
		CreatedBy:    d.CreatedBy,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		// LastMessage and UnreadCount are not stored in MongoDB —
		// repo populates them separately after fetching.
	}
}

func fromDomain(c *chat.Conversation) *conversationDoc {
	participants := make([]participantDoc, len(c.Participants))
	for i, p := range c.Participants {
		participants[i] = participantFromDomain(p)
	}
	return &conversationDoc{
		ID:           c.ID,
		Type:         string(c.Type),
		Participants: participants,
		Name:         c.Name,
		AvatarKey:    c.AvatarKey,
		ClientConvID: c.ClientConvID,
		CreatedBy:    c.CreatedBy,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
		// LastMsgID is updated separately after SendMessage.
	}
}
