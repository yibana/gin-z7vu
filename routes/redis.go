package routes

import (
	"encoding/json"
	"fmt"
	"gin/db"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

func RedisSet(c *gin.Context) {
	var result redisResult
	var req redisReq
	var err error
	defer func() {
		if err != nil {
			result.Error = err.Error()
			result.Status = "error"
		} else {
			result.Status = "ok"
		}
		c.JSON(200, result)
	}()
	err = json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		return
	}
	key := fmt.Sprintf("RedisSet:%s", req.Key)
	value := req.Value
	exp_int := req.Exp
	err = db.RedisCacheInstance.TextSet(key, value, time.Duration(exp_int))
	if err != nil {
		return
	}
	result.Key = key
	result.Value = value
}

func RedisGet(c *gin.Context) {
	var result redisResult
	var err error
	defer func() {
		if err != nil {
			result.Error = err.Error()
			result.Status = "error"
		} else {
			result.Status = "ok"
		}
		c.JSON(200, result)
	}()
	key := c.DefaultQuery("key", "test")
	key = fmt.Sprintf("RedisSet:%s", key)
	value, err := db.RedisCacheInstance.TextGet(key)
	if err != nil {
		return
	}
	result.Key = key
	result.Value = value
}

func TaskPaths(c *gin.Context) {
	var result taskPathsResult
	redisKey := "RedisSet:CategoryTree:checked"
	var err error
	defer func() {
		if err != nil {
			result.Error = err.Error()
			result.Status = "error"
		} else {
			result.Status = "ok"
		}
		c.JSON(200, result)
	}()
	var checked string
	checked, err = db.RedisCacheInstance.TextGet(redisKey)
	if err != nil {
		return
	}
	var checkedArr []string
	err = json.Unmarshal([]byte(checked), &checkedArr)
	if err != nil {
		return
	}
	var paths []string
	for _, v := range checkedArr {
		// 取文本[Endxxx]sss 中的sss
		s := strings.Split(v, "]")
		if len(s) < 2 {
			continue
		}
		paths = append(paths, strings.Split(v, "]")[1])
	}
	result.Result = paths

}

func RedisCase(c *gin.Context) {
	var result RedisCaseResult
	var req RedisCaseReq
	var err error
	defer func() {
		if err != nil {
			result.Error = err.Error()
			result.Status = "error"
		} else {
			result.Status = "ok"
		}
		c.JSON(200, result)
	}()
	err = json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		return
	}
	req.Key = fmt.Sprintf("RedisCase:%s", req.Key)
	result.Key = req.Key
	result.Result = req.Value
	switch req.Case {
	case "TextListPush":
		err = db.RedisCacheInstance.TextListPush(req.Key, req.Value.(string))
	case "TextListPop":
		result.Result, err = db.RedisCacheInstance.TextListPop(req.Key)
	case "TextListGet":
		result.Result, err = db.RedisCacheInstance.TextListGet(req.Key)
	case "TextListInit":
		var value []string
		for _, i := range req.Value.([]interface{}) {
			switch i.(type) {
			case string:
				value = append(value, i.(string))
			case int:
				value = append(value, fmt.Sprintf("%d", i.(int)))
			}
		}
		db.RedisCacheInstance.TextListInit(req.Key, value)
	case "TextListExist":
		result.Result, err = db.RedisCacheInstance.TextListExist(req.Key, req.Value.(string))
	}

}
