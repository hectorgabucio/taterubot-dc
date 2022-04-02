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

func (repo *Repository) SetLock(guildID string, userID string) {
	previousLock := repo.lockedUsers[guildID]
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
	repo.lockedUsers[guildID] = &LockUser{
		id:   userID,
		done: done,
	}
}
func (repo *Repository) GetCurrentLock(guildID string) (string, chan bool) {
	lockUser, ok := repo.lockedUsers[guildID]
	if !ok {
		repo.SetLock(guildID, "")
		lockUser = repo.lockedUsers[guildID]

		return lockUser.id, lockUser.done
	}

	return lockUser.id, lockUser.done
}

func (repo *Repository) ReleaseUserLock(guildID string) {
	repo.lockedUsers[guildID].id = ""
}
