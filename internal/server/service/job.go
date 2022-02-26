package service

import (
	"fmt"

	"zerosrealm.xyz/tergum/internal/entity"
)

type JobCache interface {
	Get(id []byte) (*entity.Job, error)
	GetAll() ([]*entity.Job, error)

	Add(job *entity.Job) error
	Invalidate(id []byte) error
}

type JobStorage interface {
	Get(id []byte) (*entity.Job, error)
	GetAll() ([]*entity.Job, error)
	Create(job *entity.Job) (*entity.Job, error)
	Update(job *entity.Job) (*entity.Job, error)
	Delete(id []byte) error
}

type JobService struct {
	cache   JobCache
	storage JobStorage
}

func NewJobService(cache *JobCache, storage *JobStorage) *JobService {
	return &JobService{
		cache:   *cache,
		storage: *storage,
	}
}

func (svc *JobService) Get(id []byte) (*entity.Job, error) {
	if svc.cache != nil {
		job, err := svc.cache.Get(id)
		if err != nil {
			return nil, fmt.Errorf("jobSvc.Get: could not get job from cache: %w", err)
		}

		if job != nil {
			return job, nil
		}
	}

	job, err := svc.storage.Get(id)
	if err != nil {
		return nil, fmt.Errorf("jobSvc.Get: could not get job from storage: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Add(job)
		if err != nil {
			return nil, fmt.Errorf("jobSvc.Get: could not add job to cache: %w", err)
		}
	}

	return job, nil
}

func (svc *JobService) GetAll() ([]*entity.Job, error) {
	// if svc.cache != nil {
	// 	jobs, err := svc.cache.GetAll()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("jobSvc.GetAll: could not get jobs from cache: %w", err)
	// 	}

	// 	if len(jobs) > 0 {
	// 		return jobs, nil
	// 	}
	// }

	jobs, err := svc.storage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("jobSvc.GetAll: could not get jobs from storage: %w", err)
	}

	// if svc.cache != nil {
	// 	for _, job := range jobs {
	// 		err = svc.cache.Add(job)
	// 		if err != nil {
	// 			return nil, fmt.Errorf("jobSvc.GetAll: could not add job to cache: %w", err)
	// 		}
	// 	}
	// }
	return jobs, nil
}

func (svc *JobService) Create(job *entity.Job) (*entity.Job, error) {
	job, err := svc.storage.Create(job)
	if err != nil {
		return nil, fmt.Errorf("jobSvc.Create: could not create job: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Add(job)
		if err != nil {
			return nil, fmt.Errorf("jobSvc.Create: could not add job to cache: %w", err)
		}
	}

	return job, nil
}

func (svc *JobService) Update(job *entity.Job) (*entity.Job, error) {
	job, err := svc.storage.Update(job)
	if err != nil {
		return nil, fmt.Errorf("jobSvc.Update: could not update job: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Invalidate([]byte(job.ID))
		if err != nil {
			return nil, fmt.Errorf("jobSvc.Create: could not invalidate job in cache: %w", err)
		}
	}

	return job, nil
}

func (svc *JobService) Delete(id []byte) error {
	err := svc.storage.Delete(id)
	if err != nil {
		return fmt.Errorf("jobSvc.Delete: could not delete job: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Invalidate(id)
		if err != nil {
			return fmt.Errorf("jobSvc.Delete: could not invalidate job in cache: %w", err)
		}
	}
	return nil
}
