package model

type Toaster struct {
	ID     string `json:"id,omitempty" db:"id"`
	Cursor int    `json:"cursor,omitempty" db:"cursor"`

	CodeID string `json:"code_id,omitempty" db:"code_id"`

	OwnerID string `json:"owner_id,omitempty" db:"owner_id"`

	BuildCmd ArrayString `json:"build_command,omitempty" db:"build_command"`
	ExeCmd   ArrayString `json:"execution_command,omitempty" db:"execution_command"`
	Env      ArrayString `json:"environment_variables,omitempty" db:"environment_variables"`
	Image    string      `json:"image,omitempty" db:"image"`

	JoinableForSec       int `json:"joinable_for_seconds,omitempty" db:"joinable_for_seconds"`
	MaxConcurrentJoiners int `json:"max_concurrent_joiners,omitempty" db:"max_concurrent_joiners"`
	TimeoutSec           int `json:"timeout_seconds,omitempty" db:"timeout_seconds"`

	Name string `json:"name,omitempty" db:"name"` // max 120 characters

	LastModified int64 `json:"last_modified,omitempty" db:"last_modified"`
	Created      int64 `json:"created,omitempty" db:"created"`

	GitURL         string `json:"git_url,omitempty" db:"git_url"`
	GitUsername    string `json:"git_username,omitempty" db:"git_username"`
	GitBranch      string `json:"git_branch,omitempty" db:"git_branch"`
	GitAccessToken string `json:"-" db:"git_access_token"`
	GitPassword    string `json:"-" db:"git_password"`

	Files ArrayString `json:"files,omitempty" db:"files"`

	Readme   string      `json:"readme,omitempty" db:"readme"`
	Keywords ArrayString `json:"keywords,omitempty" db:"keywords"`
}
