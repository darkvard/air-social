package conversation

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
				{Key: fieldParticipantUserID_Find, Value: 1},
				{Key: fieldUpdatedAt, Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: fieldParticipantUserID_Find, Value: 1},
				{Key: fieldType, Value: 1},
			},
		},
		{
			Keys:    bson.D{{Key: fieldClientConvID, Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	return err
}
