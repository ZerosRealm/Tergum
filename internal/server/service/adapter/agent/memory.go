package agent

import (
	"fmt"
	"sync"

	"zerosrealm.xyz/tergum/internal/types"
)

/*
	Cache
*/

type MemoryCache struct {
	mutex sync.RWMutex
	agents map[string]*types.Agent
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		mutex: sync.RWMutex{},
		agents: make(map[string]*types.Agent),
	}
}

func (s *MemoryCache) Get(id []byte) (*types.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	agent, ok := s.agents[string(id)]
	if !ok {
		return nil, nil
	}

	return agent, nil
}

// TODO: Implement pagination.
func (s *MemoryCache) GetAll() ([]*types.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	agents := make([]*types.Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

func (s *MemoryCache) Add(agent *types.Agent) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.agents[fmt.Sprint(agent.ID)] = agent
	return nil
}

func (s *MemoryCache) Invalidate(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.agents, string(id))
	return nil
}

/*
	Storage
*/

type MemoryStorage struct {
	mutex sync.RWMutex
	agents map[string]*types.Agent
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mutex: sync.RWMutex{},
		agents: make(map[string]*types.Agent),
	}
}

func (s *MemoryStorage) Get(id []byte) (*types.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	agent, ok := s.agents[string(id)]
	if !ok {
		return nil, nil
	}

	return agent, nil
}

// TODO: Implement pagination.
func (s *MemoryStorage) GetAll() ([]*types.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	agents := make([]*types.Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

func (s *MemoryStorage) Create(agent *types.Agent) (*types.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := len(s.agents) + 1
	agent.ID = id

	s.agents[fmt.Sprint(agent.ID)] = agent

	return agent, nil
}

func (s *MemoryStorage) Update(agent *types.Agent) (*types.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.agents[fmt.Sprint(agent.ID)] = agent

	return agent, nil
}

func (s *MemoryStorage) Delete(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.agents, string(id))
	return nil
}
