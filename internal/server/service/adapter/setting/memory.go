package setting

import (
	"sync"

	"zerosrealm.xyz/tergum/internal/entity"
)

/*
	Cache
*/

type MemoryCache struct {
	mutex    sync.RWMutex
	settings map[string]*entity.Setting
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		mutex:    sync.RWMutex{},
		settings: make(map[string]*entity.Setting),
	}
}

func (s *MemoryCache) Get(id []byte) (*entity.Setting, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	setting, ok := s.settings[string(id)]
	if !ok {
		return nil, nil
	}

	return setting, nil
}

// TODO: Implement pagination.
func (s *MemoryCache) GetAll() ([]*entity.Setting, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	settings := make([]*entity.Setting, 0, len(s.settings))
	for _, setting := range s.settings {
		settings = append(settings, setting)
	}

	return settings, nil
}

func (s *MemoryCache) Add(setting *entity.Setting) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.settings[setting.Key] = setting
	return nil
}

func (s *MemoryCache) Invalidate(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.settings, string(id))
	return nil
}

/*
	Storage
*/

type MemoryStorage struct {
	mutex    sync.RWMutex
	settings map[string]*entity.Setting
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mutex:    sync.RWMutex{},
		settings: make(map[string]*entity.Setting),
	}
}

func (s *MemoryStorage) Get(id []byte) (*entity.Setting, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	setting, ok := s.settings[string(id)]
	if !ok {
		return nil, nil
	}

	return setting, nil
}

// TODO: Implement pagination.
func (s *MemoryStorage) GetAll() ([]*entity.Setting, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	settings := make([]*entity.Setting, 0, len(s.settings))
	for _, setting := range s.settings {
		settings = append(settings, setting)
	}

	return settings, nil
}

func (s *MemoryStorage) Create(setting *entity.Setting) (*entity.Setting, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.settings[setting.Key] = setting

	return setting, nil
}

func (s *MemoryStorage) Update(setting *entity.Setting) (*entity.Setting, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.settings[setting.Key] = setting

	return setting, nil
}

func (s *MemoryStorage) Delete(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.settings, string(id))
	return nil
}
