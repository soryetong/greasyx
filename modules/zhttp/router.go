package zhttp

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"os"
)

type ZRouterProvider interface {
	Register() *gin.Engine
	ExitCallback() *ExitCallbackMap
}

// 退出时的回调函数, 严格按照注册的顺序执行
type ExitCallbackMap struct {
	funcMap  map[string]func()
	nameList []string
}

func NewExitCallbackMap() *ExitCallbackMap {
	return &ExitCallbackMap{
		funcMap:  make(map[string]func()),
		nameList: make([]string, 0),
	}
}

func (self *ExitCallbackMap) Insert(funcName string, value func()) {
	if _, exists := self.funcMap[funcName]; !exists {
		self.nameList = append(self.nameList, funcName)
	}
	self.funcMap[funcName] = value
}

func (self *ExitCallbackMap) Iterate() {
	for _, funcName := range self.nameList {
		_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\033[34m [GREASYX-info] "+
			"即将执行退出的回调函数: %s \033[0m\n", funcName))
		self.funcMap[funcName]()
	}
}
