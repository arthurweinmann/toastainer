package sqldb

import (
	"database/sql"

	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/model"
)

func (c *Client) CreateToaster(toaster *model.Toaster) error {
	_, err := c.db.Exec("INSERT INTO toasters(id, code_id, owner_id, build_command, execution_command, image, environment_variables, joinable_for_seconds, max_concurrent_joiners, timeout_seconds, name, last_modified, created, git_url, git_username, git_branch, git_access_token, git_password, files, readme, keywords) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		toaster.ID,
		toaster.CodeID,
		toaster.OwnerID,
		toaster.BuildCmd,
		toaster.ExeCmd,
		toaster.Image,
		toaster.Env,
		toaster.JoinableForSec,
		toaster.MaxConcurrentJoiners,
		toaster.TimeoutSec,
		toaster.Name,
		toaster.LastModified,
		toaster.Created,
		toaster.GitURL,
		toaster.GitUsername,
		toaster.GitBranch,
		toaster.GitAccessToken,
		toaster.GitPassword,
		toaster.Files,
		toaster.Readme,
		toaster.Keywords,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateToaster(toaster *model.Toaster) error {
	_, err := c.db.Exec("UPDATE toasters SET code_id = ?, build_command = ?, execution_command = ?, image = ?, environment_variables = ?, joinable_for_seconds = ?, max_concurrent_joiners = ?, timeout_seconds = ?, name = ?, last_modified = ?, git_url = ?, git_username = ?, git_branch = ?, git_access_token = ?, git_password = ?, files = ?, readme = ?, keywords = ?, picture_ext = ? WHERE id = ?",
		toaster.CodeID,
		toaster.BuildCmd,
		toaster.ExeCmd,
		toaster.Image,
		toaster.Env,
		toaster.JoinableForSec,
		toaster.MaxConcurrentJoiners,
		toaster.TimeoutSec,
		toaster.Name,
		toaster.LastModified,
		toaster.GitURL,
		toaster.GitUsername,
		toaster.GitBranch,
		toaster.GitAccessToken,
		toaster.GitPassword,
		toaster.Files,
		toaster.Readme,
		toaster.Keywords,
		toaster.PictureExtension,
		toaster.ID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return objectdberror.ErrNotFound
		}
		return err
	}

	return nil
}

func (c *Client) GetUserToaster(userid, toasterid string) (*model.Toaster, error) {
	toasters := []model.Toaster{}
	err := c.db.Select(&toasters, "SELECT * FROM toasters WHERE id = ? AND owner_id = ?", toasterid, userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, objectdberror.ErrNotFound
		}
		return nil, err
	}
	if len(toasters) == 0 {
		return nil, objectdberror.ErrNotFound
	}

	return &toasters[0], nil
}

func (c *Client) ListUsertoasters(userid string) ([]*model.Toaster, error) {
	toasters := []*model.Toaster{}
	err := c.db.Select(&toasters, "SELECT * FROM toasters WHERE owner_id = ?", userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return toasters, nil
		}
		return nil, err
	}

	return toasters, nil
}

func (c *Client) CheckToasterOwnership(userid, toasterid string) (bool, error) {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT * FROM toasters WHERE id = ? AND owner_id = ?", toasterid, userid).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return exists, nil
}

func (c *Client) DelToaster(userid, toasterid string) error {
	_, err := c.db.Exec("DELETE FROM toasters WHERE id = ? AND owner_id = ?", toasterid, userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return objectdberror.ErrNotFound
		}
		return err
	}

	return nil
}
