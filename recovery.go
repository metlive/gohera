package gohera

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/metlive/gohera/logger"
)

type panicEx struct {
	Err     string `json:"error"`
	Request string `json:"request"`
	Stack   string `json:"stack"`
}

// HandlerRecovery 捕获 Panic 并恢复，记录堆栈信息
func HandlerRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				// 安全类型断言：panic 的值可能不是 error 类型（如 panic(42)、panic("str")）
				if e, ok := err.(error); ok {
					var ne *net.OpError
					if errors.As(e, &ne) {
						var se *os.SyscallError
						if errors.As(ne.Err, &se) {
							if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
								brokenPipe = true
							}
						}
					}
				}
				pe := &panicEx{}
				pe.Err = fmt.Sprintf("%v", err)
				if brokenPipe {
					pJson, _ := json.Marshal(pe)
					logger.Error(c, string(pJson), nil)
					JsonAbort(c, ErrSystem, pe.Err)
					return
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				request := strings.Replace(string(httpRequest), "\r", "|", -1)
				req := strings.Replace(request, "\n", "|", -1)
				pe.Request = req

				if stack {
					stack1 := strings.Replace(string(debug.Stack()), "\r", "|", -1)
					stack2 := strings.Replace(stack1, "\n", "|", -1)
					pe.Stack = stack2
				}
				pJson, _ := json.Marshal(pe)
				logger.Error(c, string(pJson), nil)
				JsonAbort(c, ErrSystem, pe.Err)
			}
		}()
		c.Next()
	}
}
