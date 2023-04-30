package task

import (
	"context"
	"encoding/json"
	"fmt"
	"gin/db"
	"gin/scrape"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type threadInfo struct {
	Proxy   string `json:"proxy"`
	Succ    int    `json:"succ"`
	Fail    int    `json:"fail"`
	Info    string `json:"info"`
	LastErr string `json:"last_err"`
}

type ProductDetailTask struct {
	ctx         context.Context
	cancel      context.CancelFunc
	Status      int // 0:未开始 1:运行中 2:已停止
	lastErr     string
	lock        sync.Mutex
	threadInfos []*threadInfo
	Host        string
	Asins       []string
	TaskPaths   []string
	runingPath  string
	succCount   int64
	failCount   int64
	RandomDelay int
}

func NewProductDetailTask(host string) *ProductDetailTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &ProductDetailTask{ctx: ctx, cancel: cancel, Host: host}
}

func (t *ProductDetailTask) Stop() {
	if t.Status != 1 {
		return
	}
	t.Status = 2
	t.lastErr = "手动停止"
	t.cancel()
}

func (t *ProductDetailTask) GetStatus() interface{} {
	return struct {
		RuningPath  string        `json:"runing_path"`
		PathsCount  int           `json:"paths_count"`
		Status      string        `json:"status"`
		LastErr     string        `json:"last_err"`
		SuccCount   int64         `json:"succ_count"`
		FailCount   int64         `json:"fail_count"`
		ThreadInfos []*threadInfo `json:"thread_infos"`
	}{
		RuningPath:  t.runingPath,
		PathsCount:  len(t.TaskPaths),
		Status:      t.GetStatusString(),
		LastErr:     t.lastErr,
		ThreadInfos: t.threadInfos,
		SuccCount:   atomic.LoadInt64(&t.succCount),
		FailCount:   atomic.LoadInt64(&t.failCount),
	}
}

func (t *ProductDetailTask) SleepRandomDelay() {
	if t.RandomDelay > 0 {
		// min-max
		min := 1000
		max := t.RandomDelay
		if max < min {
			max = min
		}
		randNum := rand.Intn(max-min) + min
		time.Sleep(time.Duration(randNum) * time.Millisecond)
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
	t.RandomDelay = RandomDelay
	t.Status = 1
	t.lastErr = ""
	t.threadInfos = make([]*threadInfo, 0)
	t.threadInfos = append(t.threadInfos, &threadInfo{Proxy: ""})
	for _, proxy := range proxys {
		if len(proxy) > 0 {
			t.threadInfos = append(t.threadInfos, &threadInfo{Proxy: proxy})
		}
	}
	t.ctx, t.cancel = context.WithCancel(context.Background())
	go func() {
		defer func() {
			t.Status = 0
		}()
		synctask := sync.WaitGroup{}
		for i, _ := range t.threadInfos {
			synctask.Add(1)
			go func(i int) {
				defer func() {
					synctask.Done()
				}()
				t.Run(i)
			}(i)
		}
		synctask.Wait()
		t.lastErr = "任务完成"
	}()
	return nil
}

func (t *ProductDetailTask) Run(i int) {
	threadinfo := t.threadInfos[i]
	var GetAsinFailCount int
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			switch t.Status {
			case 1:
				asin, err := t.GetAsin()
				if err != nil {
					GetAsinFailCount++
					if GetAsinFailCount > 3 {
						threadinfo.LastErr = "获取asin失败次数过多"
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
				product, err := scrape.GetAmzProduct(t.Host, asin, threadinfo.Proxy)
				if err != nil {
					t.AddAsin(asin)
					threadinfo.Fail++
					atomic.AddInt64(&t.failCount, 1)
					threadinfo.LastErr = err.Error()
					time.Sleep(time.Second * 10)
					continue
				}
				threadinfo.Succ++
				atomic.AddInt64(&t.succCount, 1)
				// 保存到数据库
				db.AMZProductDetailInstance.SaveProductDetail(product)
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
	redisKey := "ProductDetailTask:done:paths"
	_, err := db.RedisCacheInstance.Redis_client.Del(t.ctx, redisKey).Result()
	if err != nil {
		return err
	}
	return nil
}

func (t *ProductDetailTask) ResetPaths() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	var err error
	redisKey := "RedisSet:CategoryTree:checked"
	var checked string
	checked, err = db.RedisCacheInstance.TextGet(redisKey)
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
	redisKey = "ProductDetailTask:done:paths"
	var donePaths []string
	donePaths, err = db.RedisCacheInstance.TextListGet(redisKey)
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

func (t *ProductDetailTask) GetAsin() (asin string, err error) {
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
		t.TaskPaths = t.TaskPaths[1:]
		asins, err = GettaskAsin(t.runingPath)
		if err != nil {
			return "", err
		}
		if len(asins) > 0 {
			break
		}
		// 记录已经完成的path
		redisKey := "ProductDetailTask:done:paths"
		err = db.RedisCacheInstance.TextListPush(redisKey, t.runingPath)
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

func (t *ProductDetailTask) GetStatusString() string {
	switch t.Status {
	case 0:
		return "未开始"
	case 1:
		return "运行中"
	case 2:
		return "已停止"
	default:
		return "未知"
	}
}

func GettaskAsin(path string) ([]string, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"path": bson.M{"$in": []string{path}}}},
		{"$lookup": bson.M{"from": "ProductDetail", "localField": "id", "foreignField": "asin", "as": "ProductDetail"}},
		{"$match": bson.M{"ProductDetail.asin": bson.M{"$exists": false}}},
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
