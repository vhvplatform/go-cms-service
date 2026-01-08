package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient wraps MongoDB client with common functionality
type MongoClient struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// ConnectMongo establishes a connection to MongoDB
func ConnectMongo(ctx context.Context, uri, database string, timeout time.Duration) (*MongoClient, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, err
	}

	db := client.Database(database)

	return &MongoClient{
		Client:   client,
		Database: db,
	}, nil
}

// Close closes the MongoDB connection
func (mc *MongoClient) Close(ctx context.Context) error {
	if mc.Client != nil {
		return mc.Client.Disconnect(ctx)
	}
	return nil
}

// Ping verifies the connection is still alive
func (mc *MongoClient) Ping(ctx context.Context) error {
	return mc.Client.Ping(ctx, nil)
}

// Collection returns a handle to a collection
func (mc *MongoClient) Collection(name string) *mongo.Collection {
	return mc.Database.Collection(name)
}
