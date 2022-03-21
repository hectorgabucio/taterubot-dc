package inmemory

type LockUser struct {
	id   string
	done chan bool
}

type Repository struct {
	lockedUsers map[string]*LockUser
}

func NewLockedUserRepository() *Repository {
	return &Repository{lockedUsers: map[string]*LockUser{}}
}

func (repo *Repository) SetLock(guildId string, userId string) {
	previousLock := repo.lockedUsers[guildId]
	if previousLock == nil {
		previousLock = &LockUser{
			id:   "",
			done: make(chan bool),
		}
	}
	done := previousLock.done
	if done == nil {
		done = make(chan bool)
	}
	repo.lockedUsers[guildId] = &LockUser{
		id:   userId,
		done: done,
	}

}

func (repo *Repository) GetCurrentLock(guildId string) (string, chan bool) {
	lockUser, ok := repo.lockedUsers[guildId]
	if !ok {
		repo.SetLock(guildId, "")
		lockUser = repo.lockedUsers[guildId]
		return lockUser.id, lockUser.done
	}
	return lockUser.id, lockUser.done
}

func (repo *Repository) ReleaseUserLock(guildId string) {
	repo.lockedUsers[guildId].id = ""
}
