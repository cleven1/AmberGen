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
	memoryMgr    *MemoryManager
	taskID       string
}

// NewBaseAgent 创建新的基础Agent
func NewBaseAgent(name string, capabilities []string, client *oneapi.Client, memoryMgr *MemoryManager, taskID string) *BaseAgent {
	return &BaseAgent{
		name:         name,
		capabilities: capabilities,
		client:       client,
		memoryMgr:    memoryMgr,
		taskID:       taskID,
		memory:       memoryMgr.GetMemory(taskID),
	}
}

func (b *BaseAgent) Name() string {
	return b.name
}

func (b *BaseAgent) GetCapabilities() []string {
	return b.capabilities
}

// ClearTaskMemory 清理当前任务的Memory
func (b *BaseAgent) ClearTaskMemory() {
	if b.memoryMgr != nil {
		b.memoryMgr.ClearTaskMemory(b.taskID)
	}
}
