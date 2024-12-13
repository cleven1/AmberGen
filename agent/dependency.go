package agent

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type DependencyGraph struct {
	nodes        map[string]*Node
	maxRounds    int                 // 整体最大轮数
	roundResults []map[string]string // 每轮的结果
	callback     OutputCallback      // 添加回调接口
	mu           sync.RWMutex
}

type Node struct {
	agent        Agent
	dependencies []*Node
	capability   float64 // 能力分数，用于选择最合适的Agent
	visited      bool    // 用于追踪是否已参与当前轮次讨论
}

func NewDependencyGraph(maxRounds int, callback OutputCallback) *DependencyGraph {
	if maxRounds < 1 {
		maxRounds = 1 // 确保至少有一轮讨论
	}
	return &DependencyGraph{
		nodes:        make(map[string]*Node),
		maxRounds:    maxRounds,
		roundResults: make([]map[string]string, 0),
		callback:     callback,
	}
}

func (d *DependencyGraph) AddAgent(agent Agent) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 基于Agent的能力计算初始分数
	capability := calculateAgentCapability(agent)

	d.nodes[agent.Name()] = &Node{
		agent:        agent,
		dependencies: make([]*Node, 0),
		capability:   capability,
		visited:      false,
	}
}

func (d *DependencyGraph) AddDependency(dependent, dependency string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	depNode, ok := d.nodes[dependent]
	if !ok {
		return fmt.Errorf("%w: %s", ErrAgentNotFound, dependent)
	}

	depOnNode, ok := d.nodes[dependency]
	if !ok {
		return fmt.Errorf("%w: %s", ErrAgentNotFound, dependency)
	}

	// 检查是否已存在该依赖
	for _, dep := range depNode.dependencies {
		if dep.agent.Name() == dependency {
			return nil
		}
	}

	depNode.dependencies = append(depNode.dependencies, depOnNode)
	return nil
}

// Execute 执行整个图的对话
func (d *DependencyGraph) Execute(ctx context.Context, input string) ([]map[string]string, error) {
	d.mu.Lock()

	// 为所有Agent设置回调
	for _, node := range d.nodes {
		if expertAgent, ok := node.agent.(*ExpertAgent); ok {
			expertAgent.SetCallback(d.callback)
		}
	}

	defer d.mu.Unlock()

	// 初始化结果存储
	d.roundResults = make([]map[string]string, 0, d.maxRounds)
	for i := range d.roundResults {
		d.roundResults[i] = make(map[string]string)
	}

	// 第一轮：按依赖顺序执行
	firstRoundResults, err := d.executeFirstRound(ctx, input)
	if err != nil {
		return nil, err
	}
	d.roundResults = append(d.roundResults, firstRoundResults)

	// 后续轮次：选择最合适的Agent，并传递完整上下文
	currentInput := input + "\n\n" + d.combineRoundResults(0, firstRoundResults)

	for round := 1; round < d.maxRounds; round++ {
		roundResults, err := d.executeSubsequentRound(ctx, currentInput, round)
		if err != nil {
			return nil, err
		}

		d.roundResults = append(d.roundResults, roundResults)
		// 更新输入，包含所有历史信息
		currentInput = input + "\n\n" + d.combineRoundResults(round, roundResults)
	}

	// 通知所有执行完成
	if d.callback != nil {
		d.callback.OnAllComplete(d.roundResults)
	}

	return d.roundResults, nil
}

// executeFirstRound 执行第一轮讨论（按依赖顺序）
func (d *DependencyGraph) executeFirstRound(ctx context.Context, input string) (map[string]string, error) {
	results := make(map[string]string)
	visited := make(map[string]bool)

	if d.maxRounds > 1 {
		// 构建第一轮的提示词
		input = fmt.Sprintf(`这是多轮讨论的第一轮。

  讨论主题：
  %s
  
  请根据你的专业角度提供意见。`, input)
	}

	// 重置所有节点的访问状态
	for _, node := range d.nodes {
		node.visited = false
	}

	// 按依赖顺序执行
	var execute func(*Node) error
	execute = func(node *Node) error {
		if visited[node.agent.Name()] {
			return nil
		}

		// 先执行依赖
		for _, dep := range node.dependencies {
			if err := execute(dep); err != nil {
				return err
			}
		}
		// 执行当前节点
		result, err := node.agent.Execute(ctx, input)
		if err != nil {
			return err
		}

		results[node.agent.Name()] = result
		visited[node.agent.Name()] = true
		node.visited = true

		return nil
	}

	// 如果没有明确的依赖关系，按能力分数排序执行
	if !d.hasAnyDependencies() {
		sortedNodes := d.getSortedNodesByCapability()
		for _, node := range sortedNodes {
			if err := execute(node); err != nil {
				return nil, err
			}
		}
	} else {
		// 有依赖关系时，确保所有节点都被执行
		for _, node := range d.nodes {
			if err := execute(node); err != nil {
				return nil, err
			}
		}
	}

	if d.callback != nil {
		d.callback.OnRoundComplete(0, results)
	}

	return results, nil
}

// executeSubsequentRound 执行后续轮次（选择最合适的Agent）
func (d *DependencyGraph) executeSubsequentRound(ctx context.Context, input string, round int) (map[string]string, error) {
	results := make(map[string]string)

	// 获取按能力排序的节点
	sortedNodes := d.getSortedNodesByCapability()

	// 选择最合适的Agent参与讨论（这里可以根据具体需求调整选择策略）
	selectedCount := len(sortedNodes)/2 + 1 // 至少选择一半的Agent

	if d.maxRounds > 1 {
		// 构建后续轮次的提示词
		input = fmt.Sprintf(`这是多轮讨论的第 %d 轮。

  %s
  
  请基于上述讨论内容，从你的专业角度提出新的见解或补充意见。特别注意：
  1. 避免重复之前已经提到的观点
  2. 对之前的讨论进行补充和完善
  3. 如果发现之前的讨论中存在问题，请指出并提供改进建议
  `, round+1, input)
	}

	for i := 0; i < selectedCount; i++ {
		node := sortedNodes[i]

		if d.callback != nil {
			d.callback.OnStart(node.agent.Name())
		}

		// 执行时传入完整上下文
		result, err := node.agent.Execute(ctx, input)
		if err != nil {
			return nil, err
		}

		results[node.agent.Name()] = result
		// 更新节点能力分数
		d.updateNodeCapability(node, input, result)
	}

	if d.callback != nil {
		d.callback.OnRoundComplete(round, results)
	}

	return results, nil
}

// 辅助方法
func (d *DependencyGraph) hasAnyDependencies() bool {
	for _, node := range d.nodes {
		if len(node.dependencies) > 0 {
			return true
		}
	}
	return false
}

func (d *DependencyGraph) getSortedNodesByCapability() []*Node {
	nodes := make([]*Node, 0, len(d.nodes))
	for _, node := range d.nodes {
		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].capability > nodes[j].capability
	})

	return nodes
}

func (d *DependencyGraph) updateNodeCapability(node *Node, input, output string) {
	// 这里可以实现更复杂的能力评分更新逻辑
	relevanceScore := calculateRelevance(input, output)
	node.capability = (node.capability + relevanceScore) / 2
}

func (d *DependencyGraph) combineRoundResults(round int, results map[string]string) string {
	var sb strings.Builder

	// 添加历史讨论记录
	if round > 0 {
		sb.WriteString("前面的讨论总结：\n")
		for i := 0; i < round; i++ {
			sb.WriteString(fmt.Sprintf("\n第 %d 轮讨论：\n", i+1))
			for name, result := range d.roundResults[i] {
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
