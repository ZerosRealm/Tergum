package service

import (
	"fmt"
	"strconv"

	"zerosrealm.xyz/tergum/internal/entity"
)

type BackupSubscriberCache interface {
	Get(backupID []byte) (*entity.BackupSubscribers, error)
	GetAll() ([]*entity.BackupSubscribers, error)

	Add(backupSubscriber *entity.BackupSubscribers) error
	Invalidate(backupID []byte) error
}

type BackupSubscriberStorage interface {
	Get(backupID []byte) (*entity.BackupSubscribers, error)
	GetAll() ([]*entity.BackupSubscribers, error)
	Create(backupSubscriber *entity.BackupSubscribers) (*entity.BackupSubscribers, error)
	Update(backupSubscriber *entity.BackupSubscribers) (*entity.BackupSubscribers, error)
	Delete(backupID []byte) error
}

type BackupSubscriberService struct {
	cache   BackupSubscriberCache
	storage BackupSubscriberStorage
}

func NewBackupSubscriberService(cache *BackupSubscriberCache, storage *BackupSubscriberStorage) *BackupSubscriberService {
	return &BackupSubscriberService{
		cache:   *cache,
		storage: *storage,
	}
}

func (svc *BackupSubscriberService) Get(backupID []byte) (*entity.BackupSubscribers, error) {
	if svc.cache != nil {
		backupSubscribers, err := svc.cache.Get(backupID)
		if err != nil {
			return nil, fmt.Errorf("backupSubscriberSvc.Get: could not get backupSubscriber from cache: %w", err)
		}

		if backupSubscribers != nil {
			return backupSubscribers, nil
		}
	}

	backupSubscribers, err := svc.storage.Get(backupID)
	if err != nil {
		return nil, fmt.Errorf("backupSubscriberSvc.Get: could not get backupSubscriber from storage: %w", err)
	}
	return backupSubscribers, nil
}

func (svc *BackupSubscriberService) GetAll() ([]*entity.BackupSubscribers, error) {
	if svc.cache != nil {
		backupSubscribers, err := svc.cache.GetAll()
		if err != nil {
			return nil, fmt.Errorf("backupSubscriberSvc.GetAll: could not get backupSubscribers from cache: %w", err)
		}

		if len(backupSubscribers) > 0 {
			return backupSubscribers, nil
		}
	}

	backupSubscribers, err := svc.storage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("backupSubscriberSvc.GetAll: could not get backupSubscribers from storage: %w", err)
	}
	return backupSubscribers, nil
}

func (svc *BackupSubscriberService) Create(backupSubscriber *entity.BackupSubscribers) (*entity.BackupSubscribers, error) {
	backupSubscriber, err := svc.storage.Create(backupSubscriber)
	if err != nil {
		return nil, fmt.Errorf("backupSubscriberSvc.Create: could not create backupSubscriber: %w", err)
	}

	if svc.cache != nil {
		err = (svc.cache).Add(backupSubscriber)
		if err != nil {
			return nil, fmt.Errorf("backupSubscriberSvc.Create: could not add backupSubscriber to cache: %w", err)
		}
	}

	return backupSubscriber, nil
}

func (svc *BackupSubscriberService) Update(backupSubscriber *entity.BackupSubscribers) (*entity.BackupSubscribers, error) {
	backup, err := svc.storage.Update(backupSubscriber)
	if err != nil {
		return nil, fmt.Errorf("backupSubscriberSvc.Update: could not update backupSubscriber: %w", err)
	}

	if svc.cache != nil {
		id := strconv.Itoa(backupSubscriber.BackupID)
		err = (svc.cache).Invalidate([]byte(id))
		if err != nil {
			return nil, fmt.Errorf("backupSubscriberSvc.Update: could not invalidate backupSubscriber in cache: %w", err)
		}
	}

	return backup, nil
}

func (svc *BackupSubscriberService) Delete(backupID []byte, agentID []byte) error {
	err := svc.storage.Delete(backupID)
	if err != nil {
		return fmt.Errorf("backupSubscriberSvc.Delete: could not delete backupSubscriber: %w", err)
	}

	if svc.cache != nil {
		err = (svc.cache).Invalidate(backupID)
		if err != nil {
			return fmt.Errorf("backupSubscriberSvc.Delete: could not invalidate backupSubscriber in cache: %w", err)
		}
	}
	return nil
}
