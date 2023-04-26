package config

import (
	"gin/utils"
)

var MongodbUrl = utils.GetEnv("MONGO_URL", "mongodb://mongo:v2YTBkr8w6IwJpDjMrRc@containers-us-west-146.railway.app:6098")
var MongodbName = utils.GetEnv("MONGO_NAME", "mydb")

var RedisUrl = utils.GetEnv("REDIS_URL", "redis://default:3WPIki9dXShd6ZZhGXKZ@containers-us-west-65.railway.app:7937")

var ProxyUrl = utils.GetEnv("HTTPS_PROXY", "")
