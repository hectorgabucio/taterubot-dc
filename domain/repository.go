package domain

import (
	"os"
	"time"
)

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

type VoiceData struct {
	GuildID   string
	ID        string
	Timestamp time.Time
	Name      string
	UserID    string
	Duration  int
}

// TODO save timestamp, voice audio name, user id, duration,
type VoiceDataRepository interface {
	Save(data VoiceData) error
	GetOnRange(guildID string, from time.Time, to time.Time) ([]VoiceData, error)
}
