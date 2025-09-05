package logger

import (
	"sync"
)

// Factory 日志工厂实现
type SlogFactory struct {
	loggers       map[string]Logger
	defaultLogger Logger
	mu            sync.RWMutex
}

// NewSlogFactory 创建新的日志工厂
func NewSlogFactory() *SlogFactory {
	return &SlogFactory{
		loggers: make(map[string]Logger),
	}
}

// CreateLogger 创建日志器实例
func (f *SlogFactory) CreateLogger(config *Config) (Logger, error) {
	logger, err := NewSlogLogger(config)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// 存储日志器
	if config.ServiceName != "" {
		f.loggers[config.ServiceName] = logger
	}

	// 如果是第一个日志器，设为默认
	if f.defaultLogger == nil {
		f.defaultLogger = logger
	}

	return logger, nil
}

// GetLogger 获取已创建的日志器
func (f *SlogFactory) GetLogger(name string) Logger {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if logger, exists := f.loggers[name]; exists {
		return logger
	}

	return f.defaultLogger
}

// SetDefault 设置默认日志器
func (f *SlogFactory) SetDefault(logger Logger) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defaultLogger = logger
}

// GetDefault 获取默认日志器
func (f *SlogFactory) GetDefault() Logger {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.defaultLogger
}

// 全局工厂实例
var globalFactory = NewSlogFactory()

// 便利方法 - 创建默认日志器
func New(serviceName string) Logger {
	config := DefaultConfig(serviceName)
	logger, err := globalFactory.CreateLogger(config)
	if err != nil {
		// 回退到标准输出
		return &SlogLogger{
			config:  config,
			service: serviceName,
			fields:  make(Fields),
		}
	}
	return logger
}

// 便利方法 - 使用配置创建日志器
func NewWithConfig(config *Config) Logger {
	logger, err := globalFactory.CreateLogger(config)
	if err != nil {
		// 回退到标准输出
		return &SlogLogger{
			config:  config,
			service: config.ServiceName,
			fields:  make(Fields),
		}
	}
	return logger
}

// 便利方法 - 获取日志器
func GetLogger(name string) Logger {
	return globalFactory.GetLogger(name)
}

// 便利方法 - 设置默认日志器
func SetDefault(logger Logger) {
	globalFactory.SetDefault(logger)
}

// 便利方法 - 获取默认日志器
func GetDefault() Logger {
	return globalFactory.GetDefault()
}
