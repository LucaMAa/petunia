package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/redis/go-redis/v9"
)

type TranslationService struct {
	redis *redis.Client
	base  string
}

func NewTranslationService(r *redis.Client) *TranslationService {
	base := "translations"
	if v := os.Getenv("TRANSLATIONS_DIR"); v != "" {
		base = v
	}
	return &TranslationService{redis: r, base: base}
}

func (s *TranslationService) GetTranslations(ctx context.Context, locale string) (map[string]string, error) {
	out := make(map[string]string)
	if s.redis != nil {
		key := "translations:" + locale
		if v, err := s.redis.Get(ctx, key).Result(); err == nil {
			_ = json.Unmarshal([]byte(v), &out)
			return out, nil
		}
	}

	tryPaths := []string{
		filepath.Join(s.base, locale+".json"),
		filepath.Join(s.base, locale[:2]+".json"),
	}

	for _, p := range tryPaths {
		if _, err := os.Stat(p); err == nil {
			b, err := os.ReadFile(p) // nolint:gosec
			if err == nil {
				_ = json.Unmarshal(b, &out)
				if s.redis != nil {
					if jb, err := json.Marshal(out); err == nil {
						_ = s.redis.Set(ctx, "translations:"+locale, jb, 24*time.Hour).Err()
					}
				}
				return out, nil
			}
		}
	}

	return out, nil
}
