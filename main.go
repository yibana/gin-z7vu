package main

import (
	"gin/routes"
	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello world!",
		})
	})
	r.GET("/readme", routes.Readme)
	r.GET("/categorys", routes.AllCategorys)
	r.GET("/product", routes.GetProduct)
	r.GET("/paths", routes.Paths)
	r.GET("/product/list", routes.GetProductList)
	r.GET("/task", routes.Task)
	r.Run()
}
