package backfill

import (
	"fmt"
	"gin/config"
	"gin/db"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func BackFill_up_brand() error {
	json_data := `{
    "collection":"ProductDetail",
    "time": 12345,
    "pipeline":[
    {
        "$match": {
            "brand": {
                "$ne": ""
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
            "$lookup": {
                "from": "Brands",
                "localField": "_id",
                "foreignField": "brand",
                "as": "brandInfo"
            }
        },
        {
            "$match": {
                "brandInfo.brand": {
                    "$exists": false
                }
            }
        }
]
}`
	var query bson.M
	err := bson.UnmarshalExtJSON([]byte(json_data), true, &query)
	if err != nil {
		return err
	}
	mongoQuery, err := db.AMZProductInstance.MongoAggregate(query)
	if err != nil {
		return err
	}
	count := len(mongoQuery)
	for i, v := range mongoQuery {
		brand := v["_id"].(string)
		if len(brand) == 0 {
			continue
		}
		fmt.Printf("当前:%d 剩余%d [%d]更新[%s]\n", i, count-i, v["count"], brand)
		_, err := db.AMZBrandInstance.UpBrand(config.APIClientInstance, brand)
		if err != nil {
			fmt.Println(err)
			continue
		}

	}
	return nil
}

func AutoBackFill() {
	go func() {
		for {
			err := BackFill_up_brand()
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Minute * 10)
		}
	}()
}
