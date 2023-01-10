package toaster

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/toastate/toastainer/internal/api/settings"
	"github.com/toastate/toastainer/internal/utils"
)

func readRequest(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	var err error

	if utils.IsMultipart(r) {
		err = r.ParseMultipartForm(settings.MultipartMaxMemory)
		if err != nil {
			utils.SendError(w, "could not read multipart request body: "+err.Error(), "invalidBody", 400)
			return false
		}
		js, ok := r.MultipartForm.Value["request"]
		if !ok || len(js) == 0 {
			utils.SendError(w, "you must provide toaster information in json format in the request field of your multipart HTTP request", "invalidBody", 400)
			return false
		}
		err = json.Unmarshal([]byte(js[0]), req)
		if err != nil {
			utils.SendError(w, "malformed body: "+err.Error()+"\n"+js[0], "invalidBody", 400)
			return false
		}
	} else {
		body, err := io.ReadAll(io.LimitReader(r.Body, settings.MaxBodySize))
		if err != nil {
			utils.SendError(w, "could not read request body: "+err.Error(), "invalidBody", 400)
			return false
		}

		err = json.Unmarshal(body, req)
		if err != nil {
			utils.SendError(w, "malformed body: "+err.Error(), "invalidBody", 400)
			return false
		}
	}

	return true
}
