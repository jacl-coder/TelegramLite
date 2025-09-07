package config

import (
	"time"

	"github.com/spf13/viper"
)

// LogConfig 日志配置 - 匹配 common/go/logger.Config
type LogConfig struct {
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

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Auth     AuthConfig     `mapstructure:"auth"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port     int    `mapstructure:"port"`
	GRPCPort int    `mapstructure:"grpc_port"`
	Mode     string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	AuthServiceURL string `mapstructure:"auth_service_url"`
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	ExpireHours        int    `mapstructure:"expire_hours"`
	RefreshExpireHours int    `mapstructure:"refresh_expire_hours"`
}

func (j JWTConfig) ExpireDuration() time.Duration {
	return time.Duration(j.ExpireHours) * time.Hour
}

func (j JWTConfig) RefreshExpireDuration() time.Duration {
	return time.Duration(j.RefreshExpireHours) * time.Hour
}

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	// 设置环境变量支持
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
