package user

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/toastate/toastainer/internal/api/auth"
	"github.com/toastate/toastainer/internal/api/routes/toaster"
	"github.com/toastate/toastainer/internal/api/settings"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectstorage/objectstoragerror"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/email"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

type ForgottenPasswordSendLinkRequest struct {
	Email string `json:"email,omitempty"`
}

type ForgottenPasswordSendLinkResponse struct {
	Success bool `json:"success,omitempty"`
}

func ForgottenPasswordSendLink(w http.ResponseWriter, r *http.Request) {
	req := &ForgottenPasswordSendLinkRequest{}
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

	err = utils.IsValidEmail(req.Email)
	if err != nil {
		utils.SendError(w, "invalid email address: "+err.Error(), "invalidBody", 400)
		return
	}

	usr, err := objectdb.Client.GetUserByEmail(req.Email)
	if err != nil {
		if err == objectstoragerror.ErrNotFound {
			// Sending success is necessary not to leak email addresses
			utils.SendSuccess(w, &ForgottenPasswordSendLinkResponse{Success: true})
			return
		}

		utils.SendInternalError(w, "ForgottenPasswordSendLink:objectdb.Client.GetUserByEmail", err)
		return
	}

	if config.EmailProvider.Name == "" {
		utils.SendError(w, "your toastainer instance is not configured to send emails", "notConfigured", 403)
		return
	}

	token, err := utils.UniqueSecureID120()
	if err != nil {
		utils.SendInternalError(w, "ForgottenPasswordSendLink:utils.UniqueSecureID120", err)
		return
	}

	err = redisdb.GetClient().Set(context.Background(), "pwdreset_"+token, usr.ID, 15*time.Minute).Err()
	if err != nil {
		utils.SendInternalError(w, "ForgottenPasswordSendLink:redis.Set", err)
		return
	}

	link := "https://" + config.DashboardDomain + "/reset-password.html?token=" + token

	err = email.Client.Send([]string{usr.Email}, "Toastainer Password Reset", "Please follow this link in order to reset your password", email.ResetPasswordTemplate(link))
	if err != nil {
		utils.SendInternalError(w, "ForgottenPasswordSendLink:email.Send", err)
		return
	}

	utils.SendSuccess(w, &ForgottenPasswordSendLinkResponse{Success: true})
}

type ForgottenPasswordResetRequest struct {
	Token       string `json:"token,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}

type ForgottenPasswordResetResponse struct {
	Success bool `json:"success,omitempty"`
}

func ForgottenPasswordReset(w http.ResponseWriter, r *http.Request) {
	req := &ForgottenPasswordResetRequest{}
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

	err = utils.IsValidPassword(req.NewPassword)
	if err != nil {
		utils.SendError(w, "invalid new password: "+err.Error(), "invalidBody", 400)
		return
	}

	usrid, err := redisdb.GetClient().Get(context.Background(), "pwdreset_"+req.Token).Result()
	if err != nil {
		if err == redisdb.ErrNil {
			utils.SendError(w, "token not found", "notFound", 404)
			return
		}

		utils.SendInternalError(w, "ForgottenPasswordReset:redis.Get", err)
		return
	}

	usr, err := objectdb.Client.GetUserByID(usrid)
	if err != nil {
		utils.SendInternalError(w, "ForgottenPasswordReset:objectdb.Get", err)
		return
	}

	p, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.SendInternalError(w, "ForgottenPasswordReset:utils.HashPassword", err)
		return
	}

	usr.Password = p

	err = objectdb.Client.UpdateUser(usr)
	if err != nil {
		utils.SendInternalError(w, "ForgottenPasswordReset:objectdb.update", err)
		return
	}

	utils.SendSuccess(w, &ForgottenPasswordResetResponse{Success: true})
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
		Name:     "toastainer",
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

type GetUserResponse struct {
	Success bool        `json:"success,omitempty"`
	User    *model.User `json:"user,omitempty"`
}

func GetUser(w http.ResponseWriter, r *http.Request, userid string) {
	usr, err := objectdb.Client.GetUserByID(userid)
	if err != nil {
		utils.SendInternalError(w, "GetUser:objectdb.Client.GetUserByID", err)
		return
	}

	utils.SendSuccess(w, &GetUserResponse{
		Success: true,
		User:    usr,
	})
}
