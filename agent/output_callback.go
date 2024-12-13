package agent

// OutputCallback 定义输出回调接口
type OutputCallback interface {
	// OnStart 当Agent开始输出时调用
	OnStart(agentName string)

	// OnContent 处理Agent的输出内容
	OnContent(agentName string, content string)

	// OnComplete 当Agent完成输出时调用
	OnComplete(agentName string)

	// OnRoundComplete 当一轮对话完成时调用
	OnRoundComplete(round int, results map[string]string)

	// OnComplete 当整个讨论完成时调用
	OnAllComplete(allResults []map[string]string)
}
