package db

import "gin/config"

var AMZProductInstance *AMZ_Product_MonGo
var AMZProductDetailInstance *AMZ_ProductDetail_Mongo

var RedisCacheInstance *RedisCacheManger

func init() {
	var err error
	AMZProductInstance, err = NewAMZ_Product_MonGo()
	if err != nil {
		panic(err)
	}

	AMZProductDetailInstance, err = NewAMZ_ProductDetail_Mongo()
	if err != nil {
		panic(err)
	}

	RedisCacheInstance, err = NewRedisCacheManger(config.RedisUrl)
	if err != nil {
		panic(err)
	}
}
