package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func ConnectMongo() *mongo.Database {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://admin:password@127.0.0.1:27017"
	}

	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "air_social_chat"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("mongo connect failed: %v", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("mongo ping failed: %v", err)
	}

	return client.Database(dbName)
}
