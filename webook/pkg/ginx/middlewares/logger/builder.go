package logger

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
	"io"
	"time"
)

// MiddlewareBuilder 注意点
// 1. 小心日志内容过多，URL 可能很长, 请求体、响应体都可能很长，你要考虑是否完全输出到日志。
// 2. 考虑到 1 的问题，以及用户可能切换不同的日志框架，所以要有足够的灵活性
// 3. 考虑动态开关，结合监听配置文件，要小心并发安全
type MiddlewareBuilder struct {
	allowReqBody  atomic.Bool
	allowRespBody atomic.Bool
	loggerFunc    func(ctx context.Context, al *AccessLog)
}

func NewBuilder(fn func(ctx context.Context, al *AccessLog)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		loggerFunc:    fn,
		allowReqBody:  atomic.Bool{},
		allowRespBody: atomic.Bool{},
	}
}

func (b *MiddlewareBuilder) AllowReqBody(ok bool) *MiddlewareBuilder {
	b.allowReqBody.Store(ok)
	return b
}

func (b *MiddlewareBuilder) AllowRespBody() *MiddlewareBuilder {
	b.allowReqBody.Store(true)
	return b
}
func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		start := time.Now()
		url := ctx.Request.URL.Path
		if len(url) > 1024 {
			url = url[:1024]
		}
		al := &AccessLog{
			Method: ctx.Request.Method,
			Url:    url,
		}
		if ctx.Request.Body != nil && b.allowReqBody.Load() {
			// Body 读完就没了
			body, _ := ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			if len(body) > 1024 {
				body = body[:1024]
			}
			// 这是很消耗 CPU 和内存的操作
			al.ReqBody = string(body)
		}
		if b.allowRespBody.Load() {
			ctx.Writer = responseWriter{
				al:             al,
				ResponseWriter: ctx.Writer,
			}
		}
		// 执行到 业务逻辑
		defer func() {
			al.Duration = time.Since(start).String()
			b.loggerFunc(ctx, al)
		}()

		ctx.Next()
	}
}

type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (w responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w responseWriter) WriteHeader(statusCode int) {
	w.al.StatusCode = statusCode
}

func (w responseWriter) WriteString(data string) (int, error) {
	w.al.RespBody = data
	return w.ResponseWriter.WriteString(data)
}

type AccessLog struct {
	// HTTP 请求的方法
	Method string
	// Url 整个请求 URL
	Url        string
	Duration   string
	ReqBody    string
	RespBody   string
	StatusCode int
}
