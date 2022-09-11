package redisdb

import (
	"context"
	"time"
)

func LockTryOnce(key string, lockExpiration time.Duration) (bool, error) {
	return GetClient().SetNX(context.Background(), "lock_"+key, "", lockExpiration).Result()
}

func Unlock(key string) error {
	return GetClient().Del(context.Background(), "lock_"+key).Err()
}
