package config

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func InitRedis() {
	if os.Getenv("REDIS_ENABLED") == "false" {
		log.Println("[redis] disabled via REDIS_ENABLED=false")
		return
	}

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")
	db := 0
	if v, err := strconv.Atoi(os.Getenv("REDIS_DB")); err == nil {
		db = v
	}

	Redis = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Redis.Ping(ctx).Err(); err != nil {
		log.Printf("[redis] connection failed: %v — running without Redis", err)
		Redis = nil
		return
	}
}
