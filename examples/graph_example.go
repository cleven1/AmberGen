package examples

import (
	"context"
	"fmt"
	"log"
	"multi-agent/agent"
	"sync"
)

// DefaultOutputCallback 提供默认的输出回调实现
type DefaultOutputCallback struct {
	mu      sync.Mutex
	isFirst map[string]bool
}

func NewDefaultOutputCallback() *DefaultOutputCallback {
	return &DefaultOutputCallback{
		isFirst: make(map[string]bool),
	}
}

func (c *DefaultOutputCallback) OnStart(agentName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.isFirst[agentName] {
		c.isFirst[agentName] = true
		fmt.Printf("\n[%s 开始回答]\n", agentName)
	}
}

func (c *DefaultOutputCallback) OnContent(agentName string, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isFirst[agentName] {
		fmt.Printf("[%s]: ", agentName)
		c.isFirst[agentName] = false
	}
	fmt.Print(content)
}

func (c *DefaultOutputCallback) OnComplete(agentName string) {
	fmt.Printf("\n[%s 回答完成]\n", agentName)
}

func (c *DefaultOutputCallback) OnRoundComplete(round int, results map[string]string) {
	fmt.Printf("\n=== 第 %d 轮讨论完成 ===\n", round+1)
}

func (c *DefaultOutputCallback) OnAllComplete(allResults []map[string]string) {
	// fmt.Println("\n=== 讨论总结 ===")
	// for round, results := range allResults {
	// 	fmt.Printf("\n第 %d 轮讨论：\n", round+1)
	// 	for agentName, result := range results {
	// 		fmt.Printf("[%s] 最终意见: %s\n", agentName, result)
	// 	}
	// }
}

func AiFutureDiscussionExample() {
	// 创建各个角色的Agent
	productAgent := agent.NewModelAgent(
		"product_manager",
		"product_management",
		"我是一位产品经理，负责AI产品规划",
		"qwen-max",
	)

	designerAgent := agent.NewAgent(
		"designer",
		"ui_ux_design",
		"我是一位资深UI/UX设计师",
	)

	developerAgent := agent.NewAgent(
		"developer",
		"ai_development",
		"我是一位AI开发工程师",
	)

	operationAgent := agent.NewAgent(
		"operation",
		"operation_management",
		"我是一位运营专家",
	)

	//  创建输出回调
	callback := NewDefaultOutputCallback()

	// 创建依赖图，设置对话轮数最大轮数为3
	graph := agent.NewDependencyGraph(3, callback)

	// 添加所有Agent
	graph.AddAgent(productAgent)
	graph.AddAgent(designerAgent)
	graph.AddAgent(developerAgent)
	graph.AddAgent(operationAgent)

	// 设置依赖关系
	graph.AddDependency("designer", "product_manager")
	graph.AddDependency("developer", "designer")
	graph.AddDependency("operation", "developer")

	// 设置Agent使用流式输出
	productAgent.SetStreamOutput(true)
	designerAgent.SetStreamOutput(true)
	developerAgent.SetStreamOutput(true)
	operationAgent.SetStreamOutput(true)

	// 开始讨论
	ctx := context.Background()
	discussionTopic := `讨论主题：AI驱动的个性化学习助手系统...`

	// 执行并获取每轮结果
	_, err := graph.Execute(ctx, discussionTopic)
	if err != nil {
		log.Fatal("执行失败:", err)
	}

	// 输出每轮结果
	// for round, results := range roundResults {
	// 	fmt.Printf("\n=== 第 %d 轮讨论 ===\n", round+1)
	// 	for agent, result := range results {
	// 		fmt.Printf("[%s]: %s\n", agent, result)
	// 	}
	// }
}
