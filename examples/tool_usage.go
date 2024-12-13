package examples

import (
	"context"
	"log"
	"multi-agent/agent"
	"multi-agent/tools"
)

func ToolUsageExample() {

	// 创建工具
	calculator := tools.NewCalculator()
	newsSearcher := tools.NewNewsSearcher("博查搜索API Key")

	// 创建工具注册中心
	registry := tools.NewToolRegistry()
	registry.Register(calculator.GetName(), calculator)
	registry.Register(newsSearcher.GetName(), newsSearcher)

	// 创建专家Agent
	researchAgent := agent.NewAgent(
		"researcher",
		"research_analysis",
		"我是一位研究分析师，负责新闻分析和数据处理",
	)
	// researchAgent.SetStreamOutput(true)
	researchAgent.AddTool(calculator)
	researchAgent.AddTool(newsSearcher)

	// 创建回调
	callback := NewDefaultOutputCallback()

	// 创建执行组
	group := agent.NewGroup(1, false, callback)
	group.AddAgent(researchAgent)

	// 执行分析任务
	ctx := context.Background()
	input := `请帮我完成以下任务：
	1. 搜索关于"人工智能"的最新新闻，并总结最近3天的相关报道
	2. 计算 235 + 567 的结果

	请先进行新闻搜索和总结，然后再进行计算。`

	_, err := group.Execute(ctx, input)
	if err != nil {
		log.Fatal("执行失败:", err)
	}
}
