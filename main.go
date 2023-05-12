package main

import (
	"gin/backfill"
	"gin/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

func main() {
	backfill.AutoBackFill()
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}

	r.Use(cors.New(config))
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello world!",
		})
	})
	r.GET("/readme", routes.Readme)
	r.GET("/categorys", routes.AllCategorys)
	r.GET("/Categorys2TreeNode", routes.Categorys2TreeNode)
	r.GET("/ConvertToTreeSelect", routes.ConvertToTreeSelect)
	r.GET("/product", routes.GetProduct)
	r.GET("/productV2", routes.GetProductV2)
	r.GET("/paths", routes.Paths)
	r.GET("/product/list", routes.GetProductList)
	r.GET("/task/categorys", routes.Taskcategorys)
	r.POST("/task/ProductDetail", routes.ProductDetail)
	r.GET("/CategoryRankCountGroupByPath", routes.GetCategoryRankCountGroupByPath)
	r.POST("/MongoAggregate", routes.MongoAggregate)
	r.POST("/MongoFind", routes.MongoFind)
	r.POST("/redis/set", routes.RedisSet)
	r.GET("/redis/get", routes.RedisGet)
	r.POST("/redis/case", routes.RedisCase)
	r.GET("/TaskPaths", routes.TaskPaths)
	r.GET("/query/brand", routes.QueryBrand)
	r.POST("/download/query", routes.DownloadQuery)
	r.Run()
}
