package redis

import (
	"github.com/go-redis/redis/v8"
	"strings"
	"time"
	"fmt"
	"context"
	"github.com/soryetong/greasyx/console"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"github.com/soryetong/greasyx/gina"
)

func init() {
	console.Append(redisCmd)
}

var redisCmd = &cobra.Command{
	Use:   "Redis",
	Short: "Init Redis",
	Long:  `加载Redis模块之后，可以通过 gina.Rdb 进行数据操作`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString("Redis.Addr")
		if addr == "" {
			_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-error] "+
				"你正在加载Redis模块，但是你未配置Redis.Addr，请先添加配置\n")
			os.Exit(124)
		}

		viper.SetDefault("Redis.IsCluster", false)
		viper.SetDefault("Redis.Db", 0)
		gina.Rdb = initRedis(
			addr,
			viper.GetString("Redis.Password"),
			viper.GetInt("Redis.Db"),
			viper.GetBool("Redis.IsCluster"),
		)
		_, _ = fmt.Fprintf(os.Stderr, "\n\033[32m [GREASYX-info] "+
			"Redis模块加载成功, 你可以使用 `gina.Rdb` 进行数据操作\033[0m\n")
	},
}

func initRedis(addr, password string, db int, cluster bool) redis.Cmdable {
	if cluster {
		return initCluster(addr, password, db)
	} else {
		return initSingleNode(addr, password, db)
	}
}

func initSingleNode(addr, password string, db int) redis.Cmdable {
	networkType := "tcp"
	if strings.Contains(addr, "/") {
		networkType = "unix"
	}

	rdbClient := redis.NewClient(&redis.Options{
		Network:  networkType,
		Addr:     addr,
		Password: password,
		DB:       db,

		// 超时
		DialTimeout:  5 * time.Second, // 连接建立超时时间，默认5秒
		ReadTimeout:  3 * time.Second, // 读超时，默认3秒， -1表示取消读超时
		WriteTimeout: 3 * time.Second, // 写超时，默认等于读超时
		PoolTimeout:  4 * time.Second, // 当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒

		// 命令执行失败时的重试策略
		MaxRetries:      0,                      // 命令执行失败时，最多重试多少次，默认为0即不重试
		MinRetryBackoff: 8 * time.Millisecond,   // 每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
		MaxRetryBackoff: 512 * time.Millisecond, // 每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔
	})

	_, err := rdbClient.Ping(context.Background()).Result()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\n[GREASYX-error] "+
			"Redis连接失败: %s\n", err))
		os.Exit(124)
	}

	return rdbClient
}

func initCluster(addr, password string, db int) redis.Cmdable {
	rdbClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    strings.Split(addr, ","),
		Password: password,

		// 超时
		DialTimeout:  5 * time.Second, // 连接建立超时时间，默认5秒
		ReadTimeout:  3 * time.Second, // 读超时，默认3秒， -1表示取消读超时
		WriteTimeout: 3 * time.Second, // 写超时，默认等于读超时
		PoolTimeout:  4 * time.Second, // 当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒

		// 命令执行失败时的重试策略
		MaxRetries:      10,                     // 命令执行失败时，最多重试多少次，默认为0即不重试
		MinRetryBackoff: 8 * time.Millisecond,   // 每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
		MaxRetryBackoff: 512 * time.Millisecond, // 每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔

		// 支持TLS接入
		// TLSConfig: &tls.Config{
		// 	InsecureSkipVerify: true,
		// },

		// 默认false，即只能在主节点上进行读写操作，如果为true则允许在从节点上执行只含读操作的命令
		ReadOnly: true,
		// 默认false，置为true则ReadOnly自动为true，表示在处理只读命令时，可以在一个slot对应的主节点和所有从节点中选取ping()的响应时长最短的一个节点来读数据
		RouteRandomly: true,
		// 默认false，置为true则ReadOnly自动为true，表示在处理只读命令时，可以在一个slot对应的主节点和所有从节点中随机选取一个节点来读数据
		RouteByLatency: true,
	})

	_, err := rdbClient.Ping(context.Background()).Result()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\n[GREASYX-error] "+
			"Redis集群连接失败: %s\n", err))
		os.Exit(124)
	}

	return rdbClient
}
