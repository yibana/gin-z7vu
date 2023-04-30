package routes

import (
	"encoding/json"
	"errors"
	"gin/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"
)

func GetCategoryRankCountGroupByPath(c *gin.Context) {
	result, err := db.AMZProductInstance.GetCategoryRankCountGroupByPath()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	bytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Data(200, "application/json", bytes)
}

func MongoFind(c *gin.Context) {
	var query bson.M
	var rsp mongoQueryResult
	var err error
	defer func() {
		if err != nil {
			rsp.Status = "error"
			rsp.Error = err.Error()
		} else {
			rsp.Status = "ok"
		}
		c.JSON(http.StatusOK, rsp)
	}()

	err = json.NewDecoder(c.Request.Body).Decode(&query)
	if err != nil {
		return
	}
	mongoQuery, err := db.AMZProductInstance.MongoFind(query)
	if err != nil {
		return
	}
	rsp.Result = mongoQuery
}

func MongoAggregate(c *gin.Context) {
	var query bson.M
	var rsp mongoAggregateResult
	var err error
	err = json.NewDecoder(c.Request.Body).Decode(&query)
	defer func() {
		if err != nil {
			rsp.Status = "error"
			rsp.Error = err.Error()
		} else {
			rsp.Status = "ok"
		}
		c.JSON(http.StatusOK, rsp)
	}()
	if err != nil {
		return
	}
	// 生成缓存key
	marshal, _ := json.Marshal(query)
	redis_key := db.NewBaseRedisKey(time.Minute*10, string(marshal))
	// 从缓存中获取
	if bytes, ok := db.RedisCacheInstance.GetAPICache(redis_key); ok {
		err = json.Unmarshal(bytes, &rsp.Result)
		return
	}
	var result []bson.M
	result, err = db.AMZProductInstance.MongoAggregate(query)
	if err != nil {
		return
	}
	if len(result) == 0 {
		err = errors.New("result is empty")
		return
	}
	var bytes []byte
	bytes, err = json.MarshalIndent(result, "", "  ")
	if err != nil {
		return
	}
	// 缓存结果
	db.RedisCacheInstance.SetAPICache(bytes, redis_key)
	rsp.Result = result
}
