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
	"github.com/toastate/toastainer/internal/db/objectstorage"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

type CreateRequest struct {
	FileContents   [][]byte `json:"file_contents,omitempty"`
	FilePaths      []string `json:"file_paths,omitempty"`
	GitURL         string   `json:"git_url,omitempty"`
	GitUsername    string   `json:"git_username,omitempty"`
	GitAccessToken string   `json:"git_access_token,omitempty"`
	GitPassword    string   `json:"git_password,omitempty"`
	GitBranch      string   `json:"git_branch,omitempty"`

	Image string `json:"image,omitempty"`

	// Executed in the root directory of the FilePaths
	BuildCmd []string `json:"build_command,omitempty"`
	ExeCmd   []string `json:"execution_command,omitempty"`
	Env      []string `json:"environment_variables,omitempty"`

	JoinableForSec       int `json:"joinable_for_seconds,omitempty"`
	MaxConcurrentJoiners int `json:"max_concurrent_joiners,omitempty"`
	TimeoutSec           int `json:"timeout_seconds,omitempty"`

	Name     string   `json:"name,omitempty"`
	Readme   string   `json:"readme,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
}

type CreateResponse struct {
	Success    bool           `json:"success"`
	BuildLogs  []byte         `json:"build_logs,omitempty"`  // base64 encoded string, see https://pkg.go.dev/encoding/json#Marshal
	BuildError []byte         `json:"build_error,omitempty"` // base64 encoded string, see https://pkg.go.dev/encoding/json#Marshal
	BuildID    string         `json:"build_id,omitempty"`
	Toaster    *model.Toaster `json:"toaster,omitempty"`
}

func Create(w http.ResponseWriter, r *http.Request, userid string) {
	req := &CreateRequest{}
	if !readRequest(w, r, req) {
		// error is sent to the client in readRequest
		return
	}

	if req.Image == "" {
		utils.SendError(w, "you must provide an OS image name", "invalidBody", 400)
		return
	}

	codecheck := len(req.FilePaths) > 0 || req.GitURL != ""
	if !codecheck && utils.IsMultipart(r) {
		for _, v := range r.MultipartForm.File {
			if len(v) > 0 {
				codecheck = true
				break
			}
		}
	}
	if !codecheck {
		utils.SendError(w, "you must provide either a Git URL or files in the JSON request or in a multipart request", "invalidBody", 400)
		return
	}

	if len(req.FilePaths) != len(req.FileContents) {
		utils.SendError(w, "file_paths and file_contents must have the same length", "invalidBody", 400)
		return
	}

	if req.JoinableForSec < 5 {
		req.JoinableForSec = 300
	}
	if req.MaxConcurrentJoiners <= 0 {
		req.MaxConcurrentJoiners = 250
	}
	if req.TimeoutSec < 5 {
		req.TimeoutSec = 330
	}

	tid, err := utils.UniqueSecureID36()
	if err != nil {
		utils.SendInternalError(w, "CreateToaster:utils.UniqueSecureID60", err)
		return
	}

	toaster := &model.Toaster{
		ID:      "t_" + tid, // the t_ is mostly used to forbid subdomain from taking a toaster id like name
		CodeID:  xid.New().String(),
		OwnerID: userid,

		BuildCmd:             req.BuildCmd,
		ExeCmd:               req.ExeCmd,
		Env:                  req.Env,
		Image:                req.Image,
		JoinableForSec:       req.JoinableForSec,
		MaxConcurrentJoiners: req.MaxConcurrentJoiners,
		TimeoutSec:           req.TimeoutSec,
		Name:                 req.Name,
		Readme:               req.Readme,
		Keywords:             req.Keywords,
		LastModified:         time.Now().Unix(),
		Created:              time.Now().Unix(),
	}

	tmpdir, err := ioutil.TempDir("", "toasterCode_"+toaster.ID)
	if err != nil {
		utils.SendInternalError(w, "CreateToaster:ioutil.TempDir", err)
		return
	}
	defer os.RemoveAll(tmpdir)

	tmpFolder := filepath.Join(tmpdir, "code")

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
	case req.GitURL != "":
		opts := &git.CloneOptions{
			URL:               req.GitURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		}
		if req.GitBranch != "" {
			opts.ReferenceName = plumbing.NewBranchReferenceName(req.GitBranch)
			opts.SingleBranch = true
		}
		if req.GitUsername != "" {
			tmp := &auther.BasicAuth{
				Username: req.GitUsername,
			}

			switch {
			case req.GitAccessToken != "":
				tmp.Password = req.GitAccessToken
			case req.GitUsername != "":
				tmp.Password = req.GitUsername
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

	buildid, buildLogs, err := buildToasterCode(toaster, tarpath)
	if err != nil {
		if err == ErrUnsuccessfulBuild {
			utils.SendSuccess(w, &CreateResponse{
				Success:    false,
				BuildError: buildLogs,
				Toaster:    toaster,
			})
		} else {
			utils.SendInternalError(w, "CreateToaster:buildToasterCode", err)
		}
		return
	}

	err = objectdb.Client.CreateToaster(toaster)
	if err != nil {
		utils.SendInternalError(w, "CreateToaster:objectdb.Client.CreateToaster", err)
		return
	}

	err = objectstorage.Client.UploadFolder(tmpFolder, filepath.Join("clearcode", toaster.ID, toaster.CodeID))
	if err != nil {
		utils.SendInternalError(w, "CreateToaster:objectstorage.Client.UploadFolder", err)
		return
	}

	utils.SendSuccess(w, &CreateResponse{
		Success:   true,
		BuildLogs: buildLogs,
		BuildID:   buildid,
		Toaster:   toaster,
	})
}
