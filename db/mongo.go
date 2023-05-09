package db

import (
	"context"
	"gin/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
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

func (m *MongoManger) GetClient() *mongo.Client {
	return m.client
}

func (m *MongoManger) GetCollection() *mongo.Collection {
	return m.collection
}

func (m *MongoManger) MongoAggregate(query bson.M) ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := m.collection
	if CollectionName, ok := query["collection"]; ok {
		collection = m.client.Database(config.MongodbName).Collection(CollectionName.(string))
	}
	pipeline := query["pipeline"]
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var result []bson.M
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *MongoManger) MongoFind(query bson.M) ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var options options.FindOptions
	if limit, ok := query["limit"]; ok {
		options.SetLimit(int64(limit.(float64)))
	} else {
		options.SetLimit(10)
	}

	if page, ok := query["page"]; ok {
		options.SetSkip(int64((page.(float64))-1) * *options.Limit)
	}

	if sort, ok := query["sort"]; ok {
		options.SetSort(sort)
	}

	filter, ok := query["filter"]
	if !ok {
		filter = bson.M{}
	}
	collection := m.collection
	if CollectionName, ok := query["collection"]; ok {
		collection = m.client.Database(config.MongodbName).Collection(CollectionName.(string))
	}

	cursor, err := collection.Find(ctx, filter, &options)
	if err != nil {
		return nil, err
	}
	var result []bson.M
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
