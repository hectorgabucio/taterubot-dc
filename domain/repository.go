package domain

import "os"

type LockedUserRepository interface {
	GetCurrentLock(guildID string) (string, chan bool)
	ReleaseUserLock(guildID string)
	SetLock(guildID string, userID string)
}

type FileRepository interface {
	GetFullPath(fileName string) string
	Open(fileName string) (*os.File, error)
	CreateEmpty(fileName string) (*os.File, error)
	DeleteAll(fileNames ...string)
}
