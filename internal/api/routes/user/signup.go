package user

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/xid"
	"github.com/toastate/toastcloud/internal/api/settings"
	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/db/objectdb"
	"github.com/toastate/toastcloud/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastcloud/internal/email"
	"github.com/toastate/toastcloud/internal/model"
	"github.com/toastate/toastcloud/internal/utils"
)

type SignupRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type SignupResponse struct {
	Success bool        `json:"success,omitempty"`
	User    *model.User `json:"user,omitempty"`
}

func Signup(w http.ResponseWriter, r *http.Request) {
	req := &SignupRequest{}
	body, err := io.ReadAll(io.LimitReader(r.Body, settings.MaxBodySize))
	if err != nil {
		utils.SendError(w, "could not read request body: "+err.Error(), "invalidBody", 400)
		return
	}

	err = json.Unmarshal(body, req)
	if err != nil {
		utils.SendError(w, "malformed body: "+err.Error(), "invalidBody", 400)
		return
	}

	err = utils.IsValidPassword(req.Password)
	if err != nil {
		utils.SendError(w, "invalid password: "+err.Error(), "invalidBody", 400)
		return
	}

	p, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.SendInternalError(w, "signup:utils.HashPassword", err)
		return
	}

	usr := &model.User{
		ID:       xid.New().String(),
		Email:    req.Email,
		Password: p,
	}

	err = objectdb.Client.CreateUser(usr)
	if err != nil {
		if err == objectdberror.ErrAlreadyExists {
			utils.SendError(w, "this email address is already used", "alreadyUsed", 403)
			return
		}

		utils.SendInternalError(w, "signup:objectdb.Client.CreateUser", err)
		return
	}

	if config.EmailProvider.Name != "" {
		err = email.Client.Send([]string{usr.Email}, "Toastcloud Signup", "thanks for signin up with Toastcloud", email.SignupTemplate())
		if err != nil {
			utils.Error("msg", "sendSignupEmail", err)
		}
	}

	utils.SendSuccess(w, &SignupResponse{
		Success: true,
		User:    usr,
	})
}
