<h1 align="center">greasyx</h1>

<p align="center"> 这是一个为了快速构建一个go项目而生的Go+Gin的Web项目脚手架</p>

<p align="center"> 它是站在巨人的肩膀上而实现的，使用了大量优质的第三方包，介意请慎用</p>

<p align="center">千人千面，它可能不适合你，但完全满足我的需求，欢迎技术讨论，但不喜勿喷</p>

## 介绍

- 它只是我为了快速构建自己的项目而创建的，让我不必在每新建一个新项目都复制这些代码
- 它只是把日常工作中常用的工具和第三方包进行了整合，没有什么高深的技术，也不会有任何的性能影响
- 在项目中, 把常用的MySQL、HTTP、Redis、MongoDB等以模块化的方式按需加载
- 目前 main 分支一直在更新，后续功能完善后，将会提供精简版本的

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
# Group Auth 表示这个模块属于哪个 Group
service SystemTest Group Auth {
    # get 表示请求类型，支持get、post、put、delete, returns前表示请求参数, returns后表示返回参数
    get test1 (TestReq) returns (TestResp)
    post test2 returns
}
```

**Group 是指这个模块所属的组，目前支持 `Public`、`Auth`、`Token`**

> 它会将路由分成不同的组，从而让不同的组使用不同的中间件
    
    `Public`：不使用任何中间件
    
    `Auth`：使用 `Casbin` 权限校验和 `Jwt` token中间件，适用于管理后台
    
    `Token`：使用 `Jwt` token中间件

生成的路由代码如下：

```go
publicGroup := r.Group("/api/v1")
{
    // 健康监测
    publicGroup.GET("/health", func(c *gin.Context) {
        c.JSON(200, "ok")
    })

    // your_router
}

privateAuthGroup := r.Group("/api/v1")
privateAuthGroup.Use(middleware.Casbin()).Use(middleware.Jwt())
{ 
    // your_router
}

privateTokenGroup := r.Group("/api/v1")
privateTokenGroup.Use(middleware.Jwt())
{
    // your_router
}
```

📢注意：这一步不是必须的，只是有了api文件，就可以自动生成`Struct`、`Router`、`Handler`、`Logic`

5. 自动生成代码

```bash
# src表示api文件路径，output表示生成的代码路径
go run main.go autoc src=./api_desc output=./internal

