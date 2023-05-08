package routes

import (
	"encoding/json"
	"fmt"
	"gin/db"
	"gin/task"
	"github.com/gin-gonic/gin"
	"strconv"
)

func Taskcategorys(c *gin.Context) {
	cmd := c.DefaultQuery("cmd", "status")
	pointer, _ := db.RedisCacheInstance.GetCategoryPathPointer()
	p := c.DefaultQuery("p", fmt.Sprintf("%d", pointer))
	n := c.DefaultQuery("n", "1")
	var result = []byte(`{"status":"ok"}`)
	switch cmd {
	case "status":
		result = []byte(task.CategoryTaskInstance.GetStatus())
	case "start":
		sp, _ := strconv.Atoi(p)
		sn, _ := strconv.Atoi(n)
		task.CategoryTaskInstance.Start(sp, sn)
		result = []byte(task.CategoryTaskInstance.GetStatus())
	case "stop":
		task.CategoryTaskInstance.Stop()
		result = []byte(task.CategoryTaskInstance.GetStatus())
	case "RandProxy":
		result = []byte(task.CategoryTaskInstance.RandProxy())
	}

	c.Data(200, "application/json", result)
}

func TaskProducts(c *gin.Context) {
	var ProductDetailReq taskProductDetailReq
	var err error
	var result taskProductDetailResult
	defer func() {
		if err != nil {
			result.Error = err.Error()
			result.Status = "error"
		} else {
			result.Status = "ok"
		}
		c.JSON(200, result)
	}()
	err = json.NewDecoder(c.Request.Body).Decode(&ProductDetailReq)
	if err != nil {
		return
	}
	cmd := ProductDetailReq.Cmd
	switch cmd {
	case "start":
		err = task.ProductTaskInstance.Start(ProductDetailReq.Proxys, ProductDetailReq.RandomDelay, nil)

	}
}

func ProductDetail(c *gin.Context) {
	var ProductDetailReq taskProductDetailReq
	var err error
	var result taskProductDetailResult
	defer func() {
		if err != nil {
			result.Error = err.Error()
			result.Status = "error"
		} else {
			result.Status = "ok"
		}
		c.JSON(200, result)
	}()
	err = json.NewDecoder(c.Request.Body).Decode(&ProductDetailReq)
	if err != nil {
		return
	}
	cmd := ProductDetailReq.Cmd
	switch cmd {
	case "start":
		err = task.ProductDetailTaskInstance.Start(ProductDetailReq.Proxys, ProductDetailReq.RandomDelay)
	case "ResetPaths":
		err = task.ProductDetailTaskInstance.ResetPaths()
		result.Result = nil
		return
	case "ResetDonePaths":
		err = task.ProductDetailTaskInstance.ResetDonePaths()

	case "status":
		//result.Result = task.ProductDetailTaskInstance.GetStatus()
	case "stop":
		task.ProductDetailTaskInstance.Stop()
	}
	result.Result = task.ProductDetailTaskInstance.GetStatus()
}
