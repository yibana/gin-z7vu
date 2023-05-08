package task

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type ScrapeTask interface {
	GetStatus() interface{}
	Stop()
	Start(proxys []string, RandomDelay int, Run func(int)) error
	SleepRandomDelay()
	AddSuccessCount()
	AddFailCount()
}

type taskStatus struct {
	StatusStr   string        `json:"status_str"`
	Status      int           `json:"status"`
	LastErr     string        `json:"last_err"`
	SuccCount   int64         `json:"succ_count"`
	FailCount   int64         `json:"fail_count"`
	ThreadInfos []*threadInfo `json:"thread_infos"`
}

type threadInfo struct {
	Proxy       string `json:"proxy"`
	Succ        int    `json:"succ"`
	Fail        int    `json:"fail"`
	Info        string `json:"info"`
	LastErr     string `json:"last_err"`
	LastErrTime int64  `json:"lastErrTime"`
}

type BaseTask struct {
	ctx         context.Context
	cancel      context.CancelFunc
	Status      int // 0:未开始 1:运行中 2:已停止
	lastErr     string
	lock        sync.Mutex
	threadInfos []*threadInfo
	Host        string
	succCount   int64
	failCount   int64
	RandomDelay int
	ScrapeTask
}

func NewBaseTask(host string) *BaseTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseTask{ctx: ctx, cancel: cancel, Host: host}
}

func (t *BaseTask) AddSuccessCount() {
	atomic.AddInt64(&t.succCount, 1)
}

func (t *BaseTask) AddFailCount() {
	atomic.AddInt64(&t.failCount, 1)
}

func (t *BaseTask) Stop() {
	if t.Status != 1 {
		return
	}
	t.Status = 2
	t.lastErr = "手动停止"
	t.cancel()
}

func (t *BaseTask) GetStatus() interface{} {
	return taskStatus{
		StatusStr:   t.GetStatusString(),
		Status:      t.Status,
		LastErr:     t.lastErr,
		ThreadInfos: t.threadInfos,
		SuccCount:   atomic.LoadInt64(&t.succCount),
		FailCount:   atomic.LoadInt64(&t.failCount),
	}
}

func (t *BaseTask) GetStatusString() string {
	switch t.Status {
	case 0:
		return "未开始"
	case 1:
		return "运行中"
	case 2:
		return "已停止"
	}
	return ""
}

func (t *BaseTask) Start(proxys []string, RandomDelay int, Run func(int)) error {
	if t.Status != 0 {
		return fmt.Errorf("task is running")
	}
	if Run == nil {
		return fmt.Errorf("Run is nil")
	}
	t.Status = 1
	t.lastErr = ""
	t.RandomDelay = RandomDelay
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
				Run(i)
			}(i)
		}
		synctask.Wait()
		t.lastErr = "任务完成"
	}()
	return nil
}

func (t *BaseTask) SleepRandomDelay() {
	if t.RandomDelay > 0 {
		// min-max
		min := 2000
		max := t.RandomDelay
		if max < min {
			max = min
		}
		randNum := rand.Intn(max-min) + min
		time.Sleep(time.Duration(randNum) * time.Millisecond)
	}
}
