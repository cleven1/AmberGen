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

	// 创建搜索Agent，添加错误处理和重试机制
	searchAgent := agent.NewAgent(
		"searcher",
		"web search executor",
		`你是一个网络搜索执行专家。你的职责是：

1. 执行搜索查询并获取结果
2. 如果搜索失败，尝试使用不同的关键词重试
3. 验证搜索结果的相关性
4. 确保所有子查询都得到处理
5. 保持搜索结果的完整性
6. 直接使用用户输入作为搜索查询

处理步骤：
1. 接收拆分后的查询条件
2. 对每个查询执行搜索
3. 验证结果是否满足需求
4. 如果结果不理想，调整关键词重试
5. 返回所有搜索结果

错误处理：
1. 搜索失败时进行重试
2. 记录失败原因
3. 使用备选词重新搜索

输出格式：
{
  "query": "搜索查询",
  "results": [搜索结果],
  "status": "成功/失败",
  "retries": 重试次数
}
`,
	)

	// 添加搜索工具
	searchTool := tools.NewNewsSearcher()
	searchAgent.AddTool(searchTool)

	// 创建结果整合Agent，增强结果验证
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
6. 验证所有查询都得到了结果
7. 标记可能缺失或不完整的信息

结果验证：
1. 检查每个子查询是否都有对应结果
2. 验证结果的相关性和完整性
3. 识别并标记可能的信息缺口

输出格式：
1. 综合答案
2. 信息来源列表
3. 完整性报告
4. 可能的信息缺口

数据来源：
- <列出所有使用的数据来源>`,
	)

	// 创建依赖图
	graph := agent.NewDependencyGraph(1, callback)

	// 添加所有Agent
	graph.AddAgent(searchAgent)
	graph.AddAgent(integrateAgent)

	// 设置依赖关系
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
