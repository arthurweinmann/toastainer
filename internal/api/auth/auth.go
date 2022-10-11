package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/toastate/toastcloud/internal/db/redisdb"
	"github.com/toastate/toastcloud/internal/model"
	"github.com/toastate/toastcloud/internal/utils"
)

func Auth(w http.ResponseWriter, r *http.Request) (*model.User, string, bool) {
	// auth user with cookie or header or whatever
	authToken := r.Header.Get("X-TOASTCLOUD-SESSION")

	if authToken == "" {
		ck, err := r.Cookie("toastcloud")
		if err == nil {
			authToken = ck.Value
		}
	}

	if authToken == "" {
		utils.SendError(w, "could not find authentication cookie or header", "invalidCredentials", 401)
		return nil, authToken, false
	}

	b, err := redisdb.GetClient().Get(context.Background(), "sess_"+authToken).Bytes()
	if err != nil {
		utils.SendError(w, "invalid credentials", "invalidCredentials", 401)
		return nil, authToken, false
	}

	usr := &model.User{}
	err = json.Unmarshal(b, usr)
	if err != nil {
		utils.SendInternalError(w, "auth.Auth:json.Unmarshal", err)
		return nil, authToken, false
	}

	if usr.ID == "" {
		utils.SendInternalError(w, "auth.Auth:json.Unmarshal", fmt.Errorf("retrieved session does not contain userid"))
		return nil, authToken, false
	}

	return usr, authToken, true
}
