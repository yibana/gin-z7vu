package routes

import (
	"encoding/json"
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

// MongoAggregate
func MongoAggregate(c *gin.Context) {
	var query []bson.M
	err := json.NewDecoder(c.Request.Body).Decode(&query)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	// 生成缓存key
	marshal, _ := json.Marshal(query)
	redis_key := db.NewBaseRedisKey(time.Minute*10, string(marshal))
	// 从缓存中获取
	if bytes, ok := db.RedisCacheInstance.GetAPICache(redis_key); ok {
		c.Data(200, "application/json", bytes)
		return
	}

	result, err := db.AMZProductInstance.MongoAggregate(query)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if len(result) == 0 {
		c.Data(200, "application/json", []byte("[]"))
		return
	}

	bytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	// 缓存结果
	db.RedisCacheInstance.SetAPICache(bytes, redis_key)
	c.Data(200, "application/json", bytes)
}
