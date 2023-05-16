package task

import (
	"context"
	"encoding/json"
	"fmt"
	"gin/amazon"
	"gin/config"
	"gin/db"
	"gin/scrape"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"time"
)

type ProductDetailTask struct {
	*BaseTask
	Asins      []string
	TaskPaths  []string
	runingPath string
}

func NewProductDetailTask(host string) *ProductDetailTask {
	return &ProductDetailTask{BaseTask: NewBaseTask(host)}
}

const redisKey_ProductDetailTask_done_paths = "ProductDetailTask:done:paths:v2"
const redisKey_CategoryTree_checked = "RedisSet:CategoryTree:checked"

func (t *ProductDetailTask) GetStatus() interface{} {
	return struct {
		RuningPath string `json:"runing_path"`
		PathsCount int    `json:"paths_count"`
		taskStatus
	}{
		RuningPath: t.runingPath,
		PathsCount: len(t.TaskPaths),
		taskStatus: t.BaseTask.GetStatus().(taskStatus),
	}
}

func (t *ProductDetailTask) Start(proxys []string, RandomDelay int) error {
	if t.Status != 0 {
		return fmt.Errorf("task is running")
	}
	err := t.ResetPaths()
	if err != nil {
		return err
	}
	return t.BaseTask.Start(proxys, RandomDelay, t.Run)
}

func (t *ProductDetailTask) Run(i int) {
	threadinfo := t.threadInfos[i]
	var GetAsinFailCount int
	var robotCount int
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			switch t.Status {
			case 1:
				asin, err := t.GetAsin(threadinfo.Proxy)
				if err != nil {
					GetAsinFailCount++
					if GetAsinFailCount > 3 {
						threadinfo.LastErr = "获取asin失败次数过多"
						threadinfo.LastErrTime = time.Now().Unix()
						return
					}
					threadinfo.LastErr = err.Error()
					time.Sleep(time.Second * 5)
					continue
				}
				GetAsinFailCount = 0
				if len(asin) == 0 {
					threadinfo.Info = "没有asin了"
					return
				}
				threadinfo.Info = fmt.Sprintf("正在处理asin:%s", asin)
				product, err := scrape.GetAmzProductEx(t.Host, asin, threadinfo.Proxy)
				if err != nil {
					if !strings.Contains(err.Error(), "Not Found") {
						//t.AddAsin(asin)
					} else {
						db.AMZProductInstance.DeleteAsin(asin)
					}

					threadinfo.Fail++
					t.AddFailCount()
					threadinfo.LastErr = fmt.Sprintf("%s:%s", asin, err.Error())
					threadinfo.LastErrTime = time.Now().Unix()
					if strings.Contains(err.Error(), "robot") || strings.Contains(err.Error(), "Service Unavailable") {
						robotCount++
						fmt.Println("robot || Service Unavailable", robotCount*60)
						for i := 0; i < robotCount*60; i++ {
							time.Sleep(time.Second)
							if t.Status != 1 {
								break
							}
						}
					} else {
						time.Sleep(time.Second * 10)
					}
					continue
				}
				if strings.ToLower(asin) != strings.ToLower(product.ASIN) { // 需要更新asin
					now := time.Now()
					err := db.AMZProductInstance.UpdateAsin(asin, product.ASIN)
					if err != nil {
						threadinfo.LastErr = fmt.Sprintf("更新asin失败:%s", err.Error())
						threadinfo.LastErrTime = time.Now().Unix()
						return
					}
					fmt.Println("更新asin成功", asin, product.ASIN, time.Since(now).Seconds())
				}
				threadinfo.Succ++
				robotCount = 0
				if len(threadinfo.LastErr) > 0 && time.Now().Unix()-threadinfo.LastErrTime > 60 {
					threadinfo.LastErr = ""
				}
				t.AddSuccessCount()
				// 保存到数据库
				db.AMZProductDetailInstance.SaveProductDetail(product)
				if len(product.Brand) > 0 {
					db.AMZBrandInstance.UpBrand(config.APIClientInstance, product.Brand)
				}
				t.SleepRandomDelay()
			default:
				threadinfo.Info = fmt.Sprintf("任务已停止：%d", t.Status)
				return
			}
		}
	}
}

func (t *ProductDetailTask) ResetDonePaths() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	_, err := db.RedisCacheInstance.Redis_client.Del(context.Background(), redisKey_ProductDetailTask_done_paths).Result()
	if err != nil {
		return err
	}
	return nil
}

