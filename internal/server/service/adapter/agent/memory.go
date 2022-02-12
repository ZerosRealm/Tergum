package agent

import (
	"fmt"
	"sync"

	"zerosrealm.xyz/tergum/internal/entities"
)

/*
	Cache
*/

type MemoryCache struct {
	mutex  sync.RWMutex
	agents map[string]*entities.Agent
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		mutex:  sync.RWMutex{},
		agents: make(map[string]*entities.Agent),
	}
}

func (s *MemoryCache) Get(id []byte) (*entities.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	agent, ok := s.agents[string(id)]
	if !ok {
		return nil, nil
	}

	return agent, nil
}

// TODO: Implement pagination.
func (s *MemoryCache) GetAll() ([]*entities.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	agents := make([]*entities.Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

func (s *MemoryCache) Add(agent *entities.Agent) error {
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
	mutex  sync.RWMutex
	agents map[string]*entities.Agent
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mutex:  sync.RWMutex{},
		agents: make(map[string]*entities.Agent),
	}
}

func (s *MemoryStorage) Get(id []byte) (*entities.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	agent, ok := s.agents[string(id)]
	if !ok {
		return nil, nil
	}

	return agent, nil
}

// TODO: Implement pagination.
func (s *MemoryStorage) GetAll() ([]*entities.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	agents := make([]*entities.Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

func (s *MemoryStorage) Create(agent *entities.Agent) (*entities.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := len(s.agents) + 1
	agent.ID = id

	s.agents[fmt.Sprint(agent.ID)] = agent

	return agent, nil
}

func (s *MemoryStorage) Update(agent *entities.Agent) (*entities.Agent, error) {
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
