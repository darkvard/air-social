package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func SetupIndexes(ctx context.Context, db *mongo.Database) error {
	if err := setupConversationIndexes(ctx, db); err != nil {
		return err
	}
	return setupMessageIndexes(ctx, db)
}

func setupConversationIndexes(ctx context.Context, db *mongo.Database) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "participants.user_id", Value: 1},
				{Key: "updated_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "participants.user_id", Value: 1},
				{Key: "type", Value: 1},
			},
		},
	}
	_, err := db.Collection("conversations").Indexes().CreateMany(ctx, indexes)
	return err
}

func setupMessageIndexes(ctx context.Context, db *mongo.Database) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "conversation_id", Value: 1},
				{Key: "_id", Value: -1}, 
			},
		},
		{
			Keys:    bson.D{{Key: "client_msg_id", Value: 1}},			// todo: mess_doc.client_msg_id has omitempty (duplicate omitempty key)
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
	}
	_, err := db.Collection("messages").Indexes().CreateMany(ctx, indexes)
	return err
}
