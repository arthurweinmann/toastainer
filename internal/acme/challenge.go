package acme

import (
	"context"
	"fmt"
	"time"

	"github.com/toastate/toastainer/internal/db/redisdb"
)

type HTTPChallenger struct {
}

func (c *HTTPChallenger) Present(domain, token, keyAuth string) error {
	fmt.Println("HTTP Challenger present", domain, token, keyAuth)

	return redisdb.GetClient().Set(context.Background(), "certChal_"+domain+"_"+token, keyAuth, 30*time.Minute).Err()
}

func (c *HTTPChallenger) CleanUp(domain, token, keyAuth string) error {
	return redisdb.GetClient().Del(context.Background(), "certChal_"+domain+"_"+token).Err()
}

func GetChallenge(domain, token string) ([]byte, error) {
	return redisdb.GetClient().Get(context.Background(), "certChal_"+domain+"_"+token).Bytes()
}
