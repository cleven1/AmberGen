package examples

import (
	"context"
	"fmt"
	"log"
	"multi-agent/agent"
	"multi-agent/tools"
)

func WebSearchExample() {
	// 创建回调
	callback := NewDefaultOutputCallback()

	// 创建判断Agent，设置更明确的系统提示词
	judgeAgent := agent.NewAgent(
		"judge",
		"search necessity assessment",
		`你是一个判断专家，负责判断用户问题是否需要联网搜索。

判断标准：
1. 如果问题涉及实时性信息(如新闻、比赛结果、天气等)，需要搜索
2. 如果问题涉及具体数据、统计信息，需要搜索
3. 如果问题是纯主观、理论或通用知识，不需要搜索

请直接回答"需要搜索"或"不需要搜索"，并简要说明原因。`,
	)

	// 创建问题拆分Agent，设置更明确的系统提示词
	splitAgent := agent.NewAgent(
		"splitter",
		"question decomposition",
		`你是一个问题拆分专家，负责将需要搜索的问题拆分为多个具体的搜索查询。

要求：
1. 将问题拆分为多个独立的搜索查询
2. 每个查询应该明确且具体
3. 查询应该覆盖问题的不同方面
4. 使用最优的搜索关键词

输出格式：
 数组: [<具体查询>]
...`,
	)

	// 创建搜索结果整合Agent，设置更明确的系统提示词
	integrateAgent := agent.NewAgent(
		"integrator",
		"result integration",
		`你是一个信息整合专家，负责将多个搜索结果整合成完整的答案。

要求：
1. 综合所有搜索结果
2. 确保信息的准确性和一致性
3. 清晰标注信息来源
4. 按照逻辑顺序组织内容
5. 如果信息有冲突，说明不同来源的差异

输出格式：
[回答内容，包含完整的信息整合]

数据来源：
- <列出所有使用的数据来源>`,
	)

	// 创建Web搜索工具
	searchTool := tools.NewNewsSearcher()

	// 将工具添加到搜索Agent
	// 创建执行搜索的Agent
	searchAgent := agent.NewAgent(
		"searcher",
		"web search executor",
		"执行网络搜索并返回结果",
	)
	searchAgent.AddTool(searchTool)

	// 创建依赖图
	graph := agent.NewDependencyGraph(1, callback)

	// 添加所有Agent
	graph.AddAgent(judgeAgent)
	graph.AddAgent(splitAgent)
	graph.AddAgent(searchAgent)
	graph.AddAgent(integrateAgent)

	// 设置依赖关系
	graph.AddDependency("splitter", "judge")
	graph.AddDependency("searcher", "splitter")
	graph.AddDependency("integrator", "searcher")

	// 执行示例查询
	ctx := context.Background()
	query := "上海今天AI新闻日报"

	fmt.Printf("用户问题: %s\n", query)
	_, err := graph.Execute(ctx, query)
	if err != nil {
		log.Fatal("执行失败:", err)
	}
}
