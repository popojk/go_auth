package domain

import (
	"time"
)

type User struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username" validate:"required"`
	Password  string     `json:"password" validate:"required"`
	Avatar    string     `json:"avatar"`
	UpdatedAt *time.Time `json:"updated_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type PatchUser struct {
	ID        int64     `json:"id"`
	Username  *string   `json:"username"`
	Password  *string   `json:"password"`
	Avatar    *string   `json:"avatar"`
	UpdatedAt time.Time `json:"updated_at"`
}
