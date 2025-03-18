package httpmodule

import (
	"github.com/soryetong/greasyx/console"
)

// 回调函数, 严格按照注册的顺序执行
type CallbackMap struct {
	funcMap  map[string]func()
	nameList []string
}

// 服务停机时的回调函数, 严格按照Append的的顺序执行, 先进先出, 同名会被覆盖
func NewStopCallbackMap() *CallbackMap {
	return &CallbackMap{
		funcMap:  make(map[string]func()),
		nameList: make([]string, 0),
	}
}

func (self *CallbackMap) Append(funcName string, value func()) {
	if _, exists := self.funcMap[funcName]; !exists {
		self.nameList = append(self.nameList, funcName)
	}
	self.funcMap[funcName] = value
}

func (self *CallbackMap) Foreach() {
	if self == nil || len(self.nameList) == 0 {
		return
	}
	for _, funcName := range self.nameList {
		console.Echo.Infof("ℹ️ 提示: 即将执行退出的回调函数: %s\n", funcName)
		self.funcMap[funcName]()
	}
}
