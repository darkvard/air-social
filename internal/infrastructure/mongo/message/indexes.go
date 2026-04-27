package message

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func SetupIndexes(ctx context.Context, coll *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: fieldConversationID, Value: 1},
				{Key: fieldID, Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: fieldSenderID, Value: 1},
				{Key: fieldClientMsgID, Value: 1},
			},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(
				bson.M{fieldClientMsgID: bson.M{"$exists": true}},
			),
		},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	return err
}
