package models

type RegSession struct {
	FingerPrint string `json:"finger_print"`
	IsConfirmed bool   `json:"is_confirmed"`
	UserID      string `json:"user_id"`
}
