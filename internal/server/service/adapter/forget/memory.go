package forget

import (
	"fmt"
	"sync"

	"zerosrealm.xyz/tergum/internal/entity"
)

/*
	Cache
*/

type MemoryCache struct {
	mutex   sync.RWMutex
	forgets map[string]*entity.Forget
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		mutex:   sync.RWMutex{},
		forgets: make(map[string]*entity.Forget),
	}
}

func (s *MemoryCache) Get(id []byte) (*entity.Forget, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	forget, ok := s.forgets[string(id)]
	if !ok {
		return nil, nil
	}

	return forget, nil
}

// TODO: Implement pagination.
func (s *MemoryCache) GetAll() ([]*entity.Forget, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	forgets := make([]*entity.Forget, 0, len(s.forgets))
	for _, forget := range s.forgets {
		forgets = append(forgets, forget)
	}

	return forgets, nil
}

func (s *MemoryCache) Add(forget *entity.Forget) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.forgets[fmt.Sprint(forget.ID)] = forget
	return nil
}

func (s *MemoryCache) Invalidate(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.forgets, string(id))
	return nil
}

/*
	Storage
*/

type MemoryStorage struct {
	mutex   sync.RWMutex
	forgets map[string]*entity.Forget
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mutex:   sync.RWMutex{},
		forgets: make(map[string]*entity.Forget),
	}
}

func (s *MemoryStorage) Get(id []byte) (*entity.Forget, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	forget, ok := s.forgets[string(id)]
	if !ok {
		return nil, nil
	}

	return forget, nil
}

// TODO: Implement pagination.
func (s *MemoryStorage) GetAll() ([]*entity.Forget, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	forgets := make([]*entity.Forget, 0, len(s.forgets))
	for _, forget := range s.forgets {
		forgets = append(forgets, forget)
	}

	return forgets, nil
}

func (s *MemoryStorage) Create(forget *entity.Forget) (*entity.Forget, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := len(s.forgets) + 1
	forget.ID = id

	s.forgets[fmt.Sprint(forget.ID)] = forget

	return forget, nil
}

func (s *MemoryStorage) Update(forget *entity.Forget) (*entity.Forget, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.forgets[fmt.Sprint(forget.ID)] = forget

	return forget, nil
}

func (s *MemoryStorage) Delete(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.forgets, string(id))
	return nil
}
