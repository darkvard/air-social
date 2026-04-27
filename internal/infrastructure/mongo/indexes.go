package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"air-social/internal/infrastructure/mongo/conversation"
	"air-social/internal/infrastructure/mongo/message"
)

func SetupIndexes(ctx context.Context, db *mongo.Database) error {
	if err := conversation.SetupIndexes(ctx, db.Collection("conversations")); err != nil {
		return err
	}
	return message.SetupIndexes(ctx, db.Collection("messages"))
}
