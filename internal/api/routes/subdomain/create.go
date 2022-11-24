package subdomain

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/rs/xid"
	"github.com/toastate/toastainer/internal/api/common"
	"github.com/toastate/toastainer/internal/api/settings"
	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

type CreateSubDomainRequest struct {
	Name      string `json:"name,omitempty"`
	ToasterID string `json:"toaster_id,omitempty"`
}

type CreateSubDomainResponse struct {
	Success   bool             `json:"success,omitempty"`
	SubDomain *model.SubDomain `json:"subdomain,omitempty"`
}

func Create(w http.ResponseWriter, r *http.Request, userid string) {
	req := &CreateSubDomainRequest{}
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

	if req.Name == "" {
		utils.SendError(w, "you must provide a subdomain name", "invalidBody", 400)
		return
	}

	if !utils.IsAlphaNumHyphen(req.Name) {
		utils.SendError(w, "invalid domain name, it can only contain alphanumeric characters and hyphens", "invalidBody", 400)
		return
	}

	if strings.HasPrefix(req.Name, "t_") {
		utils.SendError(w, "invalid domain name, it cannot start with t_ because toaster ids start with this prefix", "invalidBody", 400)
		return
	}

	var toaster *model.Toaster
	if req.ToasterID != "" {
		toaster, err = objectdb.Client.GetUserToaster(userid, req.ToasterID)
		if err != nil {
			if err == objectdberror.ErrNotFound {
				utils.SendError(w, "toaster not found", "notFound", 404)
				return
			}

			utils.SendInternalError(w, "CreateSubDomain:objectdb.Client.GetUserToaster", err)
			return
		}
	}

	sub := &model.SubDomain{
		ID:        xid.New().String(),
		Name:      req.Name,
		UserID:    userid,
		ToasterID: req.ToasterID,
	}

	err = objectdb.Client.CreateSubDomain(sub)
	if err != nil {
		if err == objectdberror.ErrAlreadyExists {
			utils.SendError(w, "the subdomain name you requested is already attributed", "alreadyAttributed", 403)
			return
		}

		utils.SendInternalError(w, "CreateSubDomain:objectdb.Client.CreateSubDomain", err)
		return
	}

	if sub.ToasterID != "" {
		err = redisdb.GetClient().Set(context.Background(), "exeinfo_"+sub.Name, common.DumpToaterExeInfo(toaster), 0).Err()
		if err != nil {
			utils.SendInternalError(w, "CreateSubDomain:redis.Set", err)
			return
		}
	}

	utils.SendSuccess(w, &CreateSubDomainResponse{
		Success:   true,
		SubDomain: sub,
	})
}
