package logger

import (
	"context"
	"net/http"
	"time"
)

// HTTPRequestInfo HTTP请求信息
type HTTPRequestInfo struct {
	Method    string
	Path      string
	UserAgent string
	IP        string
	Status    int
	Duration  time.Duration
}

// HTTPMiddleware HTTP中间件适配器接口
type HTTPMiddleware interface {
	// WrapHandler 包装HTTP处理器
	WrapHandler(handler http.Handler) http.Handler
}

// httpMiddleware HTTP中间件实现
type httpMiddleware struct {
	logger Logger
}

// NewHTTPMiddleware 创建HTTP中间件
func NewHTTPMiddleware(logger Logger) HTTPMiddleware {
	return &httpMiddleware{
		logger: logger,
	}
}

// WrapHandler 包装HTTP处理器
func (m *httpMiddleware) WrapHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 生成请求ID
		requestID := generateRequestID()

		// 创建带请求ID的上下文
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// 包装ResponseWriter以捕获状态码
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		// 记录请求开始
		m.logger.InfoContext(ctx, "HTTP request started",
			Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"user_agent": r.UserAgent(),
				"ip":         getClientIP(r),
			},
		)

		// 处理请求
		handler.ServeHTTP(wrappedWriter, r)

		// 计算处理时间
		duration := time.Since(start)

		// 记录请求完成
		fields := Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     wrappedWriter.statusCode,
			"duration":   duration.Milliseconds(),
			"user_agent": r.UserAgent(),
			"ip":         getClientIP(r),
		}

		// 根据状态码选择日志级别
		if wrappedWriter.statusCode >= 500 {
			m.logger.ErrorContext(ctx, "HTTP request completed with server error", fields)
		} else if wrappedWriter.statusCode >= 400 {
			m.logger.WarnContext(ctx, "HTTP request completed with client error", fields)
		} else {
			m.logger.InfoContext(ctx, "HTTP request completed", fields)
		}
	})
}

// responseWriter 包装ResponseWriter以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getClientIP 获取客户端IP
func getClientIP(r *http.Request) string {
	// 尝试从各种header中获取真实IP
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	return r.RemoteAddr
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	// 简单的时间戳+随机数生成
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		// 使用纳秒时间作为伪随机源
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
