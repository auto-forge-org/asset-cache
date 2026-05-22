package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
)

type Config struct {
	Port         string
	CacheSize    int
	SignerKey    []byte
	LogLevel     string
}

func Load() Config {
	cfg := Config{
		Port:      getEnv("PORT", "8080"),
		CacheSize: getEnvInt("CACHE_SIZE", 256),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
	}

	if key := os.Getenv("SIGNER_KEY"); key != "" {
		cfg.SignerKey = []byte(key)
	} else {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			panic("cannot generate signer key: " + err.Error())
		}
		cfg.SignerKey = []byte(hex.EncodeToString(b))
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
