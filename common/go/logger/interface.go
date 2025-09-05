package logger

import (
	"context"
)

// Logger 统一日志接口
type Logger interface {
	// 基础日志方法
	Debug(msg string, fields ...Fields)
	Info(msg string, fields ...Fields)
	Warn(msg string, fields ...Fields)
	Error(msg string, fields ...Fields)

	// 带上下文的日志方法
	DebugContext(ctx context.Context, msg string, fields ...Fields)
	InfoContext(ctx context.Context, msg string, fields ...Fields)
	WarnContext(ctx context.Context, msg string, fields ...Fields)
	ErrorContext(ctx context.Context, msg string, fields ...Fields)

	// 结构化字段方法
	WithFields(fields Fields) Logger
	WithField(key string, value interface{}) Logger
	WithContext(ctx context.Context) Logger

	// 错误处理
	WithError(err error) Logger

	// 关闭日志器
	Close() error
}

// Entry 日志条目接口
type Entry interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	WithField(key string, value interface{}) Entry
	WithFields(fields Fields) Entry
	WithError(err error) Entry
}

// Factory 日志工厂接口
type Factory interface {
	// CreateLogger 创建日志器实例
	CreateLogger(config *Config) (Logger, error)
	// GetLogger 获取已创建的日志器
	GetLogger(name string) Logger
	// SetDefault 设置默认日志器
	SetDefault(logger Logger)
	// GetDefault 获取默认日志器
	GetDefault() Logger
}

// Hook 日志钩子接口
type Hook interface {
	// Levels 返回钩子适用的日志级别
	Levels() []Level
	// Fire 触发钩子
	Fire(entry *LogEntry) error
}

// LogEntry 日志条目结构
type LogEntry struct {
	Level     Level                  `json:"level"`
	Time      int64                  `json:"timestamp"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Service   string                 `json:"service"`
	RequestID string                 `json:"request_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
}
