package agent

import (
	"multi-agent/oneapi"
)

// BaseAgent 提供基础Agent实现
type BaseAgent struct {
	name         string
	capabilities []string
	client       *oneapi.Client
	memory       *Memory
}

// NewBaseAgent 创建新的基础Agent
func NewBaseAgent(name string, capabilities []string, client *oneapi.Client) *BaseAgent {
	return &BaseAgent{
		name:         name,
		capabilities: capabilities,
		client:       client,
		memory:       NewMemory(), // 确保初始化memory
	}
}

func (b *BaseAgent) Name() string {
	return b.name
}

func (b *BaseAgent) GetCapabilities() []string {
	return b.capabilities
}
