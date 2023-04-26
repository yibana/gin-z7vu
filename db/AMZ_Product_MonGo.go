package db

import (
	"context"
	"gin/amazon"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type AMZ_Product_MonGo struct {
	*MongoManger
}

func NewAMZ_Product_MonGo() (*AMZ_Product_MonGo, error) {
	manger, err := NewMongoManger("Products")
	return &AMZ_Product_MonGo{MongoManger: manger}, err
}

func (m *AMZ_Product_MonGo) SaveProduct(product *amazon.Product) error {
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

func (m *AMZ_Product_MonGo) SaveCategoryRank(CategoryRank *amazon.CategoryRank) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 如果path+rank已经存在，则更新，否则插入
	filter := map[string]interface{}{"path": CategoryRank.Path, "rank": CategoryRank.Rank}
	update := map[string]interface{}{"$set": CategoryRank}
	_, err := m.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func (m *AMZ_Product_MonGo) BatchSaveCategoryRank(CategoryRank []*amazon.CategoryRank) error {
	for _, rank := range CategoryRank {
		err := m.SaveCategoryRank(rank)
		if err != nil {
			return err
		}
	}
	return nil
}
