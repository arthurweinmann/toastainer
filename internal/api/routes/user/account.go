package user

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/toastate/toastcloud/internal/api/auth"
	"github.com/toastate/toastcloud/internal/api/routes/toaster"
	"github.com/toastate/toastcloud/internal/api/settings"
	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/db/objectdb"
	"github.com/toastate/toastcloud/internal/utils"
)

type ForgottenPasswordSendLinkRequest struct {
	CurrentPassword string `json:"current_password,omitempty"`
	NewPassword     string `json:"new_password,omitempty"`
}

type ForgottenPasswordSendLinkResponse struct {
	Success bool `json:"success,omitempty"`
}

func ForgottenPasswordSendLink(w http.ResponseWriter, r *http.Request) {
}

type ForgottenPasswordResetRequest struct {
	CurrentPassword string `json:"current_password,omitempty"`
	NewPassword     string `json:"new_password,omitempty"`
}

type ForgottenPasswordResetResponse struct {
	Success bool `json:"success,omitempty"`
}

func ForgottenPasswordReset(w http.ResponseWriter, r *http.Request) {
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password,omitempty"`
	NewPassword     string `json:"new_password,omitempty"`
}

type UpdatePasswordResponse struct {
	Success bool `json:"success,omitempty"`
}

func UpdatePassword(w http.ResponseWriter, r *http.Request, userid string) {
	req := &UpdatePasswordRequest{}
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

	usr, err := objectdb.Client.GetUserByID(userid)
	if err != nil {
		utils.SendInternalError(w, "User.UpdatePassword:objectdb.Client.GetUserByID", err)
		return
	}

	if !utils.CheckPasswordHash(req.CurrentPassword, usr.Password) {
		utils.SendError(w, "invalid current password", "invalidCredentials", 403)
		return
	}

	err = utils.IsValidPassword(req.NewPassword)
	if err != nil {
		utils.SendError(w, "invalid new password: "+err.Error(), "invalidBody", 400)
		return
	}

	p, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.SendInternalError(w, "UpdatePassword:utils.HashPassword", err)
		return
	}

	usr.Password = p

	err = objectdb.Client.UpdateUser(usr)
	if err != nil {
		utils.SendInternalError(w, "UpdatePassword:objectdb.Client.UpdateUser", err)
		return
	}

	utils.SendSuccess(w, &UpdatePasswordResponse{
		Success: true,
	})
}

type SignoutResponse struct {
	Success bool `json:"success,omitempty"`
}

func Signout(w http.ResponseWriter, r *http.Request, userid, sessToken string) {
	err := auth.DeleteSession(sessToken)
	if err != nil {
		utils.SendInternalError(w, "DeleteAccount:auth.DeleteSession", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "toastcloud",
		Value:    "",
		Path:     "/",
		Domain:   config.APIDomain,
		Expires:  time.Now().Add(24 * time.Hour),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	utils.SendSuccess(w, &SignoutResponse{Success: true})
}

type DeleteAccountRequest struct {
	Password string `json:"password,omitempty"`
}

type DeleteAccountResponse struct {
	Success bool `json:"success,omitempty"`
}

func DeleteAccount(w http.ResponseWriter, r *http.Request, userid, sessToken string) {
	req := &DeleteAccountRequest{}
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

	usr, err := objectdb.Client.GetUserByID(userid)
	if err != nil {
		utils.SendInternalError(w, "DeleteAccount:objectdb.Client.GetUserByID", err)
		return
	}

	if !utils.CheckPasswordHash(req.Password, usr.Password) {
		utils.SendError(w, "invalid credentials", "invalidCredentials", 401)
		return
	}

	toasters, err := objectdb.Client.ListUsertoasters(userid)
	if err != nil {
		utils.SendInternalError(w, "DeleteAccount:objectdb.Client.ListUsertoasters", err)
		return
	}

	for i := 0; i < len(toasters); i++ {
		if !toaster.DeleteToasterHelper(w, userid, toasters[i]) {
			return // DeleteToasterHelper sends the error to the client
		}
	}

	err = objectdb.Client.DeleteAllSubDomainFromUser(userid)
	if err != nil {
		utils.SendInternalError(w, "DeleteAccount:objectdb.Client.DeleteAllSubDomainFromUser", err)
		return
	}

	err = auth.DeleteSession(sessToken)
	if err != nil {
		utils.SendInternalError(w, "DeleteAccount:auth.DeleteSession", err)
		return
	}

	err = objectdb.Client.DelUser(userid)
	if err != nil {
		utils.SendInternalError(w, "DeleteAccount:objectdb.Client.DelUser", err)
		return
	}

	// stats in Redis will be automatically deleted by the billing routine and saved in objectdb
	// they should be kept for some time

	utils.SendSuccess(w, &DeleteAccountResponse{Success: true})
}
