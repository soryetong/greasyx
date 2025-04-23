package cachemodule

import (
	"time"

	"github.com/soryetong/greasyx/helper"
)

// 启动后台协程定期清理所有过期项
func (c *Cache) startCleaner() {
	if c.cleanerRunning.Swap(true) {
		return
	}
	c.cleanerStop = make(chan struct{})

	helper.SafeGo(func() {
		ticker := time.NewTicker(c.cleanInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.cleanExpired()
			case <-c.cleanerStop:
				return
			}
		}
	})
}

// 清理所有分片中过期的键
func (c *Cache) cleanExpired() {
	now := time.Now()
	for _, s := range c.shards {
		s.mu.Lock()
		for key, node := range s.items {
			if !node.expiresAt.IsZero() && now.After(node.expiresAt) {
				s.removeNode(node)
				delete(s.items, key)
				s.count.Add(-1)
			}
		}
		s.mu.Unlock()
	}
}
