package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // never return password in JSON
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
