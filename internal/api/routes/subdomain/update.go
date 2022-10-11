package subdomain

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/toastate/toastainer/internal/api/common"
	"github.com/toastate/toastainer/internal/api/settings"
	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

type UpdateSubDomainRequest struct {
	ToasterID *string `json:"toaster_id,omitempty"`
}

type UpdateSubDomainResponse struct {
	Success   bool             `json:"success,omitempty"`
	SubDomain *model.SubDomain `json:"subdomain,omitempty"`
}

func Update(w http.ResponseWriter, r *http.Request, userid, subdomainid string) {
	req := &UpdateSubDomainRequest{}
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

	sub, err := objectdb.Client.GetSubDomain(userid, subdomainid)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "subdomain not found", "notFound", 404)
			return
		}

		utils.SendInternalError(w, "UpdateSubDomain:objectdb.Client.GetSubDomain", err)
		return
	}

	if req.ToasterID == nil {
		utils.SendSuccess(w, &UpdateSubDomainResponse{
			Success:   true,
			SubDomain: sub,
		})
		return
	}

	var toaster *model.Toaster
	if *req.ToasterID != "" {
		toaster, err = objectdb.Client.GetUserToaster(userid, *req.ToasterID)
		if err != nil {
			if err == objectdberror.ErrNotFound {
				utils.SendError(w, "toaster not found", "notFound", 404)
				return
			}

			utils.SendInternalError(w, "CreateSubDomain:objectdb.Client.GetUserToaster", err)
			return
		}
	}

	sub.ToasterID = *req.ToasterID

	err = objectdb.Client.UpdateSubDomain(sub)
	if err != nil {
		utils.SendInternalError(w, "UpdateSubDomain:objectdb.Client.UpdateSubDomain", err)
		return
	}

	if toaster != nil {
		err = redisdb.GetClient().Set(context.Background(), "exeinfo_"+sub.Name, common.DumpToaterExeInfo(toaster), 0).Err()
		if err != nil {
			utils.SendInternalError(w, "CreateSubDomain:redis.Set", err)
			return
		}
	} else {
		err = redisdb.GetClient().Del(context.Background(), "exeinfo_"+sub.Name).Err()
		if err != nil {
			utils.SendInternalError(w, "CreateSubDomain:redis.Del", err)
			return
		}
	}

	utils.SendSuccess(w, &UpdateSubDomainResponse{
		Success:   true,
		SubDomain: sub,
	})
}
