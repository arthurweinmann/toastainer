package toaster

import (
	"net/http"
	"path/filepath"

	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/db/objectstorage"
	"github.com/toastate/toastainer/internal/db/objectstorage/objectstoragerror"
	"github.com/toastate/toastainer/internal/utils"
)

func GetCodeFile(w http.ResponseWriter, r *http.Request, userid, toasterid, filePath string) {
	toaster, err := objectdb.Client.GetUserToaster(userid, toasterid)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "could not find toaster "+toasterid, "notFound", 404)
			return
		}

		utils.SendInternalError(w, "GetCodeFile:objectdb.Client.GetUserToaster", err)
		return
	}

	err = objectstorage.Client.DownloadFileInto(w, filepath.Join("clearcode", toaster.ID, toaster.CodeID, filePath))
	if err != nil {
		if err == objectstoragerror.ErrNotFound {
			utils.SendError(w, "could not find file: "+filePath, "notFound", 404)
			return
		}

		utils.SendInternalError(w, "GetCodeFile:objectstorage.Client.DownloadFileInto", err)
		return
	}
}
