package domain

type ContextKey string

const (
	UserIDKey   ContextKey = "user_id"
	RoleKey     ContextKey = "role"
	UsernameKey ContextKey = "username"
	ModeKey     ContextKey = "search_mode"
)