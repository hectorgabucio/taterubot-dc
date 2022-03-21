package domain

import "os"

type LockedUserRepository interface {
	GetCurrentLock(guildId string) string
	ReleaseUserLock(guildId string)
	SetLock(guildId string, userId string)
}

type FileRepository interface {
	GetFullPath(fileName string) string
	Open(fileName string) (*os.File, error)
	CreateEmpty(fileName string) (*os.File, error)
	DeleteAll(fileNames ...string)
}
