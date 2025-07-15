package ginasrv

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/ginahelper"
	"github.com/soryetong/greasyx/libs/ginaauth"
	"github.com/soryetong/greasyx/modules/cachemodule"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const (
	LimitRuleKeyTypeIp          = "ip"
	LimitRuleKeyTypeUserid      = "userid"
	LimitRuleKeyTypeUrl         = "url"
	LimitRuleKeyTypeIpUserid    = "ip+userid"
	LimitRuleKeyTypeIpUrl       = "ip+url"
	LimitRuleKeyTypeUseridUrl   = "userid+url"
	LimitRuleKeyTypeIpUseridUrl = "ip+userid+url"
)

const (
	LimitRuleModeUri  = "uri"  // 按接口配置限流
	LimitRuleModeComm = "comm" // 通用限流
)

type LimitRule struct {
	Route   string
	KeyType string
	Rate    rate.Limit // 令牌产生速率（每秒）
	Burst   int        // 令牌桶大小（突发流量）
}

type LimiterStore struct {
	rules    []LimitRule
	limiters *cachemodule.Cache // key(string) => *rate.Limiter
	mode     string
	mu       sync.RWMutex
	ttl      time.Duration
}

type config struct {
	Mode  string        `json:"mode"`
	Ttl   time.Duration `json:"ttl"`
	Rules []LimitRule   `json:"rules"`
}

// 初始化限流器
func NewLimiterStore(rules []LimitRule, mode string, ttl time.Duration) *LimiterStore {
	return &LimiterStore{
		rules:    rules,
		mode:     mode,
		limiters: cachemodule.New(10000, 64, 0),
		ttl:      ttl,
	}
}

func NewLimiterStoreFromFile(path string) *LimiterStore {
	cfg, err := LoadLimiterRulesFromFile(path)
	if err != nil {
		console.Echo.Fatalf("❌ 错误: 读取限流规则错误: %s", err)
	}

	return NewLimiterStore(cfg.Rules, cfg.Mode, cfg.Ttl)
}

// 生成组合 key
func buildKey(ctx *gin.Context, keyType, uri string) string {
	ip := ctx.ClientIP()
	userId := ginaauth.GetTokenData[int64](ctx, "id")
	parts := strings.Split(keyType, "+")
	var vals []string
	for _, p := range parts {
		switch p {
		case LimitRuleKeyTypeIp:
			vals = append(vals, fmt.Sprintf("ip:%s", ip))
		case LimitRuleKeyTypeUserid:
			vals = append(vals, fmt.Sprintf("uid:%d", userId))
		case LimitRuleKeyTypeUrl:
			vals = append(vals, fmt.Sprintf("url:%s", uri))
		case LimitRuleKeyTypeIpUserid:
			vals = append(vals, fmt.Sprintf("ipuid:%s|%d", ip, userId))
		case LimitRuleKeyTypeIpUrl:
			vals = append(vals, fmt.Sprintf("ipurl:%s|%s", ip, uri))
		case LimitRuleKeyTypeUseridUrl:
			vals = append(vals, fmt.Sprintf("uidurl:%d|%s", userId, uri))
		case LimitRuleKeyTypeIpUseridUrl:
			vals = append(vals, fmt.Sprintf("ipuidurl:%s|%d|%s", ip, userId, uri))
		}
	}
	return strings.Join(vals, "|")
}

// 获取限流器（不存在就创建）
func (self *LimiterStore) getLimiter(key string, rule LimitRule) *rate.Limiter {
	if val, found := self.limiters.Get(key); found {
		return val.(*rate.Limiter)
	}

	limiter := rate.NewLimiter(rule.Rate, rule.Burst)
	self.limiters.Set(key, limiter, self.ttl*time.Second)
	return limiter
}

// 限流判断
func (self *LimiterStore) Allow(ctx *gin.Context) bool {
	self.mu.RLock()
	defer self.mu.RUnlock()

	if len(self.rules) == 0 {
		return true
	}
	uri := ginahelper.ConvertToRestfulURL(strings.TrimPrefix(ctx.Request.URL.Path, viper.GetString("App.RouterPrefix")))
	if self.mode == LimitRuleModeComm {
		rule := self.rules[0]
		key := buildKey(ctx, rule.KeyType, uri)
		limiter := self.getLimiter(key, rule)
		return limiter.Allow()
	}

	// 按接口配置限流
	for _, rule := range self.rules {
		if "/"+strings.Trim(rule.Route, "/") == uri {
			key := uri + "|" + buildKey(ctx, rule.KeyType, uri)
			limiter := self.getLimiter(key, rule)
			return limiter.Allow()
		}
	}

	return true
}

// 支持后期更新规则
func (self *LimiterStore) UpdateRules(newRules []LimitRule, mode string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.rules = newRules
	self.mode = mode
}

func LoadLimiterRulesFromFile(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rules config
	err = json.Unmarshal(data, &rules)

	return &rules, err
}

func WatchLimiterRulesFile(path string, store *LimiterStore) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		gina.Log.Error("[Limiter.WatchRulesFile] watch error:", zap.Error(err))
		return
	}
	err = watcher.Add(path)
	if err != nil {
		gina.Log.Error("[Limiter.WatchRulesFile] watch add error:", zap.Error(err))
		return
	}

	ginahelper.SafeGo(func() {
		defer watcher.Close()
		for {
			select {
			case e := <-watcher.Events:
				if e.Op&fsnotify.Write != 0 {
					time.Sleep(100 * time.Millisecond) // 避免文件写入中间状态
					if conf, err := LoadLimiterRulesFromFile(path); err == nil {
						store.UpdateRules(conf.Rules, conf.Mode)
						console.Echo.Infof("✅ 提示: 限流规则热更新成功")
					} else {
						gina.Log.Error("[Limiter.WatchRulesFile] 规则热更新失败:", zap.Error(err))
					}
				}
			case err := <-watcher.Errors:
				gina.Log.Error("[Limiter.WatchRulesFile] watch error:", zap.Error(err))
			}
		}
	})
}
