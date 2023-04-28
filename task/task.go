package task

import (
	"context"
	"fmt"
	"gin/amazon"
	"gin/db"
	"gin/scrape"
	"log"
	"math/rand"
	"sync"
	"time"
)

type CategoryTask struct {
	ctx                      context.Context
	cancel                   context.CancelFunc
	CategoryPathPointer      int
	lock_CategoryPathPointer sync.Mutex
	Status                   int // 0:未开始 1:运行中 2:已停止
	lastErr                  error
}

func NewTask() *CategoryTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &CategoryTask{ctx: ctx, cancel: cancel}
}

func (t *CategoryTask) GetStatus() string {
	count, err := db.AMZProductInstance.GetCategoryRankCount()
	if err != nil {
		log.Println(err.Error())
	}

	return fmt.Sprintf("CategoryPathPointer: %d Status:%d CategoryRankCount:%d lastErr:%v", t.CategoryPathPointer, t.Status, count, t.lastErr)
}

func (t *CategoryTask) GetCategory() *CategoryPath {
	t.lock_CategoryPathPointer.Lock()
	defer t.lock_CategoryPathPointer.Unlock()
	if len(CategoryPaths) <= t.CategoryPathPointer {
		return nil
	}
	s := CategoryPaths[t.CategoryPathPointer]
	t.CategoryPathPointer++
	return &s
}

func (t *CategoryTask) Check_CategoryPathFlag(CategoryPath string) (bool, error) {
	key := fmt.Sprintf("CategoryPathFlag:%s", CategoryPath)
	return db.RedisCacheInstance.Exist(key)
}

func (t *CategoryTask) Save_CategoryPathFlag(CategoryPath string) error {
	key := fmt.Sprintf("CategoryPathFlag:%s", CategoryPath)
	return db.RedisCacheInstance.Redis_client.Set(t.ctx, key, 1, -1).Err()
}

func (t *CategoryTask) Run() {
	for {
		switch t.Status {
		case 0:
			time.Sleep(time.Second * 10)
		case 1:
			c := t.GetCategory()
			if c == nil {
				t.Status = 2
				return
			}
			flag, err := t.Check_CategoryPathFlag(c.Path)
			if err != nil {
				t.lastErr = err
				fmt.Println(err.Error())
				t.Status = 2
				return
			}
			if flag {
				continue
			}

			var failCount int
			var list []amazon.CategoryRank
			for t.Status == 1 && failCount < 5 {
				// 一直循环直到成功
				proxy := ""
				if failCount > 0 {
					proxy = t.RandProxy()
				}
				list, err = scrape.GetAmzProductList(c.Url, proxy)
				if err == nil {
					break
				}
				failCount++
				t.lastErr = err
				fmt.Println(err.Error())
				time.Sleep(time.Second * 1)
			}

			if failCount >= 5 {
				t.lastErr = fmt.Errorf("failCount >= 5")
				t.Status = 2
				return
			}

			// 这里是收尾工作
			if len(list) > 0 {
				// 保存到数据库
				var slist = make([]*amazon.CategoryRank, 0, len(list))
				for i, _ := range list {
					list[i].Path = c.Path
					slist = append(slist, &list[i])
				}
				err = db.AMZProductInstance.BatchSaveCategoryRank(slist)
				if err != nil {
					t.lastErr = err
					fmt.Println(err.Error())
					t.Status = 2
					return
				}
				err = t.Save_CategoryPathFlag(c.Path)
				if err != nil {
					t.lastErr = err
					fmt.Println(err.Error())
					t.Status = 2
					return
				}
			}

		default:
			return
		}
	}
}

func (t *CategoryTask) Stop() {
	if t.Status != 1 {
		return
	}
	t.Status = 2
	t.cancel()
}

func (t *CategoryTask) RandProxy() string {
	randint := rand.Intn(10000) + 1
	return fmt.Sprintf("http://qiyuewin_dc_%d:Qq9410176.@gwsg.sky-ip.net:1000", randint)
}

func (t *CategoryTask) Start(p int, n int) {
	if t.Status != 0 {
		return
	}
	if n <= 0 {
		n = 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.ctx = ctx
	t.cancel = cancel
	if p > 0 {
		t.CategoryPathPointer = p
	}
	t.lastErr = nil
	t.Status = 1
	go func() {
		defer func() {
			t.Status = 0
			if t.CategoryPathPointer-n > 0 {
				db.RedisCacheInstance.SetCategoryPathPointer(t.CategoryPathPointer - n)
			}
		}()
		synctask := sync.WaitGroup{}
		for i := 0; i < n; i++ {
			synctask.Add(1)
			go func() {
				defer synctask.Done()
				t.Run()
			}()
		}
		synctask.Wait()
	}()
}
