package service

import (
	"context"
	"time"

	"petunia/internal/config"
)

type AvatarCache interface {
	GetDefaultAvatar(ctx context.Context) ([]byte, error)
	SetDefaultAvatar(ctx context.Context, b []byte, ttl time.Duration) error
}

type redisAvatarCache struct{}

func (r *redisAvatarCache) GetDefaultAvatar(ctx context.Context) ([]byte, error) {
	if config.Redis == nil {
		return nil, nil
	}
	return config.Redis.Get(ctx, "default_avatar").Bytes()
}

func (r *redisAvatarCache) SetDefaultAvatar(ctx context.Context, b []byte, ttl time.Duration) error {
	if config.Redis == nil {
		return nil
	}
	return config.Redis.Set(ctx, "default_avatar", b, ttl).Err()
}