# 或者直接运行，按照指令输入
go run main.go autoc
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
    "Env": "test",
    "Addr": ":18002",
    "Timeout": 1,
    "remark": "RouterPrefix表示你的路由前缀，默认为/api/v1，你可以自定义你的路由前缀",
    "RouterPrefix": "mgr/v1"
  },
  "Db": [
    {
      "Dsn": "root:123456@tcp(127.0.0.1:3307)/greasyx-admin?charset=utf8&parseTime=True&loc=Local&timeout=5s",
      "Driver": "mysql",
      "UseOrm": true,
      "remark": "以下配置是可选的，而且有些Driver是不支持有些配置的，如果没有配置，则使用默认配置",
      "LogLevel": 3,
      "EnableLogWriter": false,
      "MaxIdleConn": 10,
      "MaxConn": 200,
      "SlowThreshold": 2
    }
  ],
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
    "ModePath": "",
    "remark": "当你使用了多个数据库时，需要指定一个数据库名，只有一个时，可以忽略这个配置",
    "DbName": "mysql"
  }
}
```

- `App`：表示项目配置，包括项目名、环境、端口、超时时间等

  - `Env`：表示环境，与 `gin` 的 `EnvGinMode` 保持一致，可选项有 `debug`、`test`、`release`
  
  - `RouterPrefix`：路由前缀，非必填，但当你使用 **`Casbin`、`Limiter`这两个中间件时，将可以减少代码量**
  

- `Db`：表示数据库配置，包括DSN(必要的)、日志级别、最大空闲连接数、最大连接数、慢查询阈值等
    
  - `Driver`：数据库驱动，目前支持 `MySQL`、`PostgresSQL`、`SQLite`、`SQLServer`、`Oracle`

    - 这里也可以支持多个数据库，比如 MySQL 有一主多从，那么你需要添加多个配置，并把 `Driver` 的值改为 "mysql_master", "mysql_slave1", "mysql_slave2"
        
    - 但是必须以驱动名作为前缀，以下划线分割

  - `UseOrm`：是否使用ORM，`true` 则使用 `gorm`，`false` 则使用 `sqlx`
  
  - `EnableLogWriter`：是否使用zap日志记录数据库日志


- `Redis`：表示Redis配置，包括地址、密码、数据库、是否集群等


- `Mongo`：表示MongoDB配置，包括地址、用户名、密码、数据库等


- `Log`：表示日志配置(非必要,都有默认值)，包括日志路径、模式、是否开启Recover、最大文件大小、最大备份数、最大保存天数、是否压缩等

  - `Logrotate` 是否开启日志轮转，默认是开启的。由于日志是按照日期分目录的，这个值为 `true` 时，会在每天晚上 24 点按照日期重新创建日志目录

    - 如果你使用了 `Linux` 自带的 `logrotate` ，那么建议 `Logrotate` 设置为 `false`

  - `Mode` 支持: `file`写入文件，`both`写入文件和控制台，`console`写入控制台，`close`不写入任何地方
  
  - `Recover` zap日志库在你项目启动后删除已经生成的日志文件，将不会自动创建文件并继续写入，但如果这个设置为 `true` 则会检查并重新创建文件，但有一定的性能影响


- `Casbin`：表示Casbin配置，包括模式路径等, 没有该路径时会使用 `greasyx` 提供的默认配置

    - 目前只支持 MySQL 

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

- Db，需要先阅读一下[配置文件](#配置文件)

      _ "github.com/soryetong/greasyx/modules/dbmodule"

      当你这个匿名导入后，`greasyx` 会告诉你该怎么样使用 db，你将在控制台看到以下输出：

        INFO	✅ 提示: `mysql_master` 模块加载成功, 你可以使用 `gina.GetSqlx(mysql_master)` 进行SQL操作

        INFO	✅ 提示: `mysql_slave` 模块加载成功, 你可以使用 `gina.GetSqlx(mysql_slave)` 进行SQL操作

        INFO	✅ 提示: `mysql` 模块加载成功, 你可以使用 `gina.GMySQL()` 进行ORM操作


注意⚠️⚠️⚠️

**在业务逻辑中，你必须清楚的知道你该以哪种方式操作数据库**


- Redis

      _ "github.com/soryetong/greasyx/modules/redismodule"

      这样就可以使用 `gina.Rdb` 获取到Redis实例

- MongoDB

      _ "github.com/soryetong/greasyx/modules/mongomodule"

      就可以使用 `gina.Mdb` 获取到MongoDB实例

- Casbin

      _ "github.com/soryetong/greasyx/modules/casbinmodule"

      这样就可以搭配内置的`Casbin`中间件来进行权限校验


## QA

1. 如何使用日志链路追踪？

        如果你使用的是 `autoc` 自动生成代码，那么路由中已经使用了 `r.Use(middleware.Begin())` 中间件

        如果你没有使用 `autoc` 自动生成代码，那么你需要在路由中手动加入 `r.Use(middleware.Begin())` 中间件

        在业务逻辑中就可以通过 `gina.Log.WithCtx(ctx)`，实现链路追踪

        这个方案需要确保每个需要记录日志的方法的第一个参数都是 `ctx context.Context`

2. 限流器如何使用？

        限流器是一个中间件，在路由中加入 `r.Use(xmiddleware.Limiter())` 即可，但需要定义规则
      
        它基于 `golang.org/x/time/rate` 实现，目前支持通用限流规则和路由限流规则

        限流规则目前支持文件加载，`json` 和 `yaml` 文件都可以，内容如下：（二选一）

      ```json
      {
         "mode": "uri", 
         "rules": [
            { "Route": "health", "KeyType": "ip", "Rate": 1, "Burst": 5 }
         ]
      }
      ```
      ```yaml
      {
        "mode": "comm", # comm表示通用限流规则，uri则表示路由限流规则
        "rules": [
          { "Route": "*", "KeyType": "ip", "Rate": 1, "Burst": 5 }
        ]
      }
      ```
   
      然后再路由文件中加入以下代码即可

      ```go
      limiterStore := xapp.NewLimiterStoreFromFile("./limiter.json")
      r.Use(xmiddleware.Limiter(limiterStore))
      ```
   
3. Swagger 文档如何使用？

        前提是你使用的 `autoc` 自动生成代码，才会自动生成 `Swagger` 文档

        目前每个 controller 中都提供了 `Swagger` 的注释，而且生成了对应的 `Swagger` yaml 文件

        你有两种方式使用：
            
            1. 使用 `gin-swagger` 配合 `Swagger UI` 实现在线预览

            2. 把 `Swagger` yaml 文件导入到 `apifox`、`postman` 等工具中

        为什么 `greasyx` 没有采用方案 1 实现在线预览？
            
            对于每个项目来说，接口文档都是不可外传的，而且我使用的是自建 yapi 文档平台，所以没有采用方案1

            而且 `gin-swagger` 是一个非常优秀的库，基本的注释已经生成好了，如果你有需要，可以自行实现