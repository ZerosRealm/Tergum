package service

import (
	"fmt"
	"strconv"

	"zerosrealm.xyz/tergum/internal/entity"
)

type ForgetCache interface {
	Get(id []byte) (*entity.Forget, error)
	GetAll() ([]*entity.Forget, error)

	Add(forget *entity.Forget) error
	Invalidate(id []byte) error
}

type ForgetStorage interface {
	Get(id []byte) (*entity.Forget, error)
	GetAll() ([]*entity.Forget, error)
	Create(forget *entity.Forget) (*entity.Forget, error)
	Update(forget *entity.Forget) (*entity.Forget, error)
	Delete(id []byte) error
}

type ForgetService struct {
	cache   ForgetCache
	storage ForgetStorage
}

func NewForgetService(cache *ForgetCache, storage *ForgetStorage) *ForgetService {
	return &ForgetService{
		cache:   *cache,
		storage: *storage,
	}
}

func (svc *ForgetService) Get(id []byte) (*entity.Forget, error) {
	if svc.cache != nil {
		forget, err := svc.cache.Get(id)
		if err != nil {
			return nil, fmt.Errorf("forgetSvc.Get: could not get forget from cache: %w", err)
		}

		if forget != nil {
			return forget, nil
		}
	}

	forget, err := svc.storage.Get(id)
	if err != nil {
		return nil, fmt.Errorf("forgetSvc.Get: could not get forget from storage: %w", err)
	}
	return forget, nil
}

func (svc *ForgetService) GetAll() ([]*entity.Forget, error) {
	if svc.cache != nil {
		forgets, err := svc.cache.GetAll()
		if err != nil {
			return nil, fmt.Errorf("forgetSvc.GetAll: could not get forgets from cache: %w", err)
		}

		if len(forgets) > 0 {
			return forgets, nil
		}
	}

	forgets, err := svc.storage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("forgetSvc.GetAll: could not get forgets from cache: %w", err)
	}
	return forgets, nil
}

func (svc *ForgetService) Create(forget *entity.Forget) (*entity.Forget, error) {
	forget, err := svc.storage.Create(forget)
	if err != nil {
		return nil, fmt.Errorf("forgetSvc.Create: could not create forget: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Add(forget)
		if err != nil {
			return nil, fmt.Errorf("forgetSvc.Create: could not add forget to cache: %w", err)
		}
	}

	return forget, nil
}

func (svc *ForgetService) Update(forget *entity.Forget) (*entity.Forget, error) {
	forget, err := svc.storage.Update(forget)
	if err != nil {
		return nil, fmt.Errorf("forgetSvc.Update: could not update forget: %w", err)
	}

	if svc.cache != nil {
		id := strconv.Itoa(forget.ID)
		err = svc.cache.Invalidate([]byte(id))
		if err != nil {
			return nil, fmt.Errorf("forgetSvc.Create: could not invalidate forget in cache: %w", err)
		}
	}

	return forget, nil
}

func (svc *ForgetService) Delete(id []byte) error {
	err := svc.storage.Delete(id)
	if err != nil {
		return fmt.Errorf("forgetSvc.Delete: could not delete forget: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Invalidate(id)
		if err != nil {
			return fmt.Errorf("forgetSvc.Delete: could not invalidate forget in cache: %w", err)
		}
	}
	return nil
}
