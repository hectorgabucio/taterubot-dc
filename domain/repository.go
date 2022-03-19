package domain

import "os"

type LockedUserRepository interface {
	GetCurrentLock() string
	ReleaseUserLock()
	SetLock(id string)
}

type FileRepository interface {
	GetFullPath(fileName string) string
	Open(fileName string) (*os.File, error)
	CreateEmpty(fileName string) (*os.File, error)
	DeleteAll(fileNames ...string)
}
