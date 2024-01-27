package models

import "time"

type RegSession struct {
	FingerPrint string `json:"finger_print"`
	IsConfirmed bool   `json:"is_confirmed"`
	UserID      string `json:"user_id"`
}

type Session struct {
	UserID      string    `json:"user_id" db:"user_id"`
	FingerPrint string    `json:"finger_print" db:"finger_print"`
	ExpiredAt   time.Time `json:"expired_at" db:"exp_at"`
}
