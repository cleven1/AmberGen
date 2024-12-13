package examples

import (
	"context"
	"log"
	"multi-agent/agent"
)

func GroupDiscussionExample() {

	// 创建输出回调
	callback := NewDefaultOutputCallback()

	// 创建Agent组（设置最大轮数和是否并行）
	group := agent.NewGroup(3, true, callback)

	// 创建各个专家Agent
	productAgent := agent.NewAgent(
		"product_manager",
		"product_management",
		"我是一位产品经理，负责AI产品规划",
	)
	productAgent.SetStreamOutput(true)

	designerAgent := agent.NewAgent(
		"designer",
		"ui_ux_design",
		"我是一位资深UI/UX设计师",
	)
	designerAgent.SetStreamOutput(true)

	developerAgent := agent.NewAgent(
		"developer",
		"ai_development",
		"我是一位AI开发工程师",
	)
	developerAgent.SetStreamOutput(true)

	operationAgent := agent.NewModelAgent(
		"operation",
		"operation_management",
		"我是一位运营专家",
		"qwen-max",
	)
	operationAgent.SetStreamOutput(true)

	// 添加Agent到组
	group.AddAgent(productAgent)
	group.AddAgent(designerAgent)
	group.AddAgent(developerAgent)
	group.AddAgent(operationAgent)

	//  执行讨论
	ctx := context.Background()
	discussionTopic := `讨论主题：AI驱动的个性化学习助手系统...`

	_, err := group.Execute(ctx, discussionTopic)
	if err != nil {
		log.Fatal("执行失败:", err)
	}

	// 结果已通过回调输出
}
