package service

import (
	"fmt"

	"zerosrealm.xyz/tergum/internal/entities"
)

type SettingCache interface {
	Get(id []byte) (*entities.Setting, error)
	GetAll() ([]*entities.Setting, error)

	Add(setting *entities.Setting) error
	Invalidate(id []byte) error
}

type SettingStorage interface {
	Get(id []byte) (*entities.Setting, error)
	GetAll() ([]*entities.Setting, error)
	Create(setting *entities.Setting) (*entities.Setting, error)
	Update(setting *entities.Setting) (*entities.Setting, error)
	Delete(id []byte) error
}

type SettingService struct {
	cache   SettingCache
	storage SettingStorage
}

func NewSettingService(cache *SettingCache, storage *SettingStorage) *SettingService {
	return &SettingService{
		cache:   *cache,
		storage: *storage,
	}
}

func (svc *SettingService) Get(id []byte) (*entities.Setting, error) {
	if svc.cache != nil {
		setting, err := svc.storage.Get(id)
		if err != nil {
			return nil, fmt.Errorf("settingSvc.Get: could not get setting from cache: %w", err)
		}

		if setting != nil {
			return setting, nil
		}
	}

	setting, err := svc.storage.Get(id)
	if err != nil {
		return nil, fmt.Errorf("settingSvc.Get: could not get setting from storage: %w", err)
	}
	return setting, nil
}

func (svc *SettingService) GetAll() ([]*entities.Setting, error) {
	if svc.cache != nil {
		settings, err := svc.storage.GetAll()
		if err != nil {
			return nil, fmt.Errorf("settingSvc.GetAll: could not get settings from cache: %w", err)
		}

		if len(settings) > 0 {
			return settings, nil
		}
	}

	settings, err := svc.storage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("settingSvc.GetAll: could not get settings from storage: %w", err)
	}
	return settings, nil
}

func (svc *SettingService) Create(setting *entities.Setting) (*entities.Setting, error) {
	setting, err := svc.storage.Create(setting)
	if err != nil {
		return nil, fmt.Errorf("settingSvc.Create: could not create setting: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Add(setting)
		if err != nil {
			return nil, fmt.Errorf("settingSvc.Create: could not add setting to cache: %w", err)
		}
	}

	return setting, nil
}

func (svc *SettingService) Update(setting *entities.Setting) (*entities.Setting, error) {
	setting, err := svc.storage.Update(setting)
	if err != nil {
		return nil, fmt.Errorf("settingSvc.Update: could not update setting: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Invalidate([]byte(setting.Key))
		if err != nil {
			return nil, fmt.Errorf("settingSvc.Update: could not invalidate setting in cache: %w", err)
		}
	}

	return setting, nil
}

func (svc *SettingService) Delete(id []byte) error {
	err := svc.storage.Delete(id)
	if err != nil {
		return fmt.Errorf("settingSvc.Delete: could not delete setting: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Invalidate(id)
		if err != nil {
			return fmt.Errorf("settingSvc.Delete: could not invalidate setting in cache: %w", err)
		}
	}
	return nil
}
