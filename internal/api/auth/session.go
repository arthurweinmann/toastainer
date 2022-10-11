package auth

import (
	"context"
	"encoding/json"

	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

func CreateSession(user *model.User) (string, error) {
	b, err := json.Marshal(user)
	if err != nil {
		return "", nil
	}

	sess, err := utils.UniqueSecureID60()
	if err != nil {
		return "", nil
	}

	err = redisdb.GetClient().Set(context.Background(), "sess_"+sess, b, 0).Err()
	if err != nil {
		return "", nil
	}

	return sess, nil
}

func DeleteSession(token string) error {
	return redisdb.GetClient().Del(context.Background(), "sess_"+token).Err()
}
