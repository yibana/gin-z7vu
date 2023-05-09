package db

import (
	"context"
	"encoding/json"
	"gin/APIClient"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type AMZ_Brand_MonGo struct {
	*MongoManger
}

type Brand struct {
	Data     interface{} `json:"data"`
	LastTime int64       `json:"lasttime"`
	Brand    string      `json:"brand"`
}

func NewAMZ_Brand_MonGo() (*AMZ_Brand_MonGo, error) {
	manger, err := NewMongoManger("Brands")
	return &AMZ_Brand_MonGo{MongoManger: manger}, err
}

func (m *AMZ_Brand_MonGo) SaveBrand(brand *Brand) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	brand.LastTime = time.Now().Unix()
	filter := map[string]interface{}{"brand": brand.Brand}
	update := map[string]interface{}{"$set": brand}
	_, err := m.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

// brand查询,带过期时间参数,获取不超过过期时间的数据
func (m *AMZ_Brand_MonGo) GetBrand(brand string, expire int64) (*Brand, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := map[string]interface{}{"brand": brand, "lasttime": map[string]interface{}{"$gte": expire}}
	var result Brand
	err := m.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type searchBrand struct {
	DomIntlFilter  string      `json:"domIntlFilter"`
	Searchfield1   string      `json:"searchfield1"`
	Textfield1     string      `json:"textfield1"`
	Display        string      `json:"display"`
	MaxReturn      string      `json:"maxReturn"`
	Nicetextfield1 interface{} `json:"nicetextfield1"`
	Cipotextfield1 interface{} `json:"cipotextfield1"`
}

func (m *AMZ_Brand_MonGo) UpBrand(apiclient *APIClient.APIHttpClient, brand string) (*Brand, error) {
	// 先判断是否存在
	brandInfo, err := m.GetBrand(brand, time.Now().Unix()-3600*24*10) // 10天过期
	if err == nil {
		return brandInfo, nil
	}

	var sb = searchBrand{
		DomIntlFilter:  "1",
		Searchfield1:   "all",
		Textfield1:     brand,
		Display:        "list",
		MaxReturn:      "5",
		Nicetextfield1: nil,
		Cipotextfield1: nil,
	}
	data, err := json.Marshal(sb)
	if err != nil {
		return nil, err
	}

	rsp, err := apiclient.Post("https://ised-isde.canada.ca/cipo/trademark-search/srch", data)
	if err != nil {
		return nil, err
	}
	var result Brand
	err = json.Unmarshal(rsp, &result.Data)
	if err != nil {
		return nil, err
	}
	result.Brand = brand
	err = m.SaveBrand(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
