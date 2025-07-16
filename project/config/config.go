package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	WorkerCount  int
	QueueSize    int
	MaxRetries   int
	MetricsPort  string
}

func Load() *Config {
	return &Config{
		Port:         getEnvOrDefault("PORT", "8080"),
		WorkerCount:  getEnvIntOrDefault("WORKER_COUNT", 3),
		QueueSize:    getEnvIntOrDefault("QUEUE_SIZE", 100),
		MaxRetries:   getEnvIntOrDefault("MAX_RETRIES", 3),
		MetricsPort:  getEnvOrDefault("METRICS_PORT", "9090"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}