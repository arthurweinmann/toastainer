package subdomain

import (
	"context"
	"net/http"

	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

type DeleteResponse struct {
	Success   bool             `json:"success,omitempty"`
	SubDomain *model.SubDomain `json:"subdomain,omitempty"`
}

func Delete(w http.ResponseWriter, r *http.Request, userid, subdomainid string) {
	sub, err := objectdb.Client.GetSubDomain(userid, subdomainid)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "could not find subdomain "+subdomainid, "notFound", 404)
			return
		}

		utils.SendInternalError(w, "Subdomain.Delete:objectdb.Client.GetSubDomain", err)
		return
	}

	err = objectdb.Client.DeleteSubDomain(userid, sub.ID)
	if err != nil {
		utils.SendInternalError(w, "Subdomain.Delete:objectdb.Client.DeleteSubDomain", err)
		return
	}

	err = redisdb.GetClient().Del(context.Background(), "exeinfo_"+sub.Name).Err()
	if err != nil && err != redisdb.ErrNil {
		utils.SendInternalError(w, "Subdomain.Delete:redis.Del", err)
		return
	}

	utils.SendSuccess(w, &DeleteResponse{
		Success:   true,
		SubDomain: sub,
	})
}
