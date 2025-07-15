package ginahelper

import (
	"log"
	"runtime/debug"
)

func SafeGo(fn func()) {
	go RunSafe(fn)
}

func RunSafe(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[go func] panic: %v, funcName: %s，stack=%s \n",
				err, GetCallerName(fn), string(debug.Stack()))
		}
	}()

	fn()
}
