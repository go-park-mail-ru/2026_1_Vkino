package db

import (
	"errors"
	"maps"
	"sync"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("key already exists")
	ErrTableNotFound = errors.New("table not found")
)

type DB struct {
	mu   sync.RWMutex
	tbls map[string]map[string][]byte
}

type Named interface {
	Name() string
}

func NewDB(models []Named) *DB {
	db := &DB{tbls: make(map[string]map[string][]byte)}

	for _, model := range models {
		tableName := model.Name()
		db.tbls[tableName] = make(map[string][]byte)
	}

	return db
}

func (db *DB) Save(tableName string, key string, data []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	table, exists := db.tbls[tableName]
	if !exists {
		return ErrTableNotFound
	}

	if _, exists := table[key]; exists {
		return ErrAlreadyExists
	}

	table[key] = data
	return nil
}

func (db *DB) Update(tableName string, key string, data []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	table, exists := db.tbls[tableName]
	if !exists {
		return ErrTableNotFound
	}

	_, exists = table[key]
	if !exists {
		return ErrNotFound
	}

	table[key] = data
	return nil
}

func (db *DB) Get(tableName string, key string) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	table, exists := db.tbls[tableName]
	if !exists {
		return nil, ErrTableNotFound
	}

	data, exists := table[key]
	if !exists {
		return nil, ErrNotFound
	}

	return data, nil
}

func (db *DB) Delete(tableName string, key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	table, exists := db.tbls[tableName]
	if !exists {
		return ErrTableNotFound
	}

	if _, exists := table[key]; !exists {
		return ErrNotFound
	}

	delete(table, key)
	return nil
}

func (db *DB) GetAll(tableName string) (map[string][]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	table, exists := db.tbls[tableName]
	if !exists {
		return nil, ErrTableNotFound
	}

	result := make(map[string][]byte)
	maps.Copy(result, table)
	return result, nil
}

func (db *DB) TableExists(tableName string) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()

	_, exists := db.tbls[tableName]
	return exists
}
