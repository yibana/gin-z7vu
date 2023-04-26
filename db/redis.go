package db

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisCacheManger struct {
	Redis_client *redis.Client
}

func NewRedisCacheManger(url string) (*RedisCacheManger, error) {
	op, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return &RedisCacheManger{
		Redis_client: redis.NewClient(op),
	}, nil
}

func (rds *RedisCacheManger) Exist(key string) (bool, error) {
	result, err := rds.Redis_client.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (rds *RedisCacheManger) SetCategoryPathPointer(Pointer int) error {
	return rds.Redis_client.Set(context.Background(), "CategoryPathPointer", Pointer, -1).Err()
}

func (rds *RedisCacheManger) GetCategoryPathPointer() (int, error) {
	result, err := rds.Redis_client.Get(context.Background(), "CategoryPathPointer").Int()
	if err != nil {
		return 0, err
	}
	return result, nil
}
