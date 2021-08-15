package breaker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	//OpenError 熔断器已经打开
	OpenError = errors.New("breaker_open")

	//OpenToHalfError 熔断器熔断超时需要状态流转
	OpenToHalfError = errors.New("OPEN_TO_HALF")
)

var (
	//FuncNilError 方法为空
	FuncNilError = errors.New("runFunc nil")

	//NameNilError 策略名为空
	NameNilError = errors.New("name nil")
)

//熔断器
type breaker struct {
	//策略名
	name string

	//熔断休眠时间
	sleepWindow int64

	//周期管理
	cycleTime int64

	//统计器
	counter *SlidingWindow

	//熔断时令牌桶
	lpm *limitPoolManager
}

//熔断器管理器
type breakerManager struct {
	//读写锁
	mutex *sync.RWMutex

	//Breaker集合
	manager map[string]*breaker
}

//定义全局熔断器管理器
var bm *breakerManager

//构造全局熔断器管理器
func init() {
	bm = new(breakerManager)
	bm.manager = make(map[string]*breaker)
	bm.mutex = &sync.RWMutex{}
}

//执行函数
type runFunc func() error

//回调函数
type fallbackFunc func(error)

//方法创建一个熔断器
func newBreaker(b *breakSettingInfo) *breaker {
	lpm := NewLimitPoolManager(b.BreakerTestMax)
	counter := NewSlidingWindow(SlidingWindowSetting{CycleTime: b.Interval,
		ErrorPercent:         b.ErrorPercentThreshold,
		HalfOpenErrorPercent: b.BreakerErrorPercentThreshold,
		RecoverNum:           b.BreakerTestMax,
	})
	return &breaker{
		name:        b.Name,
		cycleTime:   time.Now().Local().Unix() + b.SleepWindow,
		sleepWindow: b.SleepWindow,
		counter:     counter,
		lpm:         lpm,
	}
}

//方法从管理器获得熔断器
func getBreakerManager(name string) (*breaker, error) {
	if name == "" {
		return nil, errors.New("no name")
	}
	bm.mutex.RLock()
	breaker, ok := bm.manager[name]
	if !ok {
		bm.mutex.RUnlock()
		bm.mutex.Lock()
		defer bm.mutex.Unlock()
		if breaker, ok := bm.manager[name]; ok {
			return breaker, nil
		}
		breakInfo, err := NewBreakSettingInfo().SetName(name).AddBreakSetting()
		if err != nil {
			return nil, err
		}
		bm.manager[name] = breakInfo
		return breakInfo, nil
	} else {
		defer bm.mutex.RUnlock()
		return breaker, nil
	}
}

//方法失败处理
func (broker *breaker) fail() {
	state := broker.counter.GetStatus()
	switch state {
	case StatusClosed:
		atomic.StoreInt64(&broker.cycleTime, time.Now().Local().Unix()+broker.sleepWindow)
		broker.counter.AddForClose(false)
	case StatusOpen:
		if time.Now().Local().Unix() > atomic.LoadInt64(&broker.cycleTime) {
			if broker.counter.AddForOpen(false) {
				defer broker.lpm.ReturnAll()
				atomic.StoreInt64(&broker.cycleTime, time.Now().Local().Unix()+broker.sleepWindow)
			}
		}
	}
}

//方法成功处理
func (broker *breaker) success() {
	state := broker.counter.GetStatus()
	switch state {
	case StatusClosed:
		atomic.StoreInt64(&broker.cycleTime, time.Now().Local().Unix()+broker.sleepWindow)
		broker.counter.AddForClose(true)
	case StatusOpen:
		if time.Now().Local().Unix() > atomic.LoadInt64(&broker.cycleTime) {
			if broker.counter.AddForOpen(true) {
				defer broker.lpm.ReturnAll()
				atomic.StoreInt64(&broker.cycleTime, time.Now().Local().Unix()+broker.sleepWindow)
			}
		}
	}
}

//包装外部回调函数
func (broker *breaker) safeCallback(fallback fallbackFunc, err error) {
	if fallback == nil {
		return
	}
	fallback(err)
}

//执行方法前的处理
func (broker *breaker) beforeDo(ctx context.Context, name string) error {
	switch broker.counter.GetStatus() {
	case StatusOpen:
		if broker.cycleTime < time.Now().Local().Unix() {
			return OpenToHalfError
		}
		return OpenError
	}
	return nil
}

//执行方法后的处理
func (broker *breaker) afterDo(ctx context.Context, run runFunc, fallback fallbackFunc, err error) error {
	switch err {
	//熔断时
	case OpenError:
		broker.safeCallback(fallback, OpenError)
		return nil
	//熔断转移到半开启
	case OpenToHalfError:
		/*取令牌*/
		if !broker.lpm.GetTicket() {
			broker.safeCallback(fallback, OpenError)
			return nil
		}
		//执行方法
		runErr := run()
		if runErr != nil {
			broker.fail()
			broker.safeCallback(fallback, runErr)
			return runErr
		}
		broker.success()
		return nil
	default:
		if err != nil {
			broker.fail()
			broker.safeCallback(fallback, err)
			return err
		}
		broker.success()
		return nil
	}
}

//Do 方法结合熔断策略执行run函数
//其中参数包括:上下文ctx,策略名name,将要执行方法run,以及回调函数fallback.其中ctx,name,run必传
//run函数的错误会直接同步返回，回调函数fallback接收除了run错误以外还会接收熔断时错误，调用方如果需要降级可在fallback中自己判断
func Do(ctx context.Context, name string, run runFunc, fallback fallbackFunc) error {
	if run == nil {
		return FuncNilError
	}
	if name == "" {
		return NameNilError
	}
	//获得熔断器
	breaker, err := getBreakerManager(name)
	if err != nil {
		fallback(err)
		return err
	}
	//判断当前是否可以请求
	beforeDoErr := breaker.beforeDo(ctx, name)
	if beforeDoErr != nil {
		//如果有错误直接交给afterDo处理
		callBackErr := breaker.afterDo(ctx, run, fallback, beforeDoErr)
		return callBackErr
	}
	runErr := run()
	//执行后的处理
	return breaker.afterDo(ctx, run, fallback, runErr)
}
