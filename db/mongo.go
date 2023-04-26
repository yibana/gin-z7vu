package db

import (
	"context"
	"gin/config"
	"go.mongodb.org/mongo-driver/mongo/options"
)
import "go.mongodb.org/mongo-driver/mongo"

type MongoManger struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoManger(CollectionName string) (*MongoManger, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(config.MongodbUrl))
	if err != nil {
		return nil, err
	}
	collection := client.Database(config.MongodbName).Collection(CollectionName)
	return &MongoManger{client: client, collection: collection}, nil
}
