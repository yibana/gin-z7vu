package config

import (
	"gin/utils"
)

var MongodbUrl = utils.GetEnv("MONGO_URL", "mongodb://mongo:v2YTBkr8w6IwJpDjMrRc@containers-us-west-146.railway.app:6098")
var MongodbName = utils.GetEnv("MONGO_NAME", "mydb")
var MongodbCollection = utils.GetEnv("MONGO_COLLECTION", "categories")

var ProxyUrl = utils.GetEnv("HTTPS_PROXY", "")
