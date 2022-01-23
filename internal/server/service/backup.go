package service

import (
	"strconv"

	"zerosrealm.xyz/tergum/internal/types"
)

type BackupCache interface {
	Get(id []byte) (*types.Backup, error)
	GetAll() ([]*types.Backup, error)

	Add(backup *types.Backup) error
	Invalidate(id []byte) error
}

type BackupStorage interface {
	Get(id []byte) (*types.Backup, error)
	GetAll() ([]*types.Backup, error)
	Create(backup *types.Backup) (*types.Backup, error)
	Update(backup *types.Backup) (*types.Backup, error)
	Delete(id []byte) error
}

type BackupService struct {
	cache   *BackupCache
	storage *BackupStorage
}

func NewBackupService(cache *BackupCache, storage *BackupStorage) *BackupService {
	return &BackupService{
		cache:   cache,
		storage: storage,
	}
}

func (svc *BackupService) Get(id []byte) (*types.Backup, error) {
	if svc.cache != nil {
		backup, err := (*svc.cache).Get(id)
		if err != nil {
			return nil, err
		}

		if backup != nil {
			return backup, nil
		}
	}

	backup, err := (*svc.storage).Get(id)
	if err != nil {
		return nil, err
	}
	return backup, nil
}

func (svc *BackupService) GetAll() ([]*types.Backup, error) {
	if svc.cache != nil {
		backups, err := (*svc.cache).GetAll()
		if err != nil {
			return nil, err
		}

		if len(backups) > 0 {
			return backups, nil
		}
	}

	backups, err := (*svc.storage).GetAll()
	if err != nil {
		return nil, err
	}
	return backups, nil
}

func (svc *BackupService) Create(backup *types.Backup) (*types.Backup, error) {
	backup, err := (*svc.storage).Create(backup)
	if err != nil {
		return nil, err
	}

	if svc.cache != nil {
		err = (*svc.cache).Add(backup)
		if err != nil {
			return nil, err
		}
	}

	return backup, nil
}

func (svc *BackupService) Update(backup *types.Backup) (*types.Backup, error) {
	backup, err := (*svc.storage).Update(backup)
	if err != nil {
		return nil, err
	}

	if svc.cache != nil {
		id := strconv.Itoa(backup.ID)
		err = (*svc.cache).Invalidate([]byte(id))
		if err != nil {
			return nil, err
		}
	}

	return backup, nil
}

func (svc *BackupService) Delete(id []byte) error {
	err := (*svc.storage).Delete(id)
	if err != nil {
		return err
	}

	if svc.cache != nil {
		err = (*svc.cache).Invalidate(id)
		if err != nil {
			return err
		}
	}
	return nil
}
