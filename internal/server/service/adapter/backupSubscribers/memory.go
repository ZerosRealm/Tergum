package backupSubscribers

import (
	"fmt"
	"sync"

	"zerosrealm.xyz/tergum/internal/entity"
)

/*
	Cache
*/

type MemoryCache struct {
	mutex             sync.RWMutex
	backupSubscribers map[string]*entity.BackupSubscribers
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		mutex:             sync.RWMutex{},
		backupSubscribers: make(map[string]*entity.BackupSubscribers),
	}
}

func (s *MemoryCache) Get(id []byte) (*entity.BackupSubscribers, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	subscribers, ok := s.backupSubscribers[string(id)]
	if !ok {
		return nil, nil
	}

	return subscribers, nil
}

// TODO: Implement pagination.
func (s *MemoryCache) GetAll() ([]*entity.BackupSubscribers, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	backupSubscribers := make([]*entity.BackupSubscribers, 0, len(s.backupSubscribers))
	for _, backup := range s.backupSubscribers {
		backupSubscribers = append(backupSubscribers, backup)
	}

	return backupSubscribers, nil
}

func (s *MemoryCache) Add(backupSubscribers *entity.BackupSubscribers) error {
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
	backupSubscribers map[string]*entity.BackupSubscribers
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mutex:             sync.RWMutex{},
		backupSubscribers: make(map[string]*entity.BackupSubscribers),
	}
}

func (s *MemoryStorage) Get(id []byte) (*entity.BackupSubscribers, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	subscribers, ok := s.backupSubscribers[string(id)]
	if !ok {
		return nil, nil
	}

	return subscribers, nil
}

// TODO: Implement pagination.
func (s *MemoryStorage) GetAll() ([]*entity.BackupSubscribers, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	backupSubscribers := make([]*entity.BackupSubscribers, 0, len(s.backupSubscribers))
	for _, backup := range s.backupSubscribers {
		backupSubscribers = append(backupSubscribers, backup)
	}

	return backupSubscribers, nil
}

func (s *MemoryStorage) Create(backupSubscribers *entity.BackupSubscribers) (*entity.BackupSubscribers, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := len(s.backupSubscribers) + 1
	backupSubscribers.BackupID = id

	s.backupSubscribers[fmt.Sprint(backupSubscribers.BackupID)] = backupSubscribers

	return backupSubscribers, nil
}

func (s *MemoryStorage) Update(backupSubscribers *entity.BackupSubscribers) (*entity.BackupSubscribers, error) {
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
