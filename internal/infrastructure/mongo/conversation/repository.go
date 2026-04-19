package conversation

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

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

func (r *repository) GetDirect(ctx context.Context, userAID int64, userBID int64) (*chat.Conversation, error) {
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

func (r *repository) GetByClientConvID(ctx context.Context, clientConvID string) (*chat.Conversation, error) {
	return r.getByID(ctx, bson.M{fieldClientConvID: clientConvID})

}

func (r *repository) GetByID(ctx context.Context, id string) (*chat.Conversation, error) {
	return r.getByID(ctx, bson.M{fieldID: id})
}

func (r *repository) GetList(ctx context.Context, params chat.GetConversationsParams) ([]chat.Conversation, error) {
	filter := bson.M{
		fieldParticipants: bson.M{
			"$elemMatch": bson.M{
				fieldUserID: params.UserID,
				fieldState:  string(params.State),
			},
		},
	}

	if params.Query.Cursor != "" {
		cursorTime, err := params.GetTimeFromCursor()
		if err != nil {
			return nil, pkg.ErrBadRequest
		}
		filter[fieldUpdatedAt] = bson.M{"$lt": cursorTime}
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: fieldUpdatedAt, Value: -1}, // newest
			{Key: fieldID, Value: -1},
		}).
		SetLimit(int64(params.Query.GetFetchLimit()))

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, pkg.MapMongoError(err)
	}
	defer cursor.Close(ctx)

	var docs []conversationDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, pkg.MapMongoError(err)
	}
	var res []chat.Conversation
	for _, doc := range docs {
		res = append(res, *doc.toDomain())
	}
	return res, nil
}

func (r *repository) AddParticipant(ctx context.Context, convID string, p chat.Participant) error {
	return nil
}

func (r *repository) RemoveParticipant(ctx context.Context, convID string, userID int64) error {
	return nil
}

func (r *repository) UpdateParticipantState(ctx context.Context, convID string, userID int64, state chat.ParticipantState) error {
	filter := bson.M{
		fieldID:                     convID,
		fieldParticipantUserID_Find: userID,
	}
	update := bson.M{"$set": bson.M{
		fieldUpdatedAt:               pkg.TimeNowUTC(),
		fieldParticipantState_Update: string(state),
	}}
	if _, err := r.coll.UpdateOne(ctx, filter, update); err != nil {
		return pkg.MapMongoError(err)
	}
	return nil
}

func (r *repository) UpdateParticipantRole(ctx context.Context, convID string, userID int64, role chat.ParticipantRole) error {
	return nil
}

func (r *repository) UpdateGroupInfo(ctx context.Context, convID string, name *string, avatarKey *string) error {
	set := bson.M{fieldUpdatedAt: pkg.TimeNowUTC()}
	if name != nil {
		set[fieldName] = *name
	}
	if avatarKey != nil {
		set[fieldAvatarKey] = *avatarKey
	}
	if _, err := r.coll.UpdateOne(ctx, bson.M{fieldID: convID}, bson.M{"$set": set}); err != nil {
		return pkg.MapMongoError(err)
	}
	return nil
}

func (r *repository) UpdateLastRead(ctx context.Context, convID string, userID int64, messageID string) error {
	return nil

}

func (r *repository) Touch(ctx context.Context, convID string, lastMsgID string) error {
	return nil

}

func (r *repository) getByID(ctx context.Context, filter any) (*chat.Conversation, error) {
	var doc conversationDoc
	if err := r.coll.FindOne(ctx, filter).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, pkg.MapMongoError(err)
	}
	return doc.toDomain(), nil
}
