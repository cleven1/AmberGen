package agent

import (
  "context"
  "sync"
)

type Agent interface {
  Execute(ctx context.Context, input string) (string, error)
  GetCapabilities() []string
  Name() string
}

type Manager struct {
  agents []Agent
  tools  map[string]Tool
  mu     sync.RWMutex
}

func NewManager() *Manager {
  return &Manager{
    agents: make([]Agent, 0),
    tools:  make(map[string]Tool),
  }
}

func (m *Manager) RegisterAgent(agent Agent) {
  m.mu.Lock()
  defer m.mu.Unlock()
  m.agents = append(m.agents, agent)
}

func (m *Manager) RegisterTool(name string, tool Tool) {
  m.mu.Lock()
  defer m.mu.Unlock()
  m.tools[name] = tool
}

func (m *Manager) FindBestAgent(task string) Agent {
  // Implement agent selection logic based on capabilities
  return m.agents[0]
}
