package config

import (
    "fmt"
    "os"
    "strconv"
)

type Config struct {
    Port         string
    GRPCPort     string
    PostgresDSN  string
    RedisAddr    string
    RedisPass    string
    JWTSecret    string // for HMAC demo; replace with RSA for prod
    AccessTTLMin int    // minutes
    RefreshTTLHr int    // hours
}

func Load() *Config {
    c := &Config{
        Port:         getEnv("AUTH_PORT", "8080"),
        GRPCPort:     getEnv("AUTH_GRPC_PORT", "9090"),
        PostgresDSN:  getEnv("POSTGRES_DSN", "postgres://postgres:1024@localhost:5432/telegramlite?sslmode=disable"),
        RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
        RedisPass:    getEnv("REDIS_PASS", ""),
        JWTSecret:    getEnv("JWT_SECRET", "dev-secret-please-change"),
        AccessTTLMin: mustInt(getEnv("ACCESS_TTL_MIN", "15")),
        RefreshTTLHr: mustInt(getEnv("REFRESH_TTL_HR", "168")), // 7 days
    }
    return c
}

func getEnv(k, d string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return d
}

func mustInt(s string) int {
    v, err := strconv.Atoi(s)
    if err != nil {
        fmt.Printf("warning: converting %q to int, defaulting to 0\n", s)
        return 0
    }
    return v
}

