package dto

// UserOptions is the actual type for [github.com/kozalosev/goSadTgBot/settings.UserOptions] in this application.
type UserOptions struct {
	Banned bool
	Role   UserRole
}

// UserRole determines the permissions granted to a user.
type UserRole string

const (
	UsualUser UserRole = "user"
	Author    UserRole = "author"
	Admin     UserRole = "admin"
)

// User is an entity for the Users table.
type User struct {
	UID    int64
	Name   string
	Banned bool
	Role   UserRole
}
