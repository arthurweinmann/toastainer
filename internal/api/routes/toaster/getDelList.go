package toaster

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/db/objectstorage"
	"github.com/toastate/toastainer/internal/db/objectstorage/objectstoragerror"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

type GetResponse struct {
	Success bool           `json:"success,omitempty"`
	Toaster *model.Toaster `json:"toaster,omitempty"`
}

func Get(w http.ResponseWriter, r *http.Request, userid, toasterid string) {
	toaster, err := objectdb.Client.GetUserToaster(userid, toasterid)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "could not find toaster "+toasterid, "notFound", 404)
			return
		}

		utils.SendInternalError(w, "GetCodeFile:objectdb.Client.GetUserToaster", err)
		return
	}

	utils.SendSuccess(w, &GetResponse{
		Success: true,
		Toaster: toaster,
	})
}

type DelResponse struct {
	Success bool           `json:"success,omitempty"`
	Toaster *model.Toaster `json:"toaster,omitempty"`
}

func Delete(w http.ResponseWriter, r *http.Request, userid, toasterid string) {
	toaster, err := objectdb.Client.GetUserToaster(userid, toasterid)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "could not find toaster "+toasterid, "notFound", 404)
			return
		}

		utils.SendInternalError(w, "Toaster.Del:objectdb.Client.GetUserToaster", err)
		return
	}

	// if an error occurs, DeleteToasterHelper sends it himself to the user
	if DeleteToasterHelper(w, userid, toaster) {
		utils.SendSuccess(w, &GetResponse{
			Success: true,
			Toaster: toaster,
		})
	}
}

func DeleteToasterHelper(w http.ResponseWriter, userid string, toaster *model.Toaster) bool {
	err := objectdb.Client.DelToaster(userid, toaster.ID)
	if err != nil && err != objectdberror.ErrNotFound {
		utils.SendInternalError(w, "Toaster.Del:objectdb.Client.GetUserToaster", err)
		return false
	}

	err = objectstorage.Client.DeleteFolder(filepath.Join("clearcode", toaster.ID))
	if err != nil && err != objectstoragerror.ErrNotFound {
		utils.SendInternalError(w, "Toaster.Del:objectstorage.Client.DeleteFolder", err)
		return false
	}

	err = redisdb.GetClient().Del(context.Background(), "exeinfo_"+toaster.ID).Err()
	if err != nil && err != redisdb.ErrNil {
		utils.SendInternalError(w, "Toaster.Del:redis.Del", err)
		return false
	}

	linkedSubs, err := objectdb.Client.GetLinkedSubDomains(toaster.ID)
	if err != nil && err != objectdberror.ErrNotFound {
		utils.SendInternalError(w, "Toaster.Del:objectdb.Client.GetLinkedSubDomains", err)
		return false
	}

	if len(linkedSubs) > 0 {
		var delInRedis []string
		for i := 0; i < len(linkedSubs); i++ {
			delInRedis = append(delInRedis, "exeinfo_"+linkedSubs[i].Name)
		}
		err = redisdb.GetClient().Del(context.Background(), delInRedis...).Err()
		if err != nil {
			utils.SendInternalError(w, "Toaster.Del:redis.Del", err)
			return false
		}
	}

	err = objectdb.Client.UnlinkAllSubdomainsFromToaster(userid, toaster.ID)
	if err != nil {
		utils.SendInternalError(w, "Toaster.Del:objectdb.Client.BatchDelSubDomains", err)
		return false
	}

	return true
}

type ListResponse struct {
	Success  bool             `json:"success,omitempty"`
	Toasters []*model.Toaster `json:"toasters,omitempty"`
}

func List(w http.ResponseWriter, r *http.Request, userid string) {
	toasters, err := objectdb.Client.ListUsertoasters(userid)
	if err != nil {
		utils.SendInternalError(w, "Toaster.List:objectdb.Client.ListUsertoasters", err)
		return
	}

	utils.SendSuccess(w, &ListResponse{
		Success:  true,
		Toasters: toasters,
	})
}
