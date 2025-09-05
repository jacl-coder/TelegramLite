package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

// SlogLogger 基于 Go 1.24.7 slog 的日志器实现
type SlogLogger struct {
	logger  *slog.Logger
	config  *Config
	level   slog.Level
	writer  io.WriteCloser
	service string
	fields  Fields // 预设字段
	err     error  // 预设错误
}

// NewSlogLogger 创建新的 slog 日志器
func NewSlogLogger(config *Config) (*SlogLogger, error) {
	if config == nil {
		config = DefaultConfig("unknown")
	}

	// 设置日志级别
	level := parseToSlogLevel(config.Level)

	// 设置输出目标
	var writer io.WriteCloser
	switch config.Output {
	case "file":
		if config.FilePath == "" {
			config.FilePath = "./logs/app.log"
		}

		// 创建日志目录
		if err := os.MkdirAll(filepath.Dir(config.FilePath), 0755); err != nil {
			return nil, err
		}

		// 打开日志文件
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		writer = file
	default:
		writer = os.Stdout
	}

	// 创建处理器选项
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 格式化时间为可读格式
			if a.Key == slog.TimeKey {
				return slog.String("time", a.Value.Time().Format("2006-01-02 15:04:05"))
			}
			// 格式化级别
			if a.Key == slog.LevelKey {
				return slog.String("level", a.Value.String())
			}
			// 格式化调用位置
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					return slog.String("caller", formatCaller(source))
				}
			}
			return a
		},
	}

	// 创建处理器
	var handler slog.Handler
	if config.Format == "text" {
		handler = slog.NewTextHandler(writer, opts)
	} else {
		handler = slog.NewJSONHandler(writer, opts)
	}

	// 创建日志器
	logger := slog.New(handler)

	return &SlogLogger{
		logger:  logger,
		config:  config,
		level:   level,
		writer:  writer,
		service: config.ServiceName,
		fields:  make(Fields),
	}, nil
}

// 实现 Logger 接口 - 基础日志方法
func (s *SlogLogger) Debug(msg string, fields ...Fields) {
	s.DebugContext(context.Background(), msg, fields...)
}

func (s *SlogLogger) Info(msg string, fields ...Fields) {
	s.InfoContext(context.Background(), msg, fields...)
}

func (s *SlogLogger) Warn(msg string, fields ...Fields) {
	s.WarnContext(context.Background(), msg, fields...)
}

func (s *SlogLogger) Error(msg string, fields ...Fields) {
	s.ErrorContext(context.Background(), msg, fields...)
}

// 实现 Logger 接口 - 带上下文的日志方法
func (s *SlogLogger) DebugContext(ctx context.Context, msg string, fields ...Fields) {
	s.logWithContext(ctx, slog.LevelDebug, msg, fields...)
}

func (s *SlogLogger) InfoContext(ctx context.Context, msg string, fields ...Fields) {
	s.logWithContext(ctx, slog.LevelInfo, msg, fields...)
}

func (s *SlogLogger) WarnContext(ctx context.Context, msg string, fields ...Fields) {
	s.logWithContext(ctx, slog.LevelWarn, msg, fields...)
}

func (s *SlogLogger) ErrorContext(ctx context.Context, msg string, fields ...Fields) {
	s.logWithContext(ctx, slog.LevelError, msg, fields...)
}

// 实现 Logger 接口 - 结构化字段方法
func (s *SlogLogger) WithFields(fields Fields) Logger {
	newFields := make(Fields)
	// 复制原有字段
	for k, v := range s.fields {
		newFields[k] = v
	}
	// 添加新字段
	for k, v := range fields {
		newFields[k] = v
	}

	return &SlogLogger{
		logger:  s.logger,
		config:  s.config,
		level:   s.level,
		writer:  s.writer,
		service: s.service,
		fields:  newFields,
		err:     s.err,
	}
}

func (s *SlogLogger) WithField(key string, value interface{}) Logger {
	newFields := make(Fields)
	// 复制原有字段
	for k, v := range s.fields {
		newFields[k] = v
	}
	// 添加新字段
	newFields[key] = value

	return &SlogLogger{
		logger:  s.logger,
		config:  s.config,
		level:   s.level,
		writer:  s.writer,
		service: s.service,
		fields:  newFields,
		err:     s.err,
	}
}

func (s *SlogLogger) WithContext(ctx context.Context) Logger {
	// 这个方法主要是为了接口兼容性，实际上下文在日志方法中传递
	return s
}

// 实现 Logger 接口 - 错误处理
func (s *SlogLogger) WithError(err error) Logger {
	return &SlogLogger{
		logger:  s.logger,
		config:  s.config,
		level:   s.level,
		writer:  s.writer,
		service: s.service,
		fields:  s.fields,
		err:     err,
	}
}

// 实现 Logger 接口 - 关闭日志器
func (s *SlogLogger) Close() error {
	if s.writer != nil && s.writer != os.Stdout && s.writer != os.Stderr {
		return s.writer.Close()
	}
	return nil
}

// 内部方法 - 带上下文的日志记录
func (s *SlogLogger) logWithContext(ctx context.Context, level slog.Level, msg string, fields ...Fields) {
	if !s.logger.Enabled(ctx, level) {
		return
	}

	// 创建属性列表
	attrs := make([]slog.Attr, 0, 10)

	// 添加服务名称
	attrs = append(attrs, slog.String("service", s.service))

	// 从上下文中提取请求信息
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		attrs = append(attrs, slog.String("user_id", userID))
	}

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		attrs = append(attrs, slog.String("trace_id", traceID))
	}

	// 添加预设错误
	if s.err != nil {
		attrs = append(attrs, slog.String("error", s.err.Error()))
	}

	// 添加预设字段
	for k, v := range s.fields {
		attrs = append(attrs, slog.Any(k, v))
	}

	// 添加动态字段
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			attrs = append(attrs, slog.Any(k, v))
		}
	}

	// 记录日志
	s.logger.LogAttrs(ctx, level, msg, attrs...)
}

// 辅助方法 - 将字符串级别转换为 slog.Level
func parseToSlogLevel(levelStr string) slog.Level {
	level := ParseLevel(levelStr)
	switch level {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// 辅助方法 - 格式化调用者信息
func formatCaller(source *slog.Source) string {
	if source == nil {
		return ""
	}

	file := source.File
	if file != "" {
		// 只保留文件名，不保留完整路径
		for i := len(file) - 1; i >= 0; i-- {
			if file[i] == '/' {
				file = file[i+1:]
				break
			}
		}
	}

	return file + ":" + strconv.Itoa(source.Line)
}
