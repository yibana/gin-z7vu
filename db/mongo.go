package db

import (
	"context"
	"gin/amazon"
	"gin/config"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)
import "go.mongodb.org/mongo-driver/mongo"

type MongoManger struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoManger() (*MongoManger, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(config.MongodbUrl))
	if err != nil {
		return nil, err
	}
	collection := client.Database(config.MongodbName).Collection(config.MongodbCollection)
	return &MongoManger{client: client, collection: collection}, nil
}

func (m *MongoManger) SaveProduct(product *amazon.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 如果asin已经存在，则更新，否则插入
	filter := map[string]interface{}{"asin": product.ASIN}
	update := map[string]interface{}{"$set": product}
	_, err := m.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}
