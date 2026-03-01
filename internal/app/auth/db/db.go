package db

import (
	"sync"
)

type Repository struct {
	mu   sync.RWMutex
	tbls map[string]map[string][]byte
}

type Named interface {
	Name() string
}

func InitRepo(models []Named) *Repository {
	repo := &Repository{tbls: make(map[string]map[string][]byte)}

	for _, model := range models {
		tableName := model.Name()
		repo.tbls[tableName] = make(map[string][]byte)
	}

	return repo
}
