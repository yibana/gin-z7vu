package routes

import (
	"encoding/json"
	"fmt"
	"gin/amazon"
	"gin/config"
	"gin/db"
	"gin/scrape"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
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
func Paths(c *gin.Context) {
	filepath := filepath.Join(".", "category.json")
	categorys, err := ioutil.ReadFile(filepath)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error reading file: %s", err.Error()))
		return
	}
	var category amazon.Category
	err = json.Unmarshal(categorys, &category)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error reading file: %s", err.Error()))
		return
	}
	// 遍历category，获取所有末尾节点的path
	var paths []string
	var f func(path []string, categorys *amazon.Category)
	f = func(path []string, categorys *amazon.Category) {
		if len(categorys.Sub) == 0 {
			paths = append(paths, strings.Join(path, " > "))
		} else {
			for _, c2 := range categorys.Sub {
				f(append(path, c2.Name), c2)
			}
		}
	}
	f([]string{category.Name}, &category)

	c.String(http.StatusOK, strings.Join(paths, "\n"))
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
	db.MongoMangerInstance.SaveProduct(product)

	c.Data(200, "application/json", marshal)
}
