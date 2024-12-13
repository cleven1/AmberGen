package tools

import (
  "sync"
)

// ToolRegistry 工具注册中心
type ToolRegistry struct {
  tools map[string]Tool
  mu    sync.RWMutex
}

func NewToolRegistry() *ToolRegistry {
  return &ToolRegistry{
    tools: make(map[string]Tool),
  }
}

func (r *ToolRegistry) Register(name string, tool Tool) {
  r.mu.Lock()
  defer r.mu.Unlock()
  r.tools[name] = tool
}

func (r *ToolRegistry) Get(name string) (Tool, bool) {
  r.mu.RLock()
  defer r.mu.RUnlock()
  tool, exists := r.tools[name]
  return tool, exists
}

func (r *ToolRegistry) List() []string {
  r.mu.RLock()
  defer r.mu.RUnlock()
  var names []string
  for name := range r.tools {
    names = append(names, name)
  }
  return names
}
