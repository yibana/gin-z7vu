package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gin/amazon"
	"gin/config"
	"gin/db"
	"gin/scrape"
	"gin/task"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func Readme(c *gin.Context) {
	filePath := filepath.Join(".", "README.md")
	markdownContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error reading file: %s", err.Error()))
		return
	}
	// 渲染模板，并将markdownContent变量替换为动态的Markdown内容
	c.HTML(200, "markdown.html", gin.H{
		"MarkdownContent": string(markdownContent),
		"title":           "README.md",
	})
}

func get_category_paths() []string {
	filepath := filepath.Join(".", "category.json")
	categorys, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil
	}
	var category amazon.Category
	err = json.Unmarshal(categorys, &category)
	if err != nil {
		return nil
	}
	// 遍历category，获取所有末尾节点的path
	var paths []string
	var f func(path []string, categorys *amazon.Category)
	f = func(path []string, categorys *amazon.Category) {
		if len(categorys.Sub) == 0 {
			paths = append(paths, strings.Join(path, " > "))
		} else {
			for _, sub := range categorys.Sub {
				f(append(path, sub.Name), sub)
			}
		}
	}
	f([]string{category.Name}, &category)
	return paths
}

func Paths(c *gin.Context) {
	bytes, err := json.MarshalIndent(task.CategoryPaths, "", "  ")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error MarshalIndent: %s", err.Error()))
		return
	}
	c.Data(200, "application/json", bytes)
}

func AllCategorys(c *gin.Context) {
	filepath := filepath.Join(".", "category.json")
	categorys, err := ioutil.ReadFile(filepath)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error reading file: %s", err.Error()))
		return
	}
	c.Data(200, "application/json", categorys)
}

func Categorys2TreeNode(c *gin.Context) {
	filepath := filepath.Join(".", "category.json")
	categorys, err := ioutil.ReadFile(filepath)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error reading file: %s", err.Error()))
		return
	}
	var category amazon.Category
	err = json.Unmarshal(categorys, &category)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error Unmarshal: %s", err.Error()))
		return
	}
	treeNode := amazon.ConvertToTreeNode("", &category)
	bytes, err := json.MarshalIndent(treeNode, "", "  ")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error MarshalIndent: %s", err.Error()))
		return
	}
	c.Data(200, "application/json", bytes)
}

func GetProduct(c *gin.Context) {
	host := c.DefaultQuery("host", "www.amazon.ca")
	asin := c.DefaultQuery("asin", "B08MR2C1T7")
	proxy := c.DefaultQuery("proxy", config.ProxyUrl)

	product, err := scrape.GetAmzProduct(host, asin, proxy)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// Print the product details
	marshal, err := json.MarshalIndent(product, "", "  ")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error MarshalIndent: %s", err.Error()))
		return
	}
	// 将product保存到mongodb数据库
	db.AMZProductInstance.SaveProduct(product)

	c.Data(200, "application/json", marshal)
}

func GetProductList(c *gin.Context) {
	url := c.DefaultQuery("url", "https://www.amazon.ca/Best-Sellers-Amazon-Devices-Accessories-Amazon-Device-Accessories/zgbs/amazon-devices/2980422011/ref=zg_bs_unv_amazon-devices_2_5500205011_2")
	proxy := c.DefaultQuery("proxy", config.ProxyUrl)
	products, err := scrape.GetAmzProductList(url, proxy)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	// Print the product details
	marshal, err := json.MarshalIndent(products, "", "  ")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error MarshalIndent: %s", err.Error()))
		return
	}
	c.Data(200, "application/json", marshal)
}

func Task(c *gin.Context) {
	cmd := c.DefaultQuery("cmd", "status")
	pointer, _ := db.RedisCacheInstance.GetCategoryPathPointer()
	p := c.DefaultQuery("p", fmt.Sprintf("%d", pointer))
	n := c.DefaultQuery("n", "1")
	var result = []byte(`{"status":"ok"}`)
	switch cmd {
	case "status":
		result = []byte(task.TaskInstance.GetStatus())
	case "start":
		sp, _ := strconv.Atoi(p)
		sn, _ := strconv.Atoi(n)
		task.TaskInstance.Start(sp, sn)
		result = []byte(task.TaskInstance.GetStatus())
	case "stop":
		task.TaskInstance.Stop()
		result = []byte(task.TaskInstance.GetStatus())
	case "RandProxy":
		result = []byte(task.TaskInstance.RandProxy())
	}

	c.Data(200, "application/json", result)
}

func GetCategoryRankCountGroupByPath(c *gin.Context) {
	result, err := db.AMZProductInstance.GetCategoryRankCountGroupByPath()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	bytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Data(200, "application/json", bytes)
}

func MongoFind(c *gin.Context) {
	var query bson.M
	var rsp mongoQueryResult
	var err error
	defer func() {
		if err != nil {
			rsp.Status = "error"
			rsp.Error = err.Error()
		} else {
			rsp.Status = "ok"
		}
		c.JSON(http.StatusOK, rsp)
	}()

	err = json.NewDecoder(c.Request.Body).Decode(&query)
	if err != nil {
		return
	}
	mongoQuery, err := db.AMZProductInstance.MongoFind(query)
	if err != nil {
		return
	}
	rsp.Result = mongoQuery
}

func MongoAggregate(c *gin.Context) {
	var query bson.M
	var rsp mongoAggregateResult
	var err error
	err = json.NewDecoder(c.Request.Body).Decode(&query)
	defer func() {
		if err != nil {
			rsp.Status = "error"
			rsp.Error = err.Error()
		} else {
			rsp.Status = "ok"
		}
		c.JSON(http.StatusOK, rsp)
	}()
	if err != nil {
		return
	}
	// 生成缓存key
	marshal, _ := json.Marshal(query)
	redis_key := db.NewBaseRedisKey(time.Minute*10, string(marshal))
	// 从缓存中获取
	if bytes, ok := db.RedisCacheInstance.GetAPICache(redis_key); ok {
		err = json.Unmarshal(bytes, &rsp.Result)
		return
	}
	var result []bson.M
	result, err = db.AMZProductInstance.MongoAggregate(query)
	if err != nil {
		return
	}
	if len(result) == 0 {
		err = errors.New("result is empty")
		return
	}
	var bytes []byte
	bytes, err = json.MarshalIndent(result, "", "  ")
	if err != nil {
		return
	}
	// 缓存结果
	db.RedisCacheInstance.SetAPICache(bytes, redis_key)
	rsp.Result = result
}

func RedisSet(c *gin.Context) {
	var result redisResult
	var req redisReq
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		result.Error = err.Error()
		result.Status = "error"
		c.JSON(200, result)
	}
	key := fmt.Sprintf("RedisSet:%s", req.Key)
	value := req.Value
	exp_int := req.Exp
	err = db.RedisCacheInstance.Redis_client.Set(context.Background(), key, value, time.Duration(exp_int)).Err()
	if err != nil {
		result.Error = err.Error()
		result.Status = "error"
	} else {
		result.Status = "ok"
		result.Key = key
		result.Value = value
	}
	c.JSON(200, result)
}

func RedisGet(c *gin.Context) {
	var result redisResult
	key := c.DefaultQuery("key", "test")
	key = fmt.Sprintf("RedisSet:%s", key)
	value, err := db.RedisCacheInstance.Redis_client.Get(context.Background(), key).Result()
	if err != nil {
		result.Error = err.Error()
		result.Status = "error"
	} else {
		result.Status = "ok"
		result.Key = key
		result.Value = value
	}
	c.JSON(200, result)
}
