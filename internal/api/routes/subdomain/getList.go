package subdomain

import (
	"net/http"

	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

type GetResponse struct {
	Success   bool             `json:"success,omitempty"`
	Subdomain *model.SubDomain `json:"subdomain,omitempty"`
}

func Get(w http.ResponseWriter, r *http.Request, userid, subdomainid string) {
	sub, err := objectdb.Client.GetSubDomain(userid, subdomainid)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "could not find subdomain "+subdomainid, "notFound", 404)
			return
		}

		utils.SendInternalError(w, "subdomain.Get:objectdb.Client.GetSubDomain", err)
		return
	}

	utils.SendSuccess(w, &GetResponse{
		Success:   true,
		Subdomain: sub,
	})
}

type ListResponse struct {
	Success    bool               `json:"success,omitempty"`
	Subdomains []*model.SubDomain `json:"subdomains,omitempty"`
}

func List(w http.ResponseWriter, r *http.Request, userid string) {
	subs, err := objectdb.Client.ListUserSubdomains(userid)
	if err != nil {
		utils.SendInternalError(w, "Subdomain.List:objectdb.Client.ListUserSubdomains", err)
		return
	}

	utils.SendSuccess(w, &ListResponse{
		Success:    true,
		Subdomains: subs,
	})
}
