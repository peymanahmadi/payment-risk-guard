package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr string

	PostgresDSN string

	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string

	RiskThreshold float64

	VelocityWindow     time.Duration
	VelocityMaxAllowed int

	AmountSpikeLookback   time.Duration
	AmountSpikeMultiplier float64
	AmountSpikeMinHistory float64

	GeoLookbackWindow time.Duration
	GeoMinGap         time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:     getEnv("HTTP_ADDR", ":8080"),
		PostgresDSN:  getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/riskguard?sslmode=disable"),
		KafkaBrokers: splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "payments.transactions"),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "risk-guard-service"),

		RiskThreshold: getEnvFloat("RISK_THRESHOLD", 50),

		VelocityWindow:     getEnvDuration("VELOCITY_WINDOW", 10*time.Minute),
		VelocityMaxAllowed: getEnvInt("VELOCITY_MAX_ALLOWED", 5),

		AmountSpikeLookback:   getEnvDuration("AMOUNT_SPIKE_LOOKBACK", 30*24*time.Hour),
		AmountSpikeMultiplier: getEnvFloat("AMOUNT_SPIKE_MULTIPLIER", 5),
		AmountSpikeMinHistory: getEnvFloat("AMOUNT_SPIKE_MIN_HISTORY", 20),

		GeoLookbackWindow: getEnvDuration("GEO_LOOKBACK_WINDOW", 6*time.Hour),
		GeoMinGap:         getEnvDuration("GEO_MIN_GAP", 3*time.Hour),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
