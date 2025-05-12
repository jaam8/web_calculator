package ports

type StorageAdapter interface {
	RegisterUser(login string, hashPassword string) (string, error)
	LoginUser(login string) (string, string, error)
}

type CacheAdapter interface {
	SaveToken(token, userID string, refresh bool) error
	GetToken(token string, refresh bool) (string, error)
	DeleteToken(token string, refresh bool) error
}
