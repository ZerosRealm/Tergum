package setting

import (
	"sync"

	"zerosrealm.xyz/tergum/internal/entities"
)

/*
	Cache
*/

type MemoryCache struct {
	mutex    sync.RWMutex
	settings map[string]*entities.Setting
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		mutex:    sync.RWMutex{},
		settings: make(map[string]*entities.Setting),
	}
}

func (s *MemoryCache) Get(id []byte) (*entities.Setting, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	setting, ok := s.settings[string(id)]
	if !ok {
		return nil, nil
	}

	return setting, nil
}

// TODO: Implement pagination.
func (s *MemoryCache) GetAll() ([]*entities.Setting, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	settings := make([]*entities.Setting, 0, len(s.settings))
	for _, setting := range s.settings {
		settings = append(settings, setting)
	}

	return settings, nil
}

func (s *MemoryCache) Add(setting *entities.Setting) error {
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
	settings map[string]*entities.Setting
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mutex:    sync.RWMutex{},
		settings: make(map[string]*entities.Setting),
	}
}

func (s *MemoryStorage) Get(id []byte) (*entities.Setting, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	setting, ok := s.settings[string(id)]
	if !ok {
		return nil, nil
	}

	return setting, nil
}

// TODO: Implement pagination.
func (s *MemoryStorage) GetAll() ([]*entities.Setting, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	settings := make([]*entities.Setting, 0, len(s.settings))
	for _, setting := range s.settings {
		settings = append(settings, setting)
	}

	return settings, nil
}

func (s *MemoryStorage) Create(setting *entities.Setting) (*entities.Setting, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.settings[setting.Key] = setting

	return setting, nil
}

func (s *MemoryStorage) Update(setting *entities.Setting) (*entities.Setting, error) {
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
