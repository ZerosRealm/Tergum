package job

import (
	"sync"

	"zerosrealm.xyz/tergum/internal/entities"
)

/*
	Cache
*/

type MemoryCache struct {
	mutex sync.RWMutex
	jobs  map[string]*entities.Job
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		mutex: sync.RWMutex{},
		jobs:  make(map[string]*entities.Job),
	}
}

func (s *MemoryCache) Get(id []byte) (*entities.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	job, ok := s.jobs[string(id)]
	if !ok {
		return nil, nil
	}

	return job, nil
}

// TODO: Implement pagination.
func (s *MemoryCache) GetAll() ([]*entities.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	jobs := make([]*entities.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *MemoryCache) Add(job *entities.Job) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.jobs[job.ID] = job
	return nil
}

func (s *MemoryCache) Invalidate(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.jobs, string(id))
	return nil
}

/*
	Storage
*/

type MemoryStorage struct {
	mutex sync.RWMutex
	jobs  map[string]*entities.Job
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mutex: sync.RWMutex{},
		jobs:  make(map[string]*entities.Job),
	}
}

func (s *MemoryStorage) Get(id []byte) (*entities.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	job, ok := s.jobs[string(id)]
	if !ok {
		return nil, nil
	}

	return job, nil
}

// TODO: Implement pagination.
func (s *MemoryStorage) GetAll() ([]*entities.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	jobs := make([]*entities.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *MemoryStorage) Create(job *entities.Job) (*entities.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.jobs[job.ID] = job

	return job, nil
}

func (s *MemoryStorage) Update(job *entities.Job) (*entities.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.jobs[job.ID] = job

	return job, nil
}

func (s *MemoryStorage) Delete(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.jobs, string(id))
	return nil
}
