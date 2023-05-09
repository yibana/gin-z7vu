package main

import (
	"context"
	"fmt"
	"gin/backfill"
	"gin/config"
	"gin/db"
	"gin/scrape"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	cmd := "upBrand"
	switch cmd {
	case "fixBrand":
		fix_brand()
	case "upBrand":
		backfill.BackFill_up_brand()
	case "test":
		db.AMZBrandInstance.UpBrand(config.APIClientInstance, "DoceMora")
	}
}

func fix_brand() {
	json_data := `{
    "collection":"ProductDetail",
    "pipeline":[
    {
        "$match": {
            "brand": {
                "$regex": "^Visit the "
            }
        }
    },
    {
        "$group": {
            "_id": "$brand",
            "count": {
                "$sum": 1
            }
        }
    },
    {
        "$sort": {
            "_id": -1
        }
    }
]
}`
	var query bson.M
	err := bson.UnmarshalExtJSON([]byte(json_data), true, &query)
	if err != nil {
		panic(err)
	}
	mongoQuery, err := db.AMZProductInstance.MongoAggregate(query)
	if err != nil {
		panic(err)
	}
	for _, v := range mongoQuery {
		newbrand := scrape.ExtractBrandName(v["_id"].(string))
		// 更新到数据库
		if len(newbrand) == 0 {
			continue
		}
		rsp, err := db.AMZProductDetailInstance.GetCollection().UpdateMany(context.Background(), bson.M{"brand": v["_id"]}, bson.M{"$set": bson.M{"brand": newbrand}})
		if err != nil {
			panic(err)
		}
		fmt.Printf("[%d]更新[%s]->[%s] 更新了%d条数据\n", v["count"], v["_id"], newbrand, rsp.ModifiedCount)
		db.AMZBrandInstance.UpBrand(config.APIClientInstance, newbrand)
	}
}
