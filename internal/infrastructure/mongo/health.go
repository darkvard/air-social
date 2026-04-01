package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Health struct {
	client *mongo.Client
}

func NewHealth(client *mongo.Client) *Health {
	return &Health{
		client: client,
	}
}

func (h *Health) Ping(ctx context.Context) error {
	return h.client.Ping(ctx, readpref.Primary())
}
