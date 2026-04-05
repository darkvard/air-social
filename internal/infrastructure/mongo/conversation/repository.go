package conversation

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"air-social/internal/domain/chat"
	"air-social/pkg"
)

// todo: $push & $pull for race condition (concurrency)

type repository struct {
	coll *mongo.Collection
}

func NewRepository(db *mongo.Database) *repository {
	return &repository{
		coll: db.Collection("conversations"),
	}
}

func (r *repository) Create(ctx context.Context, conv *chat.Conversation) error {
	doc := fromDomain(conv)
	if _, err := r.coll.InsertOne(ctx, doc); err != nil {
		return pkg.MapMongoError(err)
	}
	return nil
}

func (r *repository) FindDirect(ctx context.Context, userAID int64, userBID int64) (*chat.Conversation, error) {
	filter := bson.M{
		fieldType: string(chat.ConversationDirect),
		fieldParticipantUserID_Find: bson.M{
			"$all": bson.A{userAID, userBID},
		},
		fieldParticipants: bson.M{
			"$size": 2,
		},
	}

	var doc conversationDoc
	if err := r.coll.FindOne(ctx, filter).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, pkg.MapMongoError(err)
	}
	return doc.toDomain(), nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*chat.Conversation, error) {
	return nil, nil
}

func (r *repository) GetParticipantConversations(ctx context.Context, params chat.GetConversationsParams) ([]chat.Conversation, error) {
	return nil, nil
}

func (r *repository) AddParticipant(ctx context.Context, convID string, p chat.Participant) error {
	return nil
}

func (r *repository) RemoveParticipant(ctx context.Context, convID string, userID int64) error {
	return nil
}

func (r *repository) UpdateParticipantState(ctx context.Context, convID string, userID int64, state chat.ParticipantState) error {
	return nil

}

func (r *repository) UpdateParticipantRole(ctx context.Context, convID string, userID int64, role chat.ParticipantRole) error {
	return nil
}

// UpdateGroupInfo only updates non-nil fields — nil means keep existing value.
func (r *repository) UpdateGroupInfo(ctx context.Context, convID string, name *string, avatarKey *string) error {
	return nil

}

func (r *repository) UpdateLastRead(ctx context.Context, convID string, userID int64, messageID string) error {
	return nil

}

func (r *repository) TouchConversation(ctx context.Context, convID string, lastMsgID string) error {
	return nil

}
