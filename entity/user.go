package entity

// import "database/sql"

type User struct {
	ID_user int     `json:"id_user"`
	Email   string  `json:"email"`
	Name    string  `json:"name"`
	Phone   *string `json:"phone"`
	Picture *string `json:"profile_picture"`
}
