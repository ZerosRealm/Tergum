package backup

import (
	"fmt"
	"sync"

	"zerosrealm.xyz/tergum/internal/entities"
)

/*
	Cache
*/

type MemoryCache struct {
	mutex   sync.RWMutex
	backups map[string]*entities.Backup
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		mutex:   sync.RWMutex{},
		backups: make(map[string]*entities.Backup),
	}
}

func (s *MemoryCache) Get(id []byte) (*entities.Backup, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	backup, ok := s.backups[string(id)]
	if !ok {
		return nil, nil
	}

	return backup, nil
}

// TODO: Implement pagination.
func (s *MemoryCache) GetAll() ([]*entities.Backup, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	backups := make([]*entities.Backup, 0, len(s.backups))
	for _, backup := range s.backups {
		backups = append(backups, backup)
	}

	return backups, nil
}

func (s *MemoryCache) Add(backup *entities.Backup) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.backups[fmt.Sprint(backup.ID)] = backup
	return nil
}

func (s *MemoryCache) Invalidate(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.backups, string(id))
	return nil
}

/*
	Storage
*/

type MemoryStorage struct {
	mutex   sync.RWMutex
	backups map[string]*entities.Backup
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mutex:   sync.RWMutex{},
		backups: make(map[string]*entities.Backup),
	}
}

func (s *MemoryStorage) Get(id []byte) (*entities.Backup, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	backup, ok := s.backups[string(id)]
	if !ok {
		return nil, nil
	}

	return backup, nil
}

// TODO: Implement pagination.
func (s *MemoryStorage) GetAll() ([]*entities.Backup, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	backups := make([]*entities.Backup, 0, len(s.backups))
	for _, backup := range s.backups {
		backups = append(backups, backup)
	}

	return backups, nil
}

func (s *MemoryStorage) Create(backup *entities.Backup) (*entities.Backup, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := len(s.backups) + 1
	backup.ID = id

	s.backups[fmt.Sprint(backup.ID)] = backup

	return backup, nil
}

func (s *MemoryStorage) Update(backup *entities.Backup) (*entities.Backup, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.backups[fmt.Sprint(backup.ID)] = backup

	return backup, nil
}

func (s *MemoryStorage) Delete(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.backups, string(id))
	return nil
}
