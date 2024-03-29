package repo

import (
	"fmt"
	"sync"

	"zerosrealm.xyz/tergum/internal/entity"
)

/*
	Cache
*/

type MemoryCache struct {
	mutex sync.RWMutex
	repos map[string]*entity.Repo
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		mutex: sync.RWMutex{},
		repos: make(map[string]*entity.Repo),
	}
}

func (s *MemoryCache) Get(id []byte) (*entity.Repo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	repo, ok := s.repos[string(id)]
	if !ok {
		return nil, nil
	}

	return repo, nil
}

// TODO: Implement pagination.
func (s *MemoryCache) GetAll() ([]*entity.Repo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	repos := make([]*entity.Repo, 0, len(s.repos))
	for _, repo := range s.repos {
		repos = append(repos, repo)
	}

	return repos, nil
}

func (s *MemoryCache) Add(repo *entity.Repo) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.repos[fmt.Sprint(repo.ID)] = repo
	return nil
}

func (s *MemoryCache) Invalidate(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.repos, string(id))
	return nil
}

/*
	Storage
*/

type MemoryStorage struct {
	mutex sync.RWMutex
	repos map[string]*entity.Repo
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mutex: sync.RWMutex{},
		repos: make(map[string]*entity.Repo),
	}
}

func (s *MemoryStorage) Get(id []byte) (*entity.Repo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	repo, ok := s.repos[string(id)]
	if !ok {
		return nil, nil
	}

	return repo, nil
}

// TODO: Implement pagination.
func (s *MemoryStorage) GetAll() ([]*entity.Repo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	repos := make([]*entity.Repo, 0, len(s.repos))
	for _, repo := range s.repos {
		repos = append(repos, repo)
	}

	return repos, nil
}

func (s *MemoryStorage) Create(repo *entity.Repo) (*entity.Repo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := len(s.repos) + 1
	repo.ID = id

	s.repos[fmt.Sprint(repo.ID)] = repo

	return repo, nil
}

func (s *MemoryStorage) Update(repo *entity.Repo) (*entity.Repo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.repos[fmt.Sprint(repo.ID)] = repo

	return repo, nil
}

func (s *MemoryStorage) Delete(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.repos, string(id))
	return nil
}
