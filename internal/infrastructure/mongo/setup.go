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

	// HeartbeatInterval and ServerSelectionTimeout are tuned below default (10s/30s)
	// to avoid multi-second latency spikes after idle: Docker silently drops TCP
	// connections, causing the driver to mark the server Unknown and block the next
	// request until the monitor confirms it is reachable again.
	clientOptions := options.Client().
		ApplyURI(mc.URI).
		SetMaxPoolSize(mc.MaxPoolSize).
		SetMinPoolSize(mc.MinPoolSize).
		SetMaxConnIdleTime(mc.MaxConnIdleTime).
		SetHeartbeatInterval(mc.HeartbeatInterval).
		SetServerSelectionTimeout(mc.ServerSelectionTimeout)

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
