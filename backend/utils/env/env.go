package env

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

func GetStringEnv(ctx context.Context, logger log.Logger, key, def string) string {
	s, ok := os.LookupEnv(key)
	if !ok {
		logger.LogInfo(ctx, "Using default value %s for %s", def, key)
		return def
	}
	logger.LogInfo(ctx, "Using value %s for %s", s, key)
	return s
}

func GetDurationEnv(ctx context.Context, logger log.Logger, key string, def time.Duration) time.Duration {
	s, ok := os.LookupEnv(key)
	if !ok {
		logger.LogInfo(ctx, "Using default value %v for %s", def, key)
		return def
	}

	parsed, err := time.ParseDuration(s)
	if err != nil {
		logger.LogInfo(ctx,
			"Using default value %v for %s because the given Duration could not be parsed:%v - %v", def, key, s, err)
		return def
	}

	logger.LogInfo(ctx, "Using value %v for %s", parsed, key)
	return parsed
}

func GetIntEnv(ctx context.Context, logger log.Logger, key string, def int) int {
	s, ok := os.LookupEnv(key)
	if !ok {
		logger.LogInfo(ctx, "Using default value %v for %s", def, key)
		return def
	}

	parsed, err := strconv.Atoi(s)
	if err != nil {
		logger.LogInfo(ctx,
			"Using default value %v for %s because the given Integer could not be parsed:%v - %v", def, key, s, err)
		return def
	}

	logger.LogInfo(ctx, "Using value %v for %s", parsed, key)
	return parsed
}

func GetInt64Env(ctx context.Context, logger log.Logger, key string, def int64) int64 {
	s, ok := os.LookupEnv(key)
	if !ok {
		logger.LogInfo(ctx, "Using default value %v for %s", def, key)
		return def
	}

	parsed, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		logger.LogInfo(ctx,
			"Using default value %v for %s because the given Integer could not be parsed:%v - %v", def, key, s, err)
		return def
	}

	logger.LogInfo(ctx, "Using value %v for %s", parsed, key)
	return parsed
}
