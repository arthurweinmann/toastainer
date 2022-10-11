package user

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/toastate/toastcloud/internal/api/auth"
	"github.com/toastate/toastcloud/internal/api/settings"
	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/db/objectdb"
	"github.com/toastate/toastcloud/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastcloud/internal/utils"
)

type SigninRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type SigninResponse struct {
	Success bool   `json:"success,omitempty"`
	Session string `json:"session,omitempty"`
}

func Signin(w http.ResponseWriter, r *http.Request) {
	sess, continu := signin(w, r)
	if continu {
		utils.SendSuccess(w, &SigninResponse{
			Success: true,
			Session: sess,
		})
	}
}

func CookieSignin(w http.ResponseWriter, r *http.Request) {
	sess, continu := signin(w, r)
	if continu {
		http.SetCookie(w, &http.Cookie{
			Name:     "toastcloud",
			Value:    sess,
			Path:     "/",
			Domain:   config.APIDomain,
			Expires:  time.Now().Add(24 * time.Hour),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		utils.SendSuccess(w, &SigninResponse{
			Success: true,
		})
	}
}

func signin(w http.ResponseWriter, r *http.Request) (string, bool) {
	req := &SigninRequest{}
	body, err := io.ReadAll(io.LimitReader(r.Body, settings.MaxBodySize))
	if err != nil {
		utils.SendError(w, "could not read request body: "+err.Error(), "invalidBody", 400)
		return "", false
	}

	err = json.Unmarshal(body, req)
	if err != nil {
		utils.SendError(w, "malformed body: "+err.Error(), "invalidBody", 400)
		return "", false
	}

	usr, err := objectdb.Client.GetUserByEmail(req.Email)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "email address not found", "notFound", 404)
			return "", false
		}

		utils.SendInternalError(w, "Signin.GetUserByEmail", err)
		return "", false
	}

	if !utils.CheckPasswordHash(req.Password, usr.Password) {
		utils.SendError(w, "invalid credentials", "invalidCredentials", 401)
		return "", false
	}

	sess, err := auth.CreateSession(usr)
	if err != nil {
		utils.SendInternalError(w, "signup:auth.CreateSession", err)
		return "", false
	}

	return sess, true
}
