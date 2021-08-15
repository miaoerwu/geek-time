package breaker

import (
	"sync"
	"sync/atomic"
	"time"
)

//Counter 滑动窗口中的计数器
type Counter struct {
	val uint32

	timeStamp int64
}

//SlidingWindow 滑动窗口
type SlidingWindow struct {
	//滑动窗口周期
	windowSize int64

	//全部请求的计数器
	allRequestCounter []*Counter

	//失败请求的计数器
	failRequestCounter []*Counter

	//熔断器关闭状态的计数通道
	closeCountChan chan bool

	//半开启请求数
	halfOpenReqNum int32

	//半开启错误数
	halfOpenFailReqNum int32

	//熔断恢复需要的请求个数
	recoverNum int32

	//开始时间
	startTime int64

	//格子时间
	gridTime int64

	//错误率
	errorPercent int

	//半开启错误率
	halfOpenErrorPercent int

	//状态
	status int32
}

const (
	StatusClosed int32 = iota

	StatusOpen
)

var gridNum = 20

//SlidingWindowSetting 滑动窗口设置
type SlidingWindowSetting struct {
	//周期
	CycleTime int64

	//错误率
	ErrorPercent int

	//半开启错误率
	HalfOpenErrorPercent int

	//熔断恢复需要的请求个数
	RecoverNum int
}

//消费资源
func (slidingWindow *SlidingWindow) consumeRes() {
	for {
		select {
		case res := <-slidingWindow.closeCountChan:
			if res {
				slidingWindow.add()
			} else {
				slidingWindow.addFail()
			}
		}
	}
}

//成功时的计数方法，因为加锁所以默认所有请求是有时序性的
func (slidingWindow *SlidingWindow) add() {
	addTime := time.Now().Local().Unix()
	diffTime := addTime - slidingWindow.startTime
	if diffTime >= slidingWindow.windowSize {
		slidingWindow.startTime = addTime
		diffTime = 0
	}
	index := diffTime / slidingWindow.gridTime

	if addTime-slidingWindow.allRequestCounter[index].timeStamp >= slidingWindow.windowSize {
		slidingWindow.allRequestCounter[index].val = 1
		slidingWindow.allRequestCounter[index].timeStamp = addTime
		return
	}
	slidingWindow.allRequestCounter[index].val++
}

//失败时的计数方法
func (slidingWindow *SlidingWindow) addFail() {
	addTime := time.Now().Local().Unix()
	diffTime := addTime - slidingWindow.startTime
	if diffTime >= slidingWindow.windowSize {
		slidingWindow.startTime = addTime
		diffTime = 0
	}
	loc := diffTime / slidingWindow.gridTime

	var wg sync.WaitGroup
	//请求计数
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if addTime-slidingWindow.allRequestCounter[loc].timeStamp >= slidingWindow.windowSize {
			slidingWindow.allRequestCounter[loc].val = 1
			slidingWindow.allRequestCounter[loc].timeStamp = addTime
			return
		}
		slidingWindow.allRequestCounter[loc].val++
	}(&wg)

	//失败计数
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if addTime-slidingWindow.failRequestCounter[loc].timeStamp >= slidingWindow.windowSize {
			slidingWindow.failRequestCounter[loc].val = 1
			slidingWindow.failRequestCounter[loc].timeStamp = addTime
			return
		}
		slidingWindow.failRequestCounter[loc].val++
	}(&wg)

	wg.Wait()

	percent, ok := slidingWindow.getFailPercentThreshold()
	if !ok {
		return
	}
	if percent >= slidingWindow.errorPercent {
		atomic.StoreInt32(&slidingWindow.status, StatusOpen)
	}
}

