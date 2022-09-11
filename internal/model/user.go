package model

type User struct {
	ID string `json:"id,omitempty" db:"id"`

	Email    string `json:"email,omitempty" db:"email"`
	Password string `json:"-" db:"password"`
}
