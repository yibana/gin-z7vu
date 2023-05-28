package config

import (
	"gin/APIClient"
	"gin/utils"
)

var MongodbUrl = utils.GetEnv("MONGO_URL", "")
var MongodbName = utils.GetEnv("MONGO_NAME", "mydb")

var RedisUrl = utils.GetEnv("REDIS_URL", "redis://localhost:6379")

var ProxyUrl = utils.GetEnv("HTTPS_PROXY", "")

var APIClientInstance = APIClient.NewAPIClient(ProxyUrl, map[string]string{
	"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"Content-Type":     "application/json",
	"X-Requested-With": "XMLHttpRequest",
})
