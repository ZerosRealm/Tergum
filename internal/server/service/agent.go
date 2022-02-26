package service

import (
	"fmt"
	"strconv"

	"zerosrealm.xyz/tergum/internal/entity"
)

type AgentCache interface {
	Get(id []byte) (*entity.Agent, error)
	GetAll() ([]*entity.Agent, error)

	Add(agent *entity.Agent) error
	Invalidate(id []byte) error
}

type AgentStorage interface {
	Get(id []byte) (*entity.Agent, error)
	GetAll() ([]*entity.Agent, error)
	Create(agent *entity.Agent) (*entity.Agent, error)
	Update(agent *entity.Agent) (*entity.Agent, error)
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

func (svc *AgentService) Get(id []byte) (*entity.Agent, error) {
	if svc.cache != nil {
		agent, err := svc.cache.Get(id)
		if err != nil {
			return nil, fmt.Errorf("agentSvc.Get: could not get agent from cache: %w", err)
		}

		if agent != nil {
			return agent, nil
		}
	}

	agent, err := svc.storage.Get(id)
	if err != nil {
		return nil, fmt.Errorf("agentSvc.Get: could not get agent from storage: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Add(agent)
		if err != nil {
			return nil, fmt.Errorf("agentSvc.Get: could not add agent to cache: %w", err)
		}
	}
	return agent, nil
}

func (svc *AgentService) GetAll() ([]*entity.Agent, error) {
	if svc.cache != nil {
		agents, err := svc.cache.GetAll()
		if err != nil {
			return nil, fmt.Errorf("agentSvc.GetAll: could not get agents from cache: %w", err)
		}

		if len(agents) > 0 {
			return agents, nil
		}
	}

	agents, err := svc.storage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("agentSvc.GetAll: could not get agents from storage: %w", err)
	}
	if svc.cache != nil {
		for _, agent := range agents {
			err = svc.cache.Add(agent)
			if err != nil {
				return nil, fmt.Errorf("agentSvc.GetAll: could not add agent to cache: %w", err)
			}
		}
	}
	return agents, nil
}

func (svc *AgentService) Create(agent *entity.Agent) (*entity.Agent, error) {
	agent, err := svc.storage.Create(agent)
	if err != nil {
		return nil, fmt.Errorf("agentSvc.Create: could not create agent: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Add(agent)
		if err != nil {
			return nil, fmt.Errorf("agentSvc.Create: could not add agent to cache: %w", err)
		}
	}

	return agent, nil
}

func (svc *AgentService) Update(agent *entity.Agent) (*entity.Agent, error) {
	agent, err := svc.storage.Update(agent)
	if err != nil {
		return nil, fmt.Errorf("agentSvc.Update: could not update agent: %w", err)
	}

	if svc.cache != nil {
		id := strconv.Itoa(agent.ID)
		err = svc.cache.Invalidate([]byte(id))
		if err != nil {
			return nil, fmt.Errorf("agentSvc.Update: could not invalidate agent in cache: %w", err)
		}
	}

	return agent, nil
}

func (svc *AgentService) Delete(id []byte) error {
	err := svc.storage.Delete(id)
	if err != nil {
		return fmt.Errorf("agentSvc.Delete: could not delete agent: %w", err)
	}

	if svc.cache != nil {
		err = svc.cache.Invalidate(id)
		if err != nil {
			return fmt.Errorf("agentSvc.Delete: could not invalidate agent in cache: %w", err)
		}
	}
	return nil
}
