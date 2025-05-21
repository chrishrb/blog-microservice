package store

type Engine interface {
	UserStore
	JWTBlacklistStore
}
