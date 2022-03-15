package inmemory

type Repository struct {
	lockedUser string
}

func New() *Repository {
	return &Repository{lockedUser: ""}
}

func (repo *Repository) SetLock(id string) {
	repo.lockedUser = id
}

func (repo *Repository) GetCurrentLock() string {
	return repo.lockedUser
}

func (repo *Repository) ReleaseUserLock() {
	repo.lockedUser = ""
}
