package helper

import (
	"fmt"

	"go.uber.org/zap"
)

func SafeGo(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("helper GoSafe panic", zap.Any("", err))
			}
		}()
		fn()
	}()
}
