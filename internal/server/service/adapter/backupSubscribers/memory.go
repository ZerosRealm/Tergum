package backupSubscribers

import (
	"fmt"
	"sync"

	"zerosrealm.xyz/tergum/internal/entities"
)

/*
	Cache
*/

type MemoryCache struct {
	mutex             sync.RWMutex
	backupSubscribers map[string]*entities.BackupSubscribers
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		mutex:             sync.RWMutex{},
		backupSubscribers: make(map[string]*entities.BackupSubscribers),
	}
}

func (s *MemoryCache) Get(id []byte) (*entities.BackupSubscribers, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	subscribers, ok := s.backupSubscribers[string(id)]
	if !ok {
		return nil, nil
	}

	return subscribers, nil
}

// TODO: Implement pagination.
func (s *MemoryCache) GetAll() ([]*entities.BackupSubscribers, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	backupSubscribers := make([]*entities.BackupSubscribers, 0, len(s.backupSubscribers))
	for _, backup := range s.backupSubscribers {
		backupSubscribers = append(backupSubscribers, backup)
	}

	return backupSubscribers, nil
}

func (s *MemoryCache) Add(backupSubscribers *entities.BackupSubscribers) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.backupSubscribers[fmt.Sprint(backupSubscribers.BackupID)] = backupSubscribers
	return nil
}

func (s *MemoryCache) Invalidate(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.backupSubscribers, string(id))
	return nil
}

/*
	Storage
*/

type MemoryStorage struct {
	mutex             sync.RWMutex
	backupSubscribers map[string]*entities.BackupSubscribers
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mutex:             sync.RWMutex{},
		backupSubscribers: make(map[string]*entities.BackupSubscribers),
	}
}

func (s *MemoryStorage) Get(id []byte) (*entities.BackupSubscribers, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	subscribers, ok := s.backupSubscribers[string(id)]
	if !ok {
		return nil, nil
	}

	return subscribers, nil
}

// TODO: Implement pagination.
func (s *MemoryStorage) GetAll() ([]*entities.BackupSubscribers, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	backupSubscribers := make([]*entities.BackupSubscribers, 0, len(s.backupSubscribers))
	for _, backup := range s.backupSubscribers {
		backupSubscribers = append(backupSubscribers, backup)
	}

	return backupSubscribers, nil
}

func (s *MemoryStorage) Create(backupSubscribers *entities.BackupSubscribers) (*entities.BackupSubscribers, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := len(s.backupSubscribers) + 1
	backupSubscribers.BackupID = id

	s.backupSubscribers[fmt.Sprint(backupSubscribers.BackupID)] = backupSubscribers

	return backupSubscribers, nil
}

func (s *MemoryStorage) Update(backupSubscribers *entities.BackupSubscribers) (*entities.BackupSubscribers, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.backupSubscribers[fmt.Sprint(backupSubscribers.BackupID)] = backupSubscribers

	return backupSubscribers, nil
}

func (s *MemoryStorage) Delete(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.backupSubscribers, string(id))
	return nil
}
