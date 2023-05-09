package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"gin/config"
	"gin/db"
	"github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"
)

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

func DownloadQuery(c *gin.Context) {
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
	results, err := db.AMZProductInstance.MongoAggregate(query)
	if err != nil {
		return
	}
	rsp.Result = results
	// 将查询结果转换为CSV格式文档
	if len(results) > 0 {
		// 获取CSV头部信息
		header := make([]string, 0, len(results[0]))
		for key := range results[0] {
			header = append(header, key)
		}
		// 将查询结果转换为 Excel 文档
		file := xlsx.NewFile()
		var sheet *xlsx.Sheet
		sheet, err = file.AddSheet("Result")
		if err != nil {
			return
		}
		row := sheet.AddRow()
		for _, h := range header {
			row.AddCell().SetValue(h)
		}
		for _, result := range results {
			row = sheet.AddRow()
			for _, h := range header {
				row.AddCell().SetValue(result[h])
			}
		}
		//file.Save("result.xlsx")
		buff := new(bytes.Buffer)
		err = file.Write(buff)
		if err != nil {
			return
		}
		data := buff.Bytes()
		//ioutil.WriteFile("result2.xlsx", data, 0644)
		rsp.Result = data
		if err != nil {
			return
		}
	}
}

func QueryBrand(c *gin.Context) {
	brand := c.DefaultQuery("brand", "")
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

	if brand == "" {
		err = errors.New("brand is empty")
		return
	}

	mongoQuery, err := db.AMZBrandInstance.UpBrand(config.APIClientInstance, brand)
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