func (t *ProductDetailTask) ResetPaths() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	var err error
	var checked string
	checked, err = db.RedisCacheInstance.TextGet(redisKey_CategoryTree_checked)
	if err != nil {
		return err
	}
	var checkedArr []string
	err = json.Unmarshal([]byte(checked), &checkedArr)
	if err != nil {
		return err
	}
	var paths []string
	for _, v := range checkedArr {
		// 取文本[Endxxx]sss 中的sss
		s := strings.Split(v, "]")
		if len(s) < 2 {
			continue
		}
		paths = append(paths, strings.Split(v, "]")[1])
	}

	// 获取已经完成的paths
	var donePaths []string
	donePaths, err = db.RedisCacheInstance.TextListGet(redisKey_ProductDetailTask_done_paths)
	if err != nil {
		return err
	}

	// 获取checkedArr-donePaths的差集
	t.TaskPaths = StringSliceDifference(paths, donePaths)
	if len(t.TaskPaths) == 0 {
		return fmt.Errorf("没有path了")
	}
	return nil
}

func (t *ProductDetailTask) AddAsin(asin string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.Asins = append(t.Asins, asin)
}

func (t *ProductDetailTask) GetAsin(Proxy string) (asin string, err error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if len(t.Asins) > 0 {
		//pop asin
		asin = t.Asins[0]
		t.Asins = t.Asins[1:]
		return
	}
	var asins []string
	for len(t.TaskPaths) > 0 {
		t.runingPath = t.TaskPaths[0]
		err = UpdateAsinList(t.runingPath, Proxy) // 实时更新asin列表
		if err != nil {
			return "", fmt.Errorf("更新asin列表失败:%s", err.Error())
		}
		t.TaskPaths = t.TaskPaths[1:]
		asins, err = GetTaskAsin(t.runingPath)
		if err != nil {
			return "", err
		}
		if len(asins) > 0 {
			break
		}
		// 记录已经完成的path
		err = db.RedisCacheInstance.TextListPush(redisKey_ProductDetailTask_done_paths, t.runingPath)
		if err != nil {
			return "", err
		}
	}
	if len(asins) == 0 {
		return "", fmt.Errorf("没有asin了")
	}

	asin = asins[0]
	t.Asins = asins[1:]
	return
}

func UpdateAsinList(path, proxy string) error {
	key := fmt.Sprintf("CategoryPathFlag:v2.1:%s", path)
	exist, err2 := db.RedisCacheInstance.Exist(key)
	if err2 != nil {
		return err2
	}
	if exist {
		return nil
	}

	var url string
	for _, categoryPath := range CategoryPaths {
		if path == categoryPath.Path {
			url = categoryPath.Url
			break
		}
	}
	if len(url) == 0 {
		return fmt.Errorf("没有找到url")
	}
	// 更新asin列表
	var list []amazon.CategoryRank
	var err error
	list, err = scrape.GetAmzProductListV2(url, proxy)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		// 保存到数据库
		var slist = make([]*amazon.CategoryRank, 0, len(list))
		for i, _ := range list {
			list[i].Path = path
			slist = append(slist, &list[i])
		}
		err = db.AMZProductInstance.BatchSaveCategoryRank(slist)
		if err != nil {
			return err
		}
		// 标记已经更新过
		err = db.RedisCacheInstance.Redis_client.Set(context.Background(), key, 1, time.Hour*24).Err()
		if err != nil {
			return err
		}

	}
	return nil
}

func GetTaskAsin(path string) ([]string, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"path": bson.M{"$in": []string{path}}}},
		{"$lookup": bson.M{"from": "ProductDetail", "localField": "id", "foreignField": "asin", "as": "ProductDetail"}},
		{"$match": bson.M{"$or": bson.A{
			bson.M{"ProductDetail.lasttime": bson.M{"$exists": false}},
			bson.M{"ProductDetail.lasttime": bson.M{"$lt": int(time.Now().Add(time.Hour*24).UnixMilli() / 1000)}},
		}}},
		{"$group": bson.M{"_id": "$id", "count": bson.M{"$sum": 1}}},
		{"$project": bson.M{"_id": 1, "count": 1}},
		{"$limit": 1000},
	}
	query := bson.M{
		"pipeline": pipeline,
	}
	aggregate, err := db.AMZProductInstance.MongoAggregate(query)
	if err != nil {
		return nil, err
	}
	var asins []string
	for _, v := range aggregate {
		asins = append(asins, v["_id"].(string))
	}
	return asins, nil
}

func StringSliceDifference(slice1 []string, slice2 []string) []string {
	map1 := make(map[string]bool)
	for _, v := range slice1 {
		map1[v] = true
	}

	for _, v := range slice2 {
		delete(map1, v)
	}

	result := make([]string, 0, len(map1))
	for k := range map1 {
		result = append(result, k)
	}

	return result
}
