package toaster

import (
	"encoding/base32"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	auther "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rs/xid"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/db/objectstorage"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

type UpdateRequest struct {
	FileContents   [][]byte `json:"file_contents,omitempty"`
	FilePaths      []string `json:"file_paths,omitempty"`
	GitURL         *string  `json:"git_url,omitempty"`
	GitUsername    *string  `json:"git_username,omitempty"`
	GitAccessToken *string  `json:"git_access_token,omitempty"`
	GitPassword    *string  `json:"git_password,omitempty"`
	GitBranch      *string  `json:"git_branch,omitempty"`

	// Executed in the root directory of the FilePaths
	BuildCmd []string `json:"build_command,omitempty"`
	ExeCmd   []string `json:"execution_command,omitempty"`
	Env      []string `json:"environment_variables,omitempty"`

	JoinableForSec       *int `json:"joinable_for_seconds,omitempty"`
	MaxConcurrentJoiners *int `json:"max_concurrent_joiners,omitempty"`
	TimeoutSec           *int `json:"timeout_seconds,omitempty"`

	Name     *string  `json:"name,omitempty"`
	Readme   *string  `json:"readme,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
}

type UpdateResponse struct {
	Success   bool           `json:"success,omitempty"`
	BuildLogs []byte         `json:"build_logs,omitempty"` // base64 encoded string, see https://pkg.go.dev/encoding/json#Marshal
	BuildID   string         `json:"build_id,omitempty"`
	Toaster   *model.Toaster `json:"toaster,omitempty"`
}

func Update(w http.ResponseWriter, r *http.Request, userid, toasterID string) {
	req := &UpdateRequest{}
	if !readRequest(w, r, req) {
		// error is sent to the client in readRequest
		return
	}

	codeUpdate := len(req.FilePaths) > 0 || (req.GitURL != nil && *req.GitURL != "")
	if !codeUpdate && utils.IsMultipart(r) {
		for _, v := range r.MultipartForm.File {
			if len(v) > 0 {
				codeUpdate = true
				break
			}
		}
	}

	if len(req.FilePaths) != len(req.FileContents) {
		utils.SendError(w, "file_paths and file_contents must have the same length", "invalidBody", 400)
		return
	}

	toaster, err := objectdb.Client.GetUserToaster(userid, toasterID)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "could not find toaster "+toasterID, "notFound", 404)
			return
		}

		utils.SendInternalError(w, "CreateToaster:objectdb.Client.GetUserToaster", err)
		return
	}

	if req.BuildCmd != nil {
		toaster.BuildCmd = req.BuildCmd
	}
	if req.ExeCmd != nil {
		toaster.ExeCmd = req.ExeCmd
	}
	if req.Env != nil {
		toaster.Env = req.Env
	}
	if req.JoinableForSec != nil {
		toaster.JoinableForSec = *req.JoinableForSec
	}
	if req.MaxConcurrentJoiners != nil {
		toaster.MaxConcurrentJoiners = *req.MaxConcurrentJoiners
	}
	if req.TimeoutSec != nil {
		toaster.TimeoutSec = *req.TimeoutSec
	}
	if req.Name != nil {
		toaster.Name = *req.Name
	}
	if req.Readme != nil {
		toaster.Readme = *req.Readme
	}
	if req.Keywords != nil {
		toaster.Keywords = req.Keywords
	}
	toaster.LastModified = time.Now().Unix()
	toaster.Created = time.Now().Unix()

	var buildLogs []byte
	var tmpFolder string
	var buildid string
	if codeUpdate {
		toaster.CodeID = xid.New().String()

		tmpdir, err := ioutil.TempDir("", "toasterCode_"+toaster.ID)
		if err != nil {
			utils.SendInternalError(w, "CreateToaster:ioutil.TempDir", err)
			return
		}
		defer os.RemoveAll(tmpdir)

		tmpFolder = filepath.Join(tmpdir, "code")

		err = os.MkdirAll(tmpFolder, 0755)
		if err != nil {
			utils.SendInternalError(w, "CreateToaster:os.MkdirAll:"+tmpFolder, err)
			return
		}

		switch {
		case utils.IsMultipart(r):
			for _, v := range r.MultipartForm.File {
				for i := 0; i < len(v); i++ {
					f1, err := v[i].Open()
					if err != nil {
						utils.SendError(w, fmt.Sprintf("could not open multipart file %s: %v", v[i].Filename, err), "invalidFile", 400)
						return
					}

					fld32, err := base32.StdEncoding.DecodeString(v[i].Filename)
					if err != nil {
						utils.SendError(w, fmt.Sprintf("invalid filepath %s, filepath must be base32 encoded: %v", v[i].Filename, err), "invalidFilename", 400)
						return
					}

					dest := filepath.Join(tmpdir, "code", string(fld32))
					if utils.DirExists(dest) {
						f1.Close()
						utils.SendError(w, fmt.Sprintf("%s is a directory", string(fld32)), "invalidFilename", 501)
						return
					}

					err = os.MkdirAll(filepath.Dir(dest), 0755)
					if err != nil {
						f1.Close()
						utils.SendInternalError(w, "CreateToaster:os.MkdirAll:"+filepath.Dir(dest), err)
						return
					}

					f2, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR, 0644)
					if err != nil {
						f1.Close()
						utils.SendInternalError(w, "CreateToaster:os.OpenFile:"+dest, err)
						return
					}

					_, err = io.Copy(f2, f1)
					f1.Close()
					f2.Close()
					if err != nil {
						utils.SendInternalError(w, "CreateToaster:io.Copy:"+dest, err)
						return
					}
				}
			}
		case len(req.FilePaths) > 0:
			for i := 0; i < len(req.FilePaths); i++ {
				dest := filepath.Join(tmpdir, "code", req.FilePaths[i])

				err = os.MkdirAll(filepath.Dir(dest), 0755)
				if err != nil {
					utils.SendInternalError(w, "CreateToaster:os.MkdirAll:"+filepath.Dir(dest), err)
					return
				}

				err = os.WriteFile(dest, req.FileContents[i], 0644)
				if err != nil {
					utils.SendInternalError(w, "CreateToaster:os.WriteFile:"+dest, err)
					return
				}
			}
		case (req.GitURL != nil && *req.GitURL != ""):
			opts := &git.CloneOptions{
				URL:               *req.GitURL,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			}
			if req.GitBranch != nil && *req.GitBranch != "" {
				opts.ReferenceName = plumbing.NewBranchReferenceName(*req.GitBranch)
				opts.SingleBranch = true
			}
			if req.GitUsername != nil && *req.GitUsername != "" {
				tmp := &auther.BasicAuth{
					Username: *req.GitUsername,
				}

				switch {
				case req.GitAccessToken != nil && *req.GitAccessToken != "":
					tmp.Password = *req.GitAccessToken
				case req.GitUsername != nil && *req.GitUsername != "":
					tmp.Password = *req.GitUsername
				}

				opts.Auth = tmp
			}
			_, err = git.PlainClone(tmpFolder, false, opts)
			if err != nil {
				utils.SendError(w, err.Error(), "gitError", 404)
				return
			}
		}

		err = filepath.Walk(tmpFolder, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				rel, err := filepath.Rel(tmpFolder, path)
				if err != nil {
					return err
				}
				toaster.Files = append(toaster.Files, rel)
			}

			return os.Chown(path, config.Runner.NonRootUID, config.Runner.NonRootGID)
		})
		if err != nil {
			utils.SendInternalError(w, "CreateToaster:filepath.Walk:"+tmpFolder, err)
			return
		}

		cmd := exec.Command("tar", "-czf", "../code.tar.gz", "./")
		cmd.Dir = tmpFolder
		out, err := cmd.CombinedOutput()
		if err != nil {
			utils.SendInternalError(w, "CreateToaster:tar", fmt.Errorf("%v: %v", err, string(out)))
			return
		}

		tarpath := filepath.Join(tmpFolder, "../code.tar.gz")

		buildid, buildLogs, err = buildToasterCode(toaster, tarpath)
		if err != nil {
			if err == ErrUnsuccessfulBuild {
				utils.SendError(w, fmt.Sprintf("build failed: %s", string(buildLogs)), "buildFailed", 400)
			} else {
				utils.SendInternalError(w, "CreateToaster:buildToasterCode", err)
			}
			return
		}
	}

	err = objectdb.Client.UpdateToaster(toaster)
	if err != nil {
		utils.SendInternalError(w, "UpdateToaster:objectdb.Client.CreateToaster", err)
		return
	}

	if codeUpdate {
		err = objectstorage.Client.UploadFolder(tmpFolder, filepath.Join("clearcode", toaster.ID, toaster.CodeID))
		if err != nil {
			utils.SendInternalError(w, "CreateToaster:objectstorage.Client.UploadFolder", err)
			return
		}
	} else {
		// if we did not call buildToasterCode that usually handles propagating exe info into redis
		err = setToasterExeInfo(toaster)
		if err != nil {
			utils.SendInternalError(w, "UpdateToaster:setToasterExeInfo", err)
			return
		}
	}

	utils.SendSuccess(w, &UpdateResponse{
		Success:   true,
		BuildLogs: buildLogs,
		BuildID:   buildid,
		Toaster:   toaster,
	})
}
