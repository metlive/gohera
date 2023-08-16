# golang http

# 使用方法

```go
go get -u 1v1.group/http

import zhttp "1v1.group/http"
内部接口，请使用带ctx的请求方法, 例如GetCtx,PostCtx
外部接口, 请使用不带ctx的请求方法，例如Get, Post

详细可使用的方法，请查看：method.go public.go 两个方法，可以对外使用的


client := zhttp.NewRequest()
//ctx 是*gin.Context nova框架中集成了trace,直接使用接口中*gin.Context
b, err := client.GetCtx(ctx, url).Byte()

如果返回json，需要自动解析
data： = &Data{}
err := client.GetCtx(ctx, url).JsonDecode(data)



如果需要设置header和cookie
client:= zhttp.NewRequest()
b, err:= client.SetHeader("test_test-test", "b")
.SetCookie("test", "d")
.GetCtx(ctx, ts.URL+"?test3=test3&test4=test4").Byte()

也支持批量header
client := zhttp.NewRequest()
b, err := client.SetHeaders(map[string]string)
.SetCookies(map[string]string)
.GetCtx(ctx, ts.URL+"?test3=test3&test4=test4").Byte()

//如果需要响应header和cookie
respHeader := client.GetRespHeader() //header
respCookie := client.GetRespCookie() //cookie

```

