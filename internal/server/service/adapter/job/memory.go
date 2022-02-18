package job

import (
	"sync"

	"zerosrealm.xyz/tergum/internal/entity"
)

/*
	Cache
*/

type MemoryCache struct {
	mutex sync.RWMutex
	jobs  map[string]*entity.Job
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		mutex: sync.RWMutex{},
		jobs:  make(map[string]*entity.Job),
	}
}

func (s *MemoryCache) Get(id []byte) (*entity.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	job, ok := s.jobs[string(id)]
	if !ok {
		return nil, nil
	}

	return job, nil
}

// TODO: Implement pagination.
func (s *MemoryCache) GetAll() ([]*entity.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	jobs := make([]*entity.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *MemoryCache) Add(job *entity.Job) error {
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
	jobs  map[string]*entity.Job
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mutex: sync.RWMutex{},
		jobs:  make(map[string]*entity.Job),
	}
}

func (s *MemoryStorage) Get(id []byte) (*entity.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	job, ok := s.jobs[string(id)]
	if !ok {
		return nil, nil
	}

	return job, nil
}

// TODO: Implement pagination.
func (s *MemoryStorage) GetAll() ([]*entity.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	jobs := make([]*entity.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *MemoryStorage) Create(job *entity.Job) (*entity.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.jobs[job.ID] = job

	return job, nil
}

func (s *MemoryStorage) Update(job *entity.Job) (*entity.Job, error) {
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
