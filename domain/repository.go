package domain

type LockedUserRepository interface {
	GetCurrentLock() string
	ReleaseUserLock()
	SetLock(id string)
}