//计算错误率
func (slidingWindow *SlidingWindow) getFailPercentThreshold() (int, bool) {
	bucket := time.Now().Local().Unix() - slidingWindow.windowSize

	var totalCount uint32 = 0
	var totalFailedCount uint32 = 0
	totalCountChan := make(chan uint32)
	totalFailedCountChan := make(chan uint32)
	go func() {
		var totalClosedCount uint32 = 0
		for _, count := range slidingWindow.allRequestCounter {
			if count.timeStamp <= bucket {
				continue
			}
			totalClosedCount += count.val
		}
		totalCountChan <- totalClosedCount
	}()

	go func() {
		var totalFailedCount uint32 = 0
		for _, count := range slidingWindow.failRequestCounter {
			if count.timeStamp <= bucket {
				continue
			}
			totalFailedCount += count.val

		}
		totalFailedCountChan <- totalFailedCount
	}()

	frequency := 0
loop:
	for {
		select {
		case totalCount = <-totalCountChan:
			frequency++
			if frequency >= 2 {
				break loop
			}
		case totalFailedCount = <-totalFailedCountChan:
			frequency++
			if frequency >= 2 {
				break loop
			}
		}
	}
	if totalCount < 10 {
		return 0, false
	}
	return int(float32(totalFailedCount)/float32(totalCount)*100 + 0.5), true
}

//NewSlidingWindow 创建一个滑动窗口
func NewSlidingWindow(slidingWindowSetting SlidingWindowSetting) *SlidingWindow {
	var allRequestCounter []*Counter
	var failRequestCounter []*Counter
	for idx := 0; idx < gridNum; idx++ {
		counter := &Counter{}
		counter.val = 0
		counter.timeStamp = 0
		allRequestCounter = append(allRequestCounter, counter)
	}
	for idx := 0; idx < gridNum; idx++ {
		counter := &Counter{}
		counter.val = 0
		counter.timeStamp = 0
		failRequestCounter = append(failRequestCounter, counter)
	}
	slidingWindow := &SlidingWindow{
		windowSize:           slidingWindowSetting.CycleTime,
		gridTime:             slidingWindowSetting.CycleTime / int64(gridNum),
		allRequestCounter:    allRequestCounter,
		failRequestCounter:   failRequestCounter,
		startTime:            0,
		closeCountChan:       make(chan bool, 10000),
		errorPercent:         slidingWindowSetting.ErrorPercent,
		halfOpenErrorPercent: slidingWindowSetting.HalfOpenErrorPercent,
		recoverNum:           int32(slidingWindowSetting.RecoverNum),
	}
	go slidingWindow.consumeRes()
	return slidingWindow
}

//AddForClose 记录关闭状态的数量
func (slidingWindow *SlidingWindow) AddForClose(res bool) {
	slidingWindow.closeCountChan <- res
}

//AddForOpen 记录开启状态的数量
func (slidingWindow *SlidingWindow) AddForOpen(res bool) bool {
	if res {
		reqTotal := atomic.AddInt32(&slidingWindow.halfOpenReqNum, 1)
		if reqTotal >= slidingWindow.recoverNum {
			defer slidingWindow.clear()
			failTotal := atomic.LoadInt32(&slidingWindow.halfOpenFailReqNum)
			if int(float32(failTotal)/float32(reqTotal)*100+0.5) >= slidingWindow.halfOpenErrorPercent {
				atomic.StoreInt32(&slidingWindow.status, StatusOpen)
			} else {
				atomic.StoreInt32(&slidingWindow.status, StatusClosed)
			}
			return true
		}
		return false
	}
	reqTotal := atomic.AddInt32(&slidingWindow.halfOpenReqNum, 1)
	failTotal := atomic.AddInt32(&slidingWindow.halfOpenFailReqNum, 1)
	if reqTotal >= slidingWindow.recoverNum {
		defer slidingWindow.clear()
		if int(float32(failTotal)/float32(reqTotal)*100+0.5) >= slidingWindow.halfOpenErrorPercent {
			atomic.StoreInt32(&slidingWindow.status, StatusOpen)
		} else {
			atomic.StoreInt32(&slidingWindow.status, StatusClosed)
		}
		return true
	}
	return false
}

//清空半开数据
func (slidingWindow *SlidingWindow) clear() {
	atomic.StoreInt32(&slidingWindow.halfOpenReqNum, 0)
	atomic.StoreInt32(&slidingWindow.halfOpenFailReqNum, 0)
}

//GetStatus 获取状态
func (slidingWindow *SlidingWindow) GetStatus() int32 {
	return atomic.LoadInt32(&slidingWindow.status)
}
