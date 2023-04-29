package main

import (
	"gin/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

func main() {
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
	r.GET("/product", routes.GetProduct)
	r.GET("/paths", routes.Paths)
	r.GET("/product/list", routes.GetProductList)
	r.GET("/task", routes.Task)
	r.GET("/CategoryRankCountGroupByPath", routes.GetCategoryRankCountGroupByPath)
	r.POST("/MongoAggregate", routes.MongoAggregate)
	r.Run()
}
