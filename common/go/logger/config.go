package logger

import (
	"strings"
	"time"
)

// Config 日志配置结构
type Config struct {
	Level       string `mapstructure:"level" json:"level" yaml:"level"`                   // 日志级别: debug, info, warn, error
	Format      string `mapstructure:"format" json:"format" yaml:"format"`                // 输出格式: json, text
	Output      string `mapstructure:"output" json:"output" yaml:"output"`                // 输出位置: stdout, file
	FilePath    string `mapstructure:"file_path" json:"file_path" yaml:"file_path"`       // 文件路径
	MaxSize     int    `mapstructure:"max_size" json:"max_size" yaml:"max_size"`          // 文件最大大小(MB)
	MaxBackups  int    `mapstructure:"max_backups" json:"max_backups" yaml:"max_backups"` // 保留文件数量
	MaxAge      int    `mapstructure:"max_age" json:"max_age" yaml:"max_age"`             // 文件保留天数
	Compress    bool   `mapstructure:"compress" json:"compress" yaml:"compress"`          // 是否压缩
	ServiceName string `json:"service_name" yaml:"service_name"`                          // 服务名称
}

// Level 日志级别枚举
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String 实现 Stringer 接口
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return "info"
	}
}

// ParseLevel 解析日志级别字符串
func ParseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// DefaultConfig 返回默认配置
func DefaultConfig(serviceName string) *Config {
	return &Config{
		Level:       "info",
		Format:      "json",
		Output:      "stdout",
		FilePath:    "",
		MaxSize:     100,
		MaxBackups:  3,
		MaxAge:      7,
		Compress:    true,
		ServiceName: serviceName,
	}
}

// Fields 通用字段结构
type Fields map[string]interface{}

// ContextKey 上下文键类型
type ContextKey string

const (
	// RequestIDKey 请求ID在上下文中的键
	RequestIDKey ContextKey = "request_id"
	// UserIDKey 用户ID在上下文中的键
	UserIDKey ContextKey = "user_id"
	// TraceIDKey 链路追踪ID在上下文中的键
	TraceIDKey ContextKey = "trace_id"
)

// RequestInfo 请求信息结构
type RequestInfo struct {
	RequestID string        `json:"request_id"`
	Method    string        `json:"method"`
	Path      string        `json:"path"`
	UserAgent string        `json:"user_agent"`
	IP        string        `json:"ip"`
	StartTime time.Time     `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	Status    int           `json:"status"`
	UserID    string        `json:"user_id,omitempty"`
}
