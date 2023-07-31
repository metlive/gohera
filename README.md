# 基于gin定制化web框架

# 特点
* 插件式、轻量级、性能高效
* 支持异常自动恢复
* 支持go mod
* 支持pprof
* 支持多种配置文件：json/toml/yaml等
* 自动参数校验，验证方式丰富
* ORM，Mysql连接池，Mysql集群
* 集成redis，支持连接池
* 集成golang统一日志包
* 支持并发调用第三方服务接口
* 支持定时任务
* 支持健康检查

# 安装方式
具体使用方法细节可以参考https://wiki.zhiyinlou.com/display/businessDahai/Go+Module
```go
// 安装方式
go get -v -t 1v1.group/gohera
// 或者 
import "1v1.group/gohera"

// 设置goproxy
go env -w GOPROXY=https://goproxy.cn/,https://mirrors.aliyun.com/goproxy/,direct
// 设置GOPRIVATE跳过私有库，比如常用的Gitlab或Gitee，关闭校验
go env -w GOPRIVATE=*.gitlab.com,*.gitee.com,*.100tal.com,1v1.group
go env -w GOSUMDB=off
// 开启GO111MODULE
go env -w GO111MODULE=on
// 在项目目录执行
go mod init
go mod tidy 或 go mod vendor  // 生成go.mod和go.sum
```
# 异常恢复
应用异常panic时，框架会捕捉异常并自动恢复重启，请求上下文异常信息和堆栈状态均写入到日志中

# 配置文件
```cassandraql
// 建议使用toml格式  
// 自定义配置 
gohera.GetConfig("a")  
gohera.GetString("a")
```
具体使用参考代码文件  
> gohera/config.go

# 健康检查
框架集成健康检查方法
```cassandraql
GET /healthz
Response {"status": 200, "env": "Development"}
```
# Http
```cassandraql
// 路由   
g := gohera.HttpEngine.Group("/a/b")  
g.POST("/c", func)  
g.GET("/d", func)  
```
针对路由组或某个路由开启中间件拦截
```cassandraql
// 请求日志，默认框架开启请求日志
Engine.Use(HandleAppAccessLog())
```  
 
配置开关
```cassandraql
[http]  
host = "localhost"  
port = 8080  
开启/关闭pprof  
pprof = 1/0  
```

使用参考  
> https://github.com/gin-gonic/gin

健康检查框架内部自动开启  

# 参数校验
```go
// Post参数验证  
type Demo struct {  
    Id        int    `json:"id" binding:"required"`  
    Name      string `json:"name" binding:"required,max=32"`  
}  
params := &Demo{}  
err := c.ShouldBind(params)   

// Get参数验证
type Demo struct {  
    Id        int    `form:"id" binding:"required"`  
    Name      string `form:"name" binding:"required,max=32"`  
}  
params := &Demo{}  
err := c.ShouldBindQuery(params)  
```

更多验证方式参考  
> omitempty  
> lt  
> eq  
> min  
> https://gopkg.in/go-playground/validator.v10

# 日志
```cassandraql
// 配置开关
[log]  
path = "/var/log/trace" 
``` 

* 支持debug/info等5种日志级别  
* 日志按天自动分割  
* 日志文件名采用{appPath} + "/" + {appName} + "_%Y%m%d.log"   

# Mysql ORM
```cassandraql
// 配置开关
[mysql]  
host = "mysql:3341,mysql2:3341,mysql3:3341"  
user = "a"   
password = "a"  
database = "a"  
// 连接池配置  
max_idle_conn = 1   //设置连接池的空闲数大小

max_open_conn = 2    //设置最大打开连接数
max_conn_lifetime = 3 // 连接生命周期
policy = 0|1|2|3|4(0 => Random, 1 => WeightRandom, 2 => RoundRobin, 3 => WeightRoundRobin, 4 => LeastConn
```
* 支持读写分离
* host格式为domain:port，集群地址通过英文逗号进行分割，第一个地址默认是master，其他的为slave地址
* policy为从库的路由策略，支持随机，权重随机，轮询，权重轮询，最小连接
* 底层采用xorm，具体使用方法可以参考http://xorm.io/

```go
// 具体使用方法参考
type TopicRepo struct {
	*gohera.Model
}

func NewTopicRepo() *TopicRepo {
	return &TopicRepo{
		gohera.NewModel("topic", gohera.Mysql),
	}
}

func (t *TopicRepo)GetTopicList(c *gin.Context) (*models.Topic, error) {
	var record []models.Topic
	e := t.Find(&record)
	if e != nil {
		gohera.Error(c, e.Error(), nil)
        return nil, e
	}
    return record, nil 
}

// 默认会通过从库读取数据，如果需要强制读主
engine := gohera.Mysql.getMaster()
engine.Find(&record)
```
# redis
封装redis相关操作
```go
gohera.Redis.Set(key, value)
gohera.Redis.Get(key)
```

# response
```go
// 通用调用
func RenderJson(c *gin.Context, code int, message string, result interface{}) {
	c.JSON(http.StatusOK, newHttpResponse(code, message, result))
}
// 成功请求调用
func RenderJsonSuccess(c *gin.Context, result interface{}) {
	c.JSON(http.StatusOK, newHttpResponse(Success, "",  result))
}
// 失败请求调用
func RenderJsonException(c *gin.Context, errCode int) {
	c.JSON(http.StatusOK, newHttpResponse(errCode, "", nil))
}
// 终止请求
func AbortJson(c *gin.Context, errCode int, message string) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, newHttpResponse(errCode, message, nil))
}
```