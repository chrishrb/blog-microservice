package store

import "time"

type User struct {
	ID        int
	Email     string
	FirstName string
	LastName  string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserStore interface {
	CreateUser(user *User) (*User, error)
	UpdateUser(user *User) (*User, error)
	LookupUser(ID int) (*User, error)
	ListUsers() ([]*User, error)
	DeleteUser(ID int) error
}
