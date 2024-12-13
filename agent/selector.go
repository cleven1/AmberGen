package agent

import (
	"context"
	"fmt"
	"multi-agent/oneapi"
)

// AgentSelector Agent选择器接口
type AgentSelector interface {
	// SelectAgents 根据输入选择最合适的Agents
	SelectAgents(input string, agents []*ExpertAgent, limit int) []*ExpertAgent
}

type DefaultSelector struct {
	client *oneapi.Client
}

func NewDefaultSelector() *DefaultSelector {
	return &DefaultSelector{
		client: oneapi.NewClient(),
	}
}

// SelectAgents 使用LLM进行Agent选择
func (s *DefaultSelector) SelectAgents(input string, agents []*ExpertAgent, limit int) []*ExpertAgent {
	if len(agents) == 0 {
		return nil
	}

	// 构建Handoff提示词
	prompt := s.buildHandoffPrompt(input, agents)

	// 询问LLM选择最合适的Agent
	selectedIndex, err := s.askLLMForSelection(context.Background(), prompt, len(agents))
	if err != nil {
		return nil
	}

	if selectedIndex >= 0 && selectedIndex < len(agents) {
		return []*ExpertAgent{agents[selectedIndex]}
	}

	return nil
}

// buildHandoffPrompt 构建Handoff提示词
func (s *DefaultSelector) buildHandoffPrompt(input string, agents []*ExpertAgent) string {
	prompt := `As an AI coordinator, analyze the user input and select the most suitable agent based on their expertise and capabilities.
  
  User Input: %s
  
  Available Agents:
  %s
  
  Your task:
  1. Analyze the input characteristics (language, content, requirements)
  2. Review each agent's expertise and capabilities
  3. Select the most suitable agent for handling this input
  4. Return ONLY the index number (0-%d) of the selected agent
  
  Response format: Single number representing the selected agent's index`

	// 构建Agent列表描述
	var agentDescriptions string
	for i, agent := range agents {
		agentDescriptions += fmt.Sprintf("%d. Name: %s\n   Expertise: %s\n   Description: %s\n\n",
			i, agent.Name(), agent.expertise, agent.description)
	}

	return fmt.Sprintf(prompt, input, agentDescriptions, len(agents)-1)
}

// askLLMForSelection 询问LLM选择Agent
func (s *DefaultSelector) askLLMForSelection(ctx context.Context, prompt string, agentCount int) (int, error) {
	req := oneapi.ChatCompletionRequest{
		Messages: []oneapi.ChatMessage{
			{
				Role:    "system",
				Content: "You are an AI coordinator responsible for selecting the most suitable agent for handling user inputs.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 10, // 只需要简短的数字响应
		Model:     s.client.Model,
	}

	resp, err := s.client.ChatCompletion(ctx, req, nil)
	if err != nil {
		return -1, fmt.Errorf("LLM selection failed: %w", err)
	}

	// 解析LLM的响应
	var selectedIndex int
	content := resp.Choices[0].Message.Content
	_, err = fmt.Sscanf(content, "%d", &selectedIndex)
	if err != nil {
		return -1, fmt.Errorf("invalid LLM response format: %w", err)
	}

	// 验证索引范围
	if selectedIndex < 0 || selectedIndex >= agentCount {
		return -1, fmt.Errorf("invalid agent index: %d", selectedIndex)
	}

	return selectedIndex, nil
}
