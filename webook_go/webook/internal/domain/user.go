package domain

import (
	"time"
)

type User struct {
	Id       int64
	Email    string
	Password string
	Nickname string
	Birthday string
	AboutMe  string
	Phone    string
	Ctime    time.Time
}
