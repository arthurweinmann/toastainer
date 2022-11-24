package model

type User struct {
	ID     string `json:"id,omitempty" db:"id"`
	Cursor int    `json:"-" db:"cursor"`

	Username string `json:"username,omitempty" db:"username"`
	Email    string `json:"email,omitempty" db:"email"`
	Password string `json:"-" db:"password"`

	PictureExtension string `json:"-" db:"picture_ext"`
	PictureLink      string `json:"picture_link" db:"-"`
}
