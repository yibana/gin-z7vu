package db

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zentures/cityhash"
	"log"
	"strings"
	"time"
)

func extractMaxN(str string, n int) string {
	if len(str) <= n {
		return str
	}
	return str[:n] + "..."
}

type RedisCacheManger struct {
	Redis_client *redis.Client
}
type BaseRedisKey struct {
	expiration   time.Duration
	key_hash_str string
	summary      string
}

func NewBaseRedisKey(expiration time.Duration, arg ...string) *BaseRedisKey {
	if len(arg) == 0 {
		return nil
	}
	if expiration.Seconds() < 1 {
		expiration = time.Second
	}
	key := strings.Join(arg, " ")
	key_hash_str := fmt.Sprintf("APICache:%x", HashData(key))
	return &BaseRedisKey{
		expiration:   expiration,
		key_hash_str: key_hash_str,
		summary:      extractMaxN(key, 120),
	}
}
func HashData(Raw string) uint64 {
	hash := cityhash.CityHash64([]byte(Raw), uint32(len(Raw)))
	return hash
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

func (rds *RedisCacheManger) GetAPICache(rk *BaseRedisKey) ([]byte, bool) {
	now := time.Now()
	bytes, err := rds.Redis_client.Get(context.Background(), rk.key_hash_str).Bytes()
	if err != nil {
		return nil, false
	}
	log.Printf("[GetAPICache]:[%s] RespSize:[%d] Elapsed_Time:%vs\n", rk.summary, len(bytes), time.Now().Sub(now).Seconds())
	return bytes, true
}

func (rds *RedisCacheManger) SetAPICache(data []byte, rk *BaseRedisKey) {
	if len(data) == 0 {
		return
	}
	err := rds.Redis_client.Set(context.Background(), rk.key_hash_str, data, rk.expiration).Err()
	if err != nil {
		log.Println(err)
	}
}
