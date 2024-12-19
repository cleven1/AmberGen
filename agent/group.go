package agent

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Group 支持并行执行的Agent组
type Group struct {
	agents       []*ExpertAgent
	maxRounds    int
	roundResults []map[string]string
	callback     OutputCallback
	selector     AgentSelector // 添加选择器
	parallel     bool          // 是否并发执行
	mu           sync.RWMutex
	// 添加memory统一管理
	memory *MemoryManager
}

// NewGroup 创建新的Agent组
func NewGroup(maxRounds int, parallel bool, callback OutputCallback) *Group {
	if maxRounds < 1 {
		maxRounds = 1
	}
	return &Group{
		agents:       make([]*ExpertAgent, 0),
		maxRounds:    maxRounds,
		roundResults: make([]map[string]string, 0, maxRounds),
		callback:     callback,
		parallel:     parallel,
		memory:       globalMemoryManager,
	}
}

// AddAgent 添加Agent到组
func (g *Group) AddAgent(agent *ExpertAgent) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 设置回调
	agent.SetCallback(g.callback)
	g.agents = append(g.agents, agent)
	g.selector = agent.selector
}

// Execute 执行组内所有Agent
func (g *Group) Execute(ctx context.Context, input string) ([]map[string]string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 初始化结果存储
	g.roundResults = make([]map[string]string, 0, g.maxRounds)

	// 第一轮：所有Agent都参与
	firstRoundResults, err := g.executeFirstRound(ctx, input)
	if err != nil {
		return nil, err
	}
	g.roundResults = append(g.roundResults, firstRoundResults)

	// 每轮结束时通知回调
	if g.callback != nil && g.selector == nil && g.maxRounds > 1 {
		g.callback.OnRoundComplete(0, firstRoundResults)
	}

	// 后续轮次：选择最合适的Agent
	currentInput := input + "\n\n" + g.combineRoundResults(0, firstRoundResults)

	for round := 1; round < g.maxRounds; round++ {
		roundResults, err := g.executeSubsequentRound(ctx, currentInput, round)
		if err != nil {
			return nil, err
		}

		g.roundResults = append(g.roundResults, roundResults)
		// 每轮结束时通知回调
		if g.callback != nil {
			g.callback.OnRoundComplete(round, roundResults)
		}
		currentInput = input + "\n\n" + g.combineRoundResults(round, roundResults)
	}

	// 通知所有执行完成
	if g.callback != nil {
		g.callback.OnAllComplete(g.roundResults)
	}

	// 执行完成后清理Memory
	g.memory.ClearTaskMemory(getOrCreateTaskID())

	return g.roundResults, nil
}

// 添加设置选择器的方法
func (g *Group) AddSelector(selector AgentSelector) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.selector = selector
}

// executeFirstRound 执行第一轮讨论
func (g *Group) executeFirstRound(ctx context.Context, input string) (map[string]string, error) {
	var selectedAgents []*ExpertAgent
	// 如果设置了选择器且不是并行执行，使用选择器选择Agent
	if g.selector != nil {
		// 选择最合适的Agent
		selectedAgents = g.selector.SelectAgents(input, g.agents, 1) // 限制为1个Agent
		if len(selectedAgents) > 0 {
			// 执行选中的Agent
			results := make(map[string]string)
			result, err := selectedAgents[0].Execute(ctx, input)
			if err != nil {
				return nil, fmt.Errorf("agent %s execution failed: %w", selectedAgents[0].Name(), err)
			}
			results[selectedAgents[0].Name()] = result
			return results, nil
		}
	}

	prompt := fmt.Sprintf(`这是多轮讨论的第一轮。

讨论主题：
%s

请根据你的专业角度提供意见。`, input)

	// 如果没有选择器或选择器没有选中Agent，使用原有逻辑
	sortedAgents := g.sortAgentsByCapability(g.agents)
	if g.parallel {
		return g.executeParallel(ctx, sortedAgents, prompt)
	}
	return g.executeSerial(ctx, sortedAgents, prompt)
}

// sortAgentsByCapability 按能力值排序Agent
func (g *Group) sortAgentsByCapability(agents []*ExpertAgent) []*ExpertAgent {
	type agentWithScore struct {
		agent *ExpertAgent
		score float64
	}

	// 计算所有Agent的能力值
	scores := make([]agentWithScore, len(agents))
	for i, agent := range agents {
		scores[i] = agentWithScore{
			agent: agent,
			score: calculateAgentCapability(agent),
		}
	}

	// 按能力值排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 转换回Agent切片
	result := make([]*ExpertAgent, len(agents))
	for i, score := range scores {
		result[i] = score.agent
	}

	return result
}

