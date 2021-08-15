package breaker

import (
	"errors"
)

var (
	DefaultInterval                     int64 = 60
	DefaultSleepWindow                  int64 = 65
	DefaultBreakerTestMax                     = 20
	DefaultErrorPercentThreshold              = 50
	DefaultBreakerErrorPercentThreshold       = 50
)

type breakSettingInfo struct {
	Name                         string
	Interval                     int64
	SleepWindow                  int64
	BreakerTestMax               int
	ErrorPercentThreshold        int
	BreakerErrorPercentThreshold int
}

//NewBreakSettingInfo 新建熔断器配置
func NewBreakSettingInfo() *breakSettingInfo {
	return &breakSettingInfo{}
}

//SetName 设置策略名
func (brokerSettingInfo *breakSettingInfo) SetName(name string) *breakSettingInfo {
	brokerSettingInfo.Name = name
	return brokerSettingInfo
}

//SetErrorPercentThreshold 设置方法错误比
func (brokerSettingInfo *breakSettingInfo) SetErrorPercentThreshold(errorPercentThreshold int) *breakSettingInfo {
	brokerSettingInfo.ErrorPercentThreshold = errorPercentThreshold
	return brokerSettingInfo
}

//SetSleepWindow 设置熔断休眠时间
func (brokerSettingInfo *breakSettingInfo) SetSleepWindow(sleepWindow int64) *breakSettingInfo {
	brokerSettingInfo.SleepWindow = sleepWindow
	return brokerSettingInfo
}

//SetInterval 设置采样周期
func (brokerSettingInfo *breakSettingInfo) SetInterval(interval int64) *breakSettingInfo {
	brokerSettingInfo.Interval = interval
	return brokerSettingInfo
}

//SetBreakerTestMax 设置熔断最大测试次数
func (brokerSettingInfo *breakSettingInfo) SetBreakerTestMax(breakerTestMax int) *breakSettingInfo {
	brokerSettingInfo.BreakerTestMax = breakerTestMax
	return brokerSettingInfo
}

//AddBreakSetting 添加配置方法，最后将期望配置添加到熔断器管理器里，如果策略名为空字符串则报错
func (brokerSettingInfo *breakSettingInfo) AddBreakSetting() (*breaker, error) {
	if brokerSettingInfo.Name == "" {
		return nil, errors.New("name nil")
	}
	if brokerSettingInfo.BreakerErrorPercentThreshold <= 0 {
		brokerSettingInfo.BreakerErrorPercentThreshold = DefaultBreakerErrorPercentThreshold
	}
	if brokerSettingInfo.ErrorPercentThreshold <= 0 {
		brokerSettingInfo.ErrorPercentThreshold = DefaultErrorPercentThreshold
	}
	if brokerSettingInfo.BreakerTestMax < DefaultBreakerTestMax {
		brokerSettingInfo.BreakerTestMax = DefaultBreakerTestMax
	}
	if brokerSettingInfo.Interval < DefaultInterval {
		brokerSettingInfo.Interval = DefaultInterval
	}
	if brokerSettingInfo.SleepWindow < DefaultInterval {
		brokerSettingInfo.SleepWindow = DefaultSleepWindow
	}
	if brokerSettingInfo.Interval%int64(gridNum) != 0 {
		return nil, errors.New("interval must 20`s multiple")
	}
	breaker := newBreaker(brokerSettingInfo)
	return breaker, nil
}
