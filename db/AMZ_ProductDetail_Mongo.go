package db

import (
	"context"
	"gin/amazon"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type AMZ_ProductDetail_Mongo struct {
	*MongoManger
}

func NewAMZ_ProductDetail_Mongo() (*AMZ_ProductDetail_Mongo, error) {
	manger, err := NewMongoManger("ProductDetail")
	return &AMZ_ProductDetail_Mongo{MongoManger: manger}, err
}

func (m *AMZ_ProductDetail_Mongo) SaveProductDetail(product *amazon.Product) error {
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