// executeSubsequentRound 执行后续轮次
func (g *Group) executeSubsequentRound(ctx context.Context, input string, round int) (map[string]string, error) {
	// 如果设置了选择器且不是并行执行，使用选择器选择Agent
	if g.selector != nil {
		selectedAgents := g.selector.SelectAgents(input, g.agents, 1) // 限制为1个Agent
		if len(selectedAgents) > 0 {
			// 执行选中的Agent
			results := make(map[string]string)
			result, err := selectedAgents[0].Execute(ctx, input)
			if err != nil {
				return nil, fmt.Errorf("agent %s execution failed: %w", selectedAgents[0].Name(), err)
			}
			results[selectedAgents[0].Name()] = result
			return results, nil
		}
	}
	prompt := fmt.Sprintf(`这是多轮讨论的第 %d 轮。

%s

请基于上述讨论内容，从你的专业角度提出新的见解或补充意见。特别注意：
1. 避免重复之前已经提到的观点
2. 对之前的讨论进行补充和完善
3. 如果发现之前的讨论中存在问题，请指出并提供改进建议`, round+1, input)

	// 如果没有选择器或选择器没有选中Agent，使用原有逻辑
	selectedCount := len(g.agents)/2 + 1
	selectedAgents := g.selectTopAgents(selectedCount)

	if g.parallel {
		return g.executeParallel(ctx, selectedAgents, prompt)
	}
	return g.executeSerial(ctx, selectedAgents, prompt)
}

// executeParallel 并行执行Agents
func (g *Group) executeParallel(ctx context.Context, agents []*ExpertAgent, input string) (map[string]string, error) {
	results := make(map[string]string)
	var mu sync.Mutex
	errors := make(chan error, len(agents))

	// 创建一个通道来控制执行顺序
	execChan := make(chan struct{}, 1)
	execChan <- struct{}{} // 初始令牌

	var wg sync.WaitGroup
	for _, agent := range agents {
		wg.Add(1)
		go func(a *ExpertAgent) {
			defer wg.Done()

			// 获取执行令牌
			<-execChan

			result, err := a.Execute(ctx, input)

			// 处理结果和错误
			if err != nil {
				errors <- fmt.Errorf("agent %s execution failed: %w", a.Name(), err)
				execChan <- struct{}{} // 释放令牌给下一个
				return
			}

			mu.Lock()
			results[a.Name()] = result
			mu.Unlock()

			// 等待一小段时间，让输出更自然
			time.Sleep(500 * time.Millisecond)

			// 释放令牌给下一个Agent
			execChan <- struct{}{}
		}(agent)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(execChan)
	close(errors)

	// 检查是否有错误
	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// executeSerial 串行执行Agents
func (g *Group) executeSerial(ctx context.Context, agents []*ExpertAgent, input string) (map[string]string, error) {
	results := make(map[string]string)

	for _, agent := range agents {
		result, err := agent.Execute(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("agent %s execution failed: %w", agent.Name(), err)
		}

		results[agent.Name()] = result

		// 等待一小段时间，让输出更自然
		time.Sleep(500 * time.Millisecond)
	}
	return results, nil
}

// selectTopAgents 选择能力值最高的Agent
func (g *Group) selectTopAgents(count int) []*ExpertAgent {
	if count > len(g.agents) {
		count = len(g.agents)
	}

	// 创建带有能力值的Agent切片
	type agentWithScore struct {
		agent *ExpertAgent
		score float64
	}

	scores := make([]agentWithScore, len(g.agents))
	for i, agent := range g.agents {
		scores[i] = agentWithScore{
			agent: agent,
			score: calculateAgentCapability(agent),
		}
	}

	// 按能力值排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 选择前count个Agent
	result := make([]*ExpertAgent, count)
	for i := 0; i < count; i++ {
		result[i] = scores[i].agent
	}

	return result
}

// combineRoundResults 组合一轮的结果
func (g *Group) combineRoundResults(round int, results map[string]string) string {
	var sb strings.Builder

	// 添加历史讨论记录
	if round > 0 {
		sb.WriteString("前面的讨论总结：\n")
		for i := 0; i < round; i++ {
			sb.WriteString(fmt.Sprintf("\n第 %d 轮讨论：\n", i+1))
			for name, result := range g.roundResults[i] {
				sb.WriteString(fmt.Sprintf("[%s]: %s\n", name, result))
			}
		}
		sb.WriteString("\n基于以上讨论，请继续：\n")
	}

	// 添加本轮讨论结果
	sb.WriteString("\n本轮讨论内容：\n")
	for name, result := range results {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", name, result))
	}

	return sb.String()
}
