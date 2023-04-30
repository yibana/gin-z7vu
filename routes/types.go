package routes

import "go.mongodb.org/mongo-driver/bson"

type ApiResult struct {
	Error  string `json:"error"`
	Status string `json:"status"`
}

type redisResult struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	ApiResult
}

type redisReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Exp   int    `json:"exp"`
}

type mongoAggregateResult struct {
	ApiResult
	Result []bson.M `json:"result"`
}

type mongoQueryResult struct {
	ApiResult
	Result []bson.M `json:"result"`
}
