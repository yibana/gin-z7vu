package config

import (
	"gin/utils"
)

var MongodbUrl = "mongodb+srv://lovefy5201314:fMnhDfmuWnv4vUrL@amz.ce6lrve.mongodb.net/test"
var MongodbName = utils.GetEnv("MONGO_NAME", "mydb")

var RedisUrl = utils.GetEnv("REDIS_URL", "redis://default:3WPIki9dXShd6ZZhGXKZ@containers-us-west-65.railway.app:7937")

var ProxyUrl = utils.GetEnv("HTTPS_PROXY", "")
