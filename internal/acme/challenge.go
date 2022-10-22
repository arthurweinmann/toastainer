package acme

import (
	"context"
	"time"

	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/utils"
)

type HTTPChallenger struct {
}

func (c *HTTPChallenger) Present(domain, token, keyAuth string) error {
	utils.Debug("msg", "HTTP Challenger present", "domain", domain, "token", token, "keyauth", keyAuth)

	return redisdb.GetClient().Set(context.Background(), "certChal_"+domain+"_"+token, keyAuth, 30*time.Minute).Err()
}

func (c *HTTPChallenger) CleanUp(domain, token, keyAuth string) error {
	return redisdb.GetClient().Del(context.Background(), "certChal_"+domain+"_"+token).Err()
}

func GetChallenge(domain, token string) ([]byte, error) {
	return redisdb.GetClient().Get(context.Background(), "certChal_"+domain+"_"+token).Bytes()
}
