package agent

import (
	"multi-agent/oneapi"
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
