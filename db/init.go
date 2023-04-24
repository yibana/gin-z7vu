package db

var MongoMangerInstance *MongoManger

func init() {
	var err error
	MongoMangerInstance, err = NewMongoManger()
	if err != nil {
		panic(err)
	}
}
