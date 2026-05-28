package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const blacklistPrefix = "blacklist:jti:"

type Blacklist struct {
	rdb *redis.Client
}

func NewBlacklist(rdb *redis.Client) *Blacklist {
	return &Blacklist{rdb: rdb}
}

// Add adaugă un jti în blacklist cu TTL egal cu durata rămasă din token
func (b *Blacklist) Add(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil // tokenul e deja expirat, nu are rost să-l blacklistăm
	}
	return b.rdb.Set(ctx, blacklistPrefix+jti, 1, ttl).Err()
}

// IsBlacklisted verifică dacă un jti e revocat
func (b *Blacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	val, err := b.rdb.Exists(ctx, blacklistPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}
