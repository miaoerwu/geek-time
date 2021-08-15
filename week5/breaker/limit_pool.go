package breaker

import "sync"

type limitPoolManager struct {
	max     int
	tickets chan *struct{}
	lock    *sync.RWMutex
}

//NewLimitPoolManager 方法返回一个限流器
func NewLimitPoolManager(max int) *limitPoolManager {
	lpm := new(limitPoolManager)
	tickets := make(chan *struct{}, max)
	for i := 0; i < max; i++ {
		tickets <- &struct{}{}
	}
	lpm.max = max
	lpm.tickets = tickets
	lpm.lock = &sync.RWMutex{}
	return lpm
}

//ReturnAll 方法填充限流器所有令牌
func (limitPoolManager *limitPoolManager) ReturnAll() {
	limitPoolManager.lock.Lock()
	defer limitPoolManager.lock.Unlock()
	if len(limitPoolManager.tickets) == 0 {
		for i := 0; i < limitPoolManager.max; i++ {
			limitPoolManager.tickets <- &struct{}{}
		}
	}
}

//GetTicket 方法返回一个令牌，得到令牌返回true，令牌用完后返回false
func (limitPoolManager *limitPoolManager) GetTicket() bool {
	limitPoolManager.lock.RLock()
	defer limitPoolManager.lock.RUnlock()
	select {
	case <-limitPoolManager.tickets:
		return true
	default:
		return false
	}
}

//GetRemainder 方法返回剩余令牌数
func (limitPoolManager *limitPoolManager) GetRemainder() int {
	limitPoolManager.lock.RLock()
	defer limitPoolManager.lock.RUnlock()
	return len(limitPoolManager.tickets)
}
