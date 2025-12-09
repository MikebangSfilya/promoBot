package config

// UserOptions is the actual type for [github.com/kozalosev/goSadTgBot/settings.UserOptions] in this application.
type UserOptions struct {
	Role UserRole
}

// UserRole determines the permissions granted to a user.
type (
	UserId           int64
	UserName         string
	UserLanguageCode string
	UserRole         string
)

const (
	UsualUser UserRole = "user"
	Author    UserRole = "author"
	Admin     UserRole = "admin"
)

// User is an entity for the Users table.
type User struct {
	UID      UserId           `yaml:"uid"`
	Name     UserName         `yaml:"name"`
	Language UserLanguageCode `yaml:"language"`
	Role     UserRole         `yaml:"role"`
}
