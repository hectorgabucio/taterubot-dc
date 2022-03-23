package localfs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Repository struct {
	basePath string
}

func NewRepository(basePath string) *Repository {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		err := os.Mkdir(basePath, 0750)
		if err != nil {
			log.Fatalln("could not create root dir for local fs", err)
		}
	}
	return &Repository{basePath: basePath}
}

func (repo *Repository) GetFullPath(fileName string) string {
	baseFilePath := repo.basePath
	return fmt.Sprintf("%s/%s", baseFilePath, fileName)
}

func (repo *Repository) Open(fileName string) (*os.File, error) {
	fileLocation := repo.sanitizePath(fileName)
	return os.Open(fileLocation)
}

func (repo *Repository) CreateEmpty(fileName string) (*os.File, error) {
	fileLocation := repo.sanitizePath(fileName)
	return os.Create(fileLocation)
}
func (repo *Repository) DeleteAll(fileNames ...string) {
	for _, fileName := range fileNames {
		fileLocation := repo.sanitizePath(fileName)
		_ = os.Remove(fileLocation)
	}
}

func (repo *Repository) sanitizePath(fileName string) string {
	fileLocation := fileName
	if !strings.HasPrefix(fileLocation, repo.basePath) {
		fileLocation = repo.GetFullPath(fileLocation)
	}
	return filepath.Clean(fileLocation)
}
