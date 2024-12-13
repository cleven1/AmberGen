package examples

import (
	"context"
	"log"
	"multi-agent/agent"
)

func SelectorTestExample() {
	// 创建输出回调
	callback := NewDefaultOutputCallback()

	// 创建Agent组（设置最大轮数）
	group := agent.NewGroup(1, false, callback)

	// 创建各个专家Agent
	englishAgent := agent.NewSelectorAgent(
		"English Agent",
		"English",
		"You only speak English And tell me that you are a English agent",
	)
	englishAgent.SetStreamOutput(true)

	spanishAgent := agent.NewSelectorAgent(
		"Spanish Agent",
		"Spanish",
		"You only speak Spanish And tell me that you are a Spanish agent",
	)
	spanishAgent.SetStreamOutput(true)

	chineseAgent := agent.NewSelectorModelAgent(
		"Chinese Agent",
		"Chinese",
		"你只会说中文，并且告诉我你是中文智能体",
		"qwen-max",
	)
	chineseAgent.SetStreamOutput(true)

	// 添加Agent到组
	group.AddAgent(englishAgent)
	group.AddAgent(spanishAgent)
	group.AddAgent(chineseAgent)

	//  执行讨论
	ctx := context.Background()
	// 测试不同语言
	inputs := []string{
		"Hello, how are you?",
		"你好，最近好吗？",
		"¡Hola! ¿Cómo estás?",
	}

	for _, input := range inputs {
		log.Printf("\nTesting: %s\n", input)
		group.Execute(ctx, input)
	}
}
