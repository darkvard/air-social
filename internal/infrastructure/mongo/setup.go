package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"air-social/internal/config"
)

func NewConnection(mc config.MongoConfig) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.ConnectTimeout)
	defer cancel()

	// The mongo-driver provides connection pooling out of the box.
	// We can configure the pool size, but the default is usually fine.
	clientOptions := options.Client().
		ApplyURI(mc.URI).
		SetMaxPoolSize(mc.MaxPoolSize).
		SetMinPoolSize(mc.MinPoolSize).
		SetMaxConnIdleTime(mc.MaxConnIdleTime)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the primary server to verify that the client can connect.
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}
