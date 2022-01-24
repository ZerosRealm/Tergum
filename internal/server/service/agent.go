package service

import (
	"strconv"

	"zerosrealm.xyz/tergum/internal/types"
)

type AgentCache interface {
	Get(id []byte) (*types.Agent, error)
	GetAll() ([]*types.Agent, error)

	Add(agent *types.Agent) error
	Invalidate(id []byte) error
}

type AgentStorage interface {
	Get(id []byte) (*types.Agent, error)
	GetAll() ([]*types.Agent, error)
	Create(agent *types.Agent) (*types.Agent, error)
	Update(agent *types.Agent) (*types.Agent, error)
	Delete(id []byte) error
}

type AgentService struct {
	cache   AgentCache
	storage AgentStorage
}

func NewAgentService(cache *AgentCache, storage *AgentStorage) *AgentService {
	return &AgentService{
		cache:   *cache,
		storage: *storage,
	}
}

func (svc *AgentService) Get(id []byte) (*types.Agent, error) {
	if svc.cache != nil {
		agent, err := svc.storage.Get(id)
		if err != nil {
			return nil, err
		}

		if agent != nil {
			return agent, nil
		}
	}

	agent, err := svc.storage.Get(id)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (svc *AgentService) GetAll() ([]*types.Agent, error) {
	if svc.cache != nil {
		agents, err := svc.storage.GetAll()
		if err != nil {
			return nil, err
		}

		if len(agents) > 0 {
			return agents, nil
		}
	}

	agents, err := svc.storage.GetAll()
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (svc *AgentService) Create(agent *types.Agent) (*types.Agent, error) {
	agent, err := svc.storage.Create(agent)
	if err != nil {
		return nil, err
	}

	if svc.cache != nil {
		err = svc.cache.Add(agent)
		if err != nil {
			return nil, err
		}
	}

	return agent, nil
}

func (svc *AgentService) Update(agent *types.Agent) (*types.Agent, error) {
	agent, err := svc.storage.Update(agent)
	if err != nil {
		return nil, err
	}

	if svc.cache != nil {
		id := strconv.Itoa(agent.ID)
		err = svc.cache.Invalidate([]byte(id))
		if err != nil {
			return nil, err
		}
	}

	return agent, nil
}

func (svc *AgentService) Delete(id []byte) error {
	err := svc.storage.Delete(id)
	if err != nil {
		return err
	}

	if svc.cache != nil {
		err = svc.cache.Invalidate(id)
		if err != nil {
			return err
		}
	}
	return nil
}
