package model

type SubDomain struct {
	ID        string `json:"id,omitempty" db:"id"`
	Name      string `json:"name,omitempty" db:"name"`
	UserID    string `json:"user_id,omitempty" db:"user_id"`
	ToasterID string `json:"toaster_id,omitempty" db:"toaster_id"`
}
