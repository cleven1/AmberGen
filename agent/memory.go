package agent

import (
	"multi-agent/oneapi"
	"sync"
)

// Memory 用于存储Agent的对话历史
type Memory struct {
	history []oneapi.ChatMessage
	maxSize int
}

// NewMemory 创建新的记忆存储
func NewMemory() *Memory {
	return &Memory{
		history: make([]oneapi.ChatMessage, 0),
		maxSize: 100,
	}
}

// AddMessage 添加新的对话消息
func (m *Memory) AddMessage(msg oneapi.ChatMessage) {
	if len(m.history) >= m.maxSize {
		// 如果超出最大容量，删除最早的消息
		m.history = m.history[1:]
	}
	m.history = append(m.history, msg)
}

// GetHistory 获取所有历史记录
func (m *Memory) GetHistory() []oneapi.ChatMessage {
	return m.history
}

// Clear 清空历史记录
func (m *Memory) Clear() {
	m.history = make([]oneapi.ChatMessage, 0)
}

// MemoryManager 管理不同任务的Memory实例
type MemoryManager struct {
	memories map[string]*Memory
	mutex    sync.RWMutex
}

// NewMemoryManager 创建新的MemoryManager
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		memories: make(map[string]*Memory),
		mutex:    sync.RWMutex{},
	}
}

// GetMemory 获取指定任务的Memory实例，如果不存在则创建新的
func (mm *MemoryManager) GetMemory(taskID string) *Memory {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if memory, exists := mm.memories[taskID]; exists {
		return memory
	}

	memory := NewMemory()
	mm.memories[taskID] = memory
	return memory
}

// ClearTaskMemory 清理指定任务的Memory
func (mm *MemoryManager) ClearTaskMemory(taskID string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	delete(mm.memories, taskID)
}
