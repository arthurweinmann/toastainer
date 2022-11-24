package toaster

import (
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/toastate/toastainer/internal/api/settings"
	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/db/objectstorage"
	"github.com/toastate/toastainer/internal/db/objectstorage/objectstoragerror"
	"github.com/toastate/toastainer/internal/utils"
)

func GetToasterPicture(w http.ResponseWriter, r *http.Request, userid, toasterid, lastURLPart string) {
	_, err := objectdb.Client.GetUserToaster(userid, toasterid)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "toaster not found", "notFound", 404)
			return
		}

		utils.SendInternalError(w, "UpdateToasterPictureRoute:objectdb.Client.GetUserToaster", err)
		return
	}

	err = objectstorage.Client.DownloadFileInto(w, filepath.Join("toaster/picture/", toasterid, lastURLPart))
	if err != nil {
		if err == objectstoragerror.ErrNotFound {
			utils.SendError(w, err.Error(), "notFound", 404)
			return
		}

		utils.SendInternalError(w, "GetUserPicture:objectstorage.Client.DownloadFileInto", err)
		return
	}
}

func UpdateToasterPictureRoute(w http.ResponseWriter, r *http.Request, userid, toasterid string) {
	if !utils.IsMultipart(r) {
		utils.SendError(w, "you must use a multipart request", "invalidRequest", 400)
		return
	}

	toaster, err := objectdb.Client.GetUserToaster(userid, toasterid)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "toaster not found", "notFound", 404)
			return
		}

		utils.SendInternalError(w, "UpdateToasterPictureRoute:objectdb.Client.GetUserToaster", err)
		return
	}

	err = r.ParseMultipartForm(settings.MultipartMaxMemory)
	if err != nil {
		utils.SendError(w, "could not read multipart request body: "+err.Error(), "invalidBody", 400)
		return
	}

	var f1 multipart.File
	var ext string

F1:
	for _, v := range r.MultipartForm.File {
		if len(v) == 0 {
			utils.SendError(w, "the request did not contain any image", "invalid", 400)
			return
		}

		f1, err = v[0].Open()
		if err != nil {
			utils.SendError(w, err.Error(), "invalidBody", 400)
			return
		}

		ext = filepath.Ext(v[0].Filename)
		break F1
	}

	if ext == "" {
		ext, err = utils.GuessImageFormat(f1)
		if err != nil {
			utils.SendError(w, "could not guess image format/extension", "invalidBody", 400)
			return
		}

		ext = "." + ext

		_, err = f1.Seek(0, 0)
		if err != nil {
			utils.SendInternalError(w, "UpdateToasterPictureRoute:seek", err)
			return
		}
	}

	err = objectstorage.Client.PushReader(f1, filepath.Join("toaster/picture/", toaster.ID, "/picture"+ext))
	if err != nil {
		utils.SendInternalError(w, "UpdateToasterPictureRoute:objectstorage.Client.PushReader", err)
		return
	}

	toaster.PictureExtension = ext
	err = objectdb.Client.UpdateToaster(toaster)
	if err != nil {
		utils.SendInternalError(w, "UpdateToasterPictureRoute:objectdb.Client.UpdateUser", err)
		return
	}

	utils.SendSuccess(w, &GetResponse{
		Success: true,
		Toaster: completeToasterDynFields(toaster, true),
	})
}
