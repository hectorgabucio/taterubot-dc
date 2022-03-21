package inmemory

type Repository struct {
	lockedUsers map[string]string
}

func NewLockedUserRepository() *Repository {
	return &Repository{lockedUsers: map[string]string{}}
}

func (repo *Repository) SetLock(guildId string, userId string) {
	repo.lockedUsers[guildId] = userId
}

func (repo *Repository) GetCurrentLock(guildId string) string {
	lockUserId, ok := repo.lockedUsers[guildId]
	if !ok {
		return ""
	}
	return lockUserId
}

func (repo *Repository) ReleaseUserLock(guildId string) {
	repo.lockedUsers[guildId] = ""
}
