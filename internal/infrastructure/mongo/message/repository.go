package message

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"air-social/internal/domain/chat"
	"air-social/pkg"
)

type repository struct {
	coll *mongo.Collection
}

func NewRepository(db *mongo.Database) *repository {
	return &repository{
		coll: db.Collection("messages"),
	}
}

func (r *repository) Create(ctx context.Context, msg *chat.Message) error {
	_, err := r.coll.InsertOne(ctx, messageFromDomain(msg))
	return pkg.MapMongoError(err)
}

func (r *repository) AddReaction(ctx context.Context, msgID string, reaction *chat.Reaction) error {
	doc := reactionFromDomain(reaction)
	// Positional $ operator matches the first array element where user_id equals reaction.UserID.
	res, err := r.coll.UpdateOne(ctx,
		bson.M{fieldID: msgID, fieldReactionUserID: reaction.UserID},
		bson.M{"$set": bson.M{"reactions.$": doc}},
	)
	if err != nil {
		return pkg.MapMongoError(err)
	}
	if res.MatchedCount == 0 {
		_, err = r.coll.UpdateOne(ctx, bson.M{fieldID: msgID}, bson.M{"$push": bson.M{fieldReactions: doc}})
		return pkg.MapMongoError(err)
	}
	return nil
}

func (r *repository) RemoveReaction(ctx context.Context, msgID string, userID int64) error {
	update := bson.M{"$pull": bson.M{fieldReactions: bson.M{fieldUserID: userID}}}
	_, err := r.coll.UpdateOne(ctx, bson.M{fieldID: msgID}, update)
	return pkg.MapMongoError(err)
}

func (r *repository) Update(ctx context.Context, msg *chat.Message) error {
	update := bson.M{"$set": bson.M{
		fieldContent:   msg.Content,
		fieldIsEdited:  msg.IsEdited,
		fieldUpdatedAt: pkg.TimeNowUTC(),
	}}
	_, err := r.coll.UpdateOne(ctx, bson.M{fieldID: msg.ID}, update)
	return pkg.MapMongoError(err)
}

func (r *repository) SoftDelete(ctx context.Context, id string) error {
	update := bson.M{"$set": bson.M{
		fieldIsDeleted: true,
		fieldContent:   "",
		fieldReactions: []reactionDoc{},
		fieldUpdatedAt: pkg.TimeNowUTC(),
	}, "$unset": bson.M{
		fieldReplyToID: "",
	}}
	res, err := r.coll.UpdateOne(ctx, bson.M{fieldID: id}, update)
	if err != nil {
		return pkg.MapMongoError(err)
	}
	if res.MatchedCount == 0 {
		return pkg.ErrNotFound
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*chat.Message, error) {
	return r.getByID(ctx, bson.M{fieldID: id})
}

func (r *repository) GetByIDs(ctx context.Context, ids []string) ([]chat.Message, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	cursor, err := r.coll.Find(ctx, bson.M{fieldID: bson.M{"$in": ids}})
	if err != nil {
		return nil, pkg.MapMongoError(err)
	}
	defer cursor.Close(ctx)

	var docs []messageDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, pkg.MapMongoError(err)
	}
	msgs := make([]chat.Message, len(docs))
	for i, doc := range docs {
		msgs[i] = *doc.toDomain()
	}
	return msgs, nil
}

func (r *repository) GetByClientMsgID(ctx context.Context, clientMsgID string, senderID int64) (*chat.Message, error) {
	return r.getByID(ctx, bson.M{fieldClientMsgID: clientMsgID, fieldSenderID: senderID})
}

func (r *repository) GetList(ctx context.Context, params chat.GetMessagesParams) ([]chat.Message, error) {
	filter := bson.M{fieldConversationID: params.ConversationID}
	if params.Query.Cursor != "" {
		filter[fieldID] = bson.M{"$lt": params.Query.Cursor} // ULID
	}

	opts := options.Find().
		SetSort(bson.D{{Key: fieldID, Value: -1}}).
		SetLimit(int64(params.Query.GetFetchLimit()))

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, pkg.MapMongoError(err)
	}
	defer cursor.Close(ctx)

	var docs []messageDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, pkg.MapMongoError(err)
	}
	var res []chat.Message
	for _, doc := range docs {
		res = append(res, *doc.toDomain())
	}
	return res, nil
}

func (r *repository) getByID(ctx context.Context, filter any) (*chat.Message, error) {
	var doc messageDoc
	if err := r.coll.FindOne(ctx, filter).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, pkg.MapMongoError(err)
	}
	return doc.toDomain(), nil
}
