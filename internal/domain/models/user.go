package models

type UserDTO struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type User struct {
	Id        string `db:"id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Sex       string `db:"sex"`
	TgUserID  string `db:"tg_user_id"`
}