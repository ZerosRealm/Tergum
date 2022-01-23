package agent

import (
	"fmt"
	"sync"

	"zerosrealm.xyz/tergum/internal/types"
)

type memoryStorage struct {
	mutex  sync.RWMutex
	agents map[string]*types.Agent
}

func NewMemoryStorage() *memoryStorage {
	return &memoryStorage{
		mutex:  sync.RWMutex{},
		agents: make(map[string]*types.Agent),
	}
}

func (s *memoryStorage) Get(id []byte) (*types.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	agent, ok := s.agents[string(id)]
	if !ok {
		return nil, nil
	}

	return agent, nil
}

// TODO: Implement pagination.
func (s *memoryStorage) GetAll() ([]*types.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	agents := make([]*types.Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

func (s *memoryStorage) Add(agent *types.Agent) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.agents[fmt.Sprint(agent.ID)] = agent
	return nil
}

func (s *memoryStorage) Invalidate(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.agents, string(id))
	return nil
}

func (s *memoryStorage) Create(agent *types.Agent) (*types.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := len(s.agents) + 1
	agent.ID = id

	s.agents[fmt.Sprint(agent.ID)] = agent

	return agent, nil
}

func (s *memoryStorage) Update(agent *types.Agent) (*types.Agent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.agents[fmt.Sprint(agent.ID)] = agent

	return agent, nil
}

func (s *memoryStorage) Delete(id []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.agents, string(id))
	return nil
}
