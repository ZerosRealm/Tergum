package service

import (
	"fmt"
	"strconv"

	"zerosrealm.xyz/tergum/internal/entities"
)

type RepoCache interface {
	Get(id []byte) (*entities.Repo, error)
	GetAll() ([]*entities.Repo, error)

	Add(repo *entities.Repo) error
	Invalidate(id []byte) error
}

type RepoStorage interface {
	Get(id []byte) (*entities.Repo, error)
	GetAll() ([]*entities.Repo, error)
	Create(repo *entities.Repo) (*entities.Repo, error)
	Update(repo *entities.Repo) (*entities.Repo, error)
	Delete(id []byte) error
}

type RepoService struct {
	cache   RepoCache
	storage RepoStorage
}

func NewRepoService(cache *RepoCache, storage *RepoStorage) *RepoService {
	return &RepoService{
		cache:   *cache,
		storage: *storage,
	}
}

func (svc *RepoService) Get(id []byte) (*entities.Repo, error) {
	if svc.cache != nil {
		repo, err := svc.cache.Get(id)
		if err != nil {
			return nil, fmt.Errorf("repoSvc.Get: could not get repo from cache: %w", err)
		}

		if repo != nil {
			return repo, nil
		}
	}

	repo, err := svc.storage.Get(id)
	if err != nil {
		return nil, fmt.Errorf("repoSvc.Get: could not get repo from storage: %w", err)
	}
	return repo, nil
}

func (svc *RepoService) GetAll() ([]*entities.Repo, error) {
	if svc.cache != nil {
		repos, err := svc.cache.GetAll()
		if err != nil {
			return nil, fmt.Errorf("repoSvc.GetAll: could not get repos from cache: %w", err)
		}

		if len(repos) > 0 {
			return repos, nil
		}
	}

	repos, err := svc.storage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("repoSvc.GetAll: could not get repos from cache: %w", err)
	}
	return repos, nil
}

func (svc *RepoService) Create(repo *entities.Repo) (*entities.Repo, error) {
	repo, err := svc.storage.Create(repo)
	if err != nil {
		return nil, fmt.Errorf("repoSvc.Create: could not create repo: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Add(repo)
		if err != nil {
			return nil, fmt.Errorf("repoSvc.Create: could not add repo to cache: %w", err)
		}
	}

	return repo, nil
}

func (svc *RepoService) Update(repo *entities.Repo) (*entities.Repo, error) {
	repo, err := svc.storage.Update(repo)
	if err != nil {
		return nil, fmt.Errorf("repoSvc.Update: could not update repo: %w", err)
	}

	if svc.cache != nil {
		id := strconv.Itoa(repo.ID)
		err = svc.cache.Invalidate([]byte(id))
		if err != nil {
			return nil, fmt.Errorf("repoSvc.Create: could not invalidate repo in cache: %w", err)
		}
	}

	return repo, nil
}

func (svc *RepoService) Delete(id []byte) error {
	err := svc.storage.Delete(id)
	if err != nil {
		return fmt.Errorf("repoSvc.Delete: could not delete repo: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Invalidate(id)
		if err != nil {
			return fmt.Errorf("repoSvc.Delete: could not invalidate repo in cache: %w", err)
		}
	}
	return nil
}
