package service

import (
	"fmt"
	"strconv"

	"zerosrealm.xyz/tergum/internal/entities"
)

type AgentCache interface {
	Get(id []byte) (*entities.Agent, error)
	GetAll() ([]*entities.Agent, error)

	Add(agent *entities.Agent) error
	Invalidate(id []byte) error
}

type AgentStorage interface {
	Get(id []byte) (*entities.Agent, error)
	GetAll() ([]*entities.Agent, error)
	Create(agent *entities.Agent) (*entities.Agent, error)
	Update(agent *entities.Agent) (*entities.Agent, error)
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

func (svc *AgentService) Get(id []byte) (*entities.Agent, error) {
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

func (svc *AgentService) GetAll() ([]*entities.Agent, error) {
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
	for _, agent := range agents {
		if svc.cache != nil {
			err = svc.cache.Add(agent)
			if err != nil {
				return nil, fmt.Errorf("agentSvc.GetAll: could not add agent to cache: %w", err)
			}
		}
	}
	return agents, nil
}

func (svc *AgentService) Create(agent *entities.Agent) (*entities.Agent, error) {
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

func (svc *AgentService) Update(agent *entities.Agent) (*entities.Agent, error) {
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
