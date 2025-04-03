<h1 align="center">greasyx</h1>

<p align="center"> 这是一个为了快速构建一个go项目而生的Go+Gin的Web项目脚手架</p>

<p align="center">千人千面，它可能不适合你，但完全满足我的需求，欢迎技术讨论，但不喜勿喷</p>

## 介绍

- 它只是我为了快速构建自己的项目而创建的，让我不必在每新建一个新项目都复制这些代码
- 它只是把日常工作中常用的工具和第三方包进行了整合，没有什么高深的技术，也不会有任何的性能影响
- 在项目中, 把常用的MySQL、HTTP、Redis、MongoDB等以模块化的方式按需加载

## 当前go版本

- go 1.24

## 使用

它只需要简单的几步就可以快速创建一个项目

1. 初始化你的项目文件夹

```bash
go mod init your_project_name

cd your_project_name
```
2. 拉取 `greasyx`

```bash
go get -u github.com/soryetong/greasyx
```

3. 创建 `main.go`

```go
package main

import (
	"github.com/soryetong/greasyx/gina"
)

func main() {
	gina.Run()
}
```

4. 创建你的api文件

```bash
mkdir api_desc

touch api_desc/test1.go
```

```api
# test1.go
# 这个最终会生成对应的结构体，内容和Gin的保持一致
# 比如GET请求就需要`form`, 需要参数校验就使用 `binding`
type TestReq {
    Page int64 `json:"page" form:"page"`
    PageSize int64 `json:"pageSize" form:"pageSize"`
}

type TestResp {
    Name string `json:"name"`
}

# 必须以 service 打头，会生成对应的接口、方法
# SystemTest 可以理解成一个模块
# Use Record 表示这个模块需要使用到的中间件, 多个中间件用英文逗号隔开
service SystemTest Use Record {
    # get 表示请求类型，支持get、post、put、delete, returns前表示请求参数, returns后表示返回参数
    get test1 (TestReq) returns (TestResp)
    post test2 returns
}
```

📢注意：这一步不是必须的，只是有了api文件，就可以自动生成`Struct`、`Router`、`Handler`、`Logic`

5. 自动生成代码

```bash
# src表示api文件路径，output表示生成的代码路径
go run main.go autoc src=./api_desc output=./internal
```
6. 加载Server和需要用到模块

```go
package main

import (
	_ "your_project_name/internal/server"

	"github.com/soryetong/greasyx/gina"
	_ "github.com/soryetong/greasyx/modules/casbinmodule"
	_ "github.com/soryetong/greasyx/modules/mysqlmodule"
)

func main() {
	gina.Run()
}
```

解释：

- `_ "your_project_name/internal/server"` 表示加载你的服务，这个是必须的, 除此之外, 其他的都是按需加载的

详见[内置模块](#已内置的模块)

7. 运行项目

## 配置文件

```json
{
  "App": {
    "Name": "app",
    "Env": "local",
    "Addr": ":18002",
    "Timeout": 1
  },
  "MySQL": {
    "Dsn": "root:123456@tcp(127.0.0.1:3307)/greasyx-admin?charset=utf8&parseTime=True&loc=Local&timeout=5s",
    "remark": "以下配置是可选的，如果没有配置，则使用默认配置",
    "LogLevel": 3,
    "EnableLogWriter": false,
    "MaxIdleConn": 10,
    "MaxConn": 200,
    "SlowThreshold": 2
  },
  "Redis": {
    "Addr": "127.0.0.1:6379",
    "Password": "",
    "Db": 0,
    "IsCluster": false
  },
  "Mongo": {
    "Url": "mongodb://admin:123123@192.168.0.13:27017/?minPoolSize=5&maxPoolSize=35&maxIdleTimeMS=30000"
  },
  "Log": {
    "remark": "日志的所有配置都是可选的，都有默认配置，可以先看一下下面关于配置的解释",
    "Path": "./logs/",
    "Logrotate": false,
    "Mode": "both",
    "Recover": true,
    "MaxSize": 1,
    "MaxBackups": 3,
    "MaxAge": 1,
    "Compress": true
  },
  "Casbin": {
    "ModePath": ""
  }
}
```

- `App`：表示项目配置，包括项目名、环境、端口、超时时间等


- `MySQL`：表示MySQL配置，包括DSN(必要的)、日志级别、最大空闲连接数、最大连接数、慢查询阈值等
    
  - `EnableLogWriter: true` 会使用zap日志记录MySQL日志


- `Redis`：表示Redis配置，包括地址、密码、数据库、是否集群等


- `Mongo`：表示MongoDB配置，包括地址、用户名、密码、数据库等


- `Log`：表示日志配置(非必要,都有默认值)，包括日志路径、模式、是否开启Recover、最大文件大小、最大备份数、最大保存天数、是否压缩等

  - `Logrotate` 是否开启日志轮转，默认是开启的。由于日志是按照日期分目录的，这个值为 `true` 时，会在每天晚上 24 点按照日期重新创建日志目录

    - 如果你使用了 `Linux` 自带的 `logrotate` ，那么建议 `Logrotate` 设置为 `false`

  - `Mode` 支持: `file`写入文件，`both`写入文件和控制台，`console`写入控制台，`close`不写入任何地方
  
  - `Recover` 在你项目启动后删除已经生成的日志文件，将不会自动创建文件并继续写入，如果这个设置为 `true` 则会检查并重新创建文件，但有一定的性能影响


- `Casbin`：表示Casbin配置，包括模式路径等, 没有该路径时会使用 `greasyx` 提供的默认配置

## 提供的模块

- HTTP

      需要直接导入 `httpmodule`

示例：

```go
package server

import (
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/modules/httpmodule"
	"github.com/spf13/viper"
)

// 注册服务, 然后在main.go中匿名导入该服务
func init() {
	gina.Register(&AdminServer{})
}

type AdminServer struct {
	*gina.IServer // 必须继承IServer

	httpModule httpmodule.IHttp // 引入模块
}

func (self *AdminServer) OnStart() (err error) {
	// 添加回调函数
	self.httpModule.OnStop(self.exitCallback())

	self.httpModule.Init(self, viper.GetString("App.Addr"), 5, `your_router`)
	err = self.httpModule.Start()

	return
}

// TODO 添加回调函数, 无逻辑可直接删除这个方法
func (self *AdminServer) exitCallback() *httpmodule.CallbackMap {
	callback := httpmodule.NewStopCallbackMap()
	callback.Append("exit", func() {
		gina.Log.Info("这是程序退出后的回调函数, 执行你想要执行的逻辑, 无逻辑可以直接删除这段代码")
	})

	return callback
}
```


> 以下模块必须在 `main.go` 中 **按需匿名导入**

- MySQL

      _ "github.com/soryetong/greasyx/modules/mysqlmodule"

      这样就可以使用 `gina.Db` 获取到Gorm实例

- Redis

      _ "github.com/soryetong/greasyx/modules/redismodule"

      这样就可以使用 `gina.Rdb` 获取到Redis实例

- MongoDB

      _ "github.com/soryetong/greasyx/modules/mongomodule"

      就可以使用 `gina.Mdb` 获取到MongoDB实例

- Casbin

      _ "github.com/soryetong/greasyx/modules/casbinmodule"

      这样就可以搭配内置的`Casbin`中间件来进行权限校验