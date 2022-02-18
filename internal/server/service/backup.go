package service

import (
	"fmt"
	"strconv"

	"zerosrealm.xyz/tergum/internal/entity"
)

type BackupCache interface {
	Get(id []byte) (*entity.Backup, error)
	GetAll() ([]*entity.Backup, error)

	Add(backup *entity.Backup) error
	Invalidate(id []byte) error
}

type BackupStorage interface {
	Get(id []byte) (*entity.Backup, error)
	GetAll() ([]*entity.Backup, error)
	Create(backup *entity.Backup) (*entity.Backup, error)
	Update(backup *entity.Backup) (*entity.Backup, error)
	Delete(id []byte) error
}

type BackupService struct {
	cache   BackupCache
	storage BackupStorage
}

func NewBackupService(cache *BackupCache, storage *BackupStorage) *BackupService {
	return &BackupService{
		cache:   *cache,
		storage: *storage,
	}
}

func (svc *BackupService) Get(id []byte) (*entity.Backup, error) {
	if svc.cache != nil {
		backup, err := svc.cache.Get(id)
		if err != nil {
			return nil, fmt.Errorf("backupSvc.Get: could not get backup from cache: %w", err)
		}

		if backup != nil {
			return backup, nil
		}
	}

	backup, err := svc.storage.Get(id)
	if err != nil {
		return nil, fmt.Errorf("backupSvc.Get: could not get backup from storage: %w", err)
	}
	return backup, nil
}

func (svc *BackupService) GetAll() ([]*entity.Backup, error) {
	if svc.cache != nil {
		backups, err := svc.cache.GetAll()
		if err != nil {
			return nil, fmt.Errorf("backupSvc.GetAll: could not get backups from cache: %w", err)
		}

		if len(backups) > 0 {
			return backups, nil
		}
	}

	backups, err := svc.storage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("backupSvc.GetAll: could not get backups from storage: %w", err)
	}
	return backups, nil
}

func (svc *BackupService) Create(backup *entity.Backup) (*entity.Backup, error) {
	backup, err := svc.storage.Create(backup)
	if err != nil {
		return nil, fmt.Errorf("backupSvc.Create: could not create backup: %w", err)
	}

	if svc.cache != nil {
		err = (svc.cache).Add(backup)
		if err != nil {
			return nil, fmt.Errorf("backupSvc.Create: could not add backup to cache: %w", err)
		}
	}

	return backup, nil
}

func (svc *BackupService) Update(backup *entity.Backup) (*entity.Backup, error) {
	backup, err := svc.storage.Update(backup)
	if err != nil {
		return nil, fmt.Errorf("backupSvc.Update: could not update backup: %w", err)
	}

	if svc.cache != nil {
		id := strconv.Itoa(backup.ID)
		err = (svc.cache).Invalidate([]byte(id))
		if err != nil {
			return nil, fmt.Errorf("backupSvc.Update: could not invalidate backup in cache: %w", err)
		}
	}

	return backup, nil
}

func (svc *BackupService) Delete(id []byte) error {
	err := svc.storage.Delete(id)
	if err != nil {
		return fmt.Errorf("backupSvc.Delete: could not delete backup: %w", err)
	}

	if svc.cache != nil {
		err = (svc.cache).Invalidate(id)
		if err != nil {
			return fmt.Errorf("backupSvc.Delete: could not invalidate backup in cache: %w", err)
		}
	}
	return nil
}
