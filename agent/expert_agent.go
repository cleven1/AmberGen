package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"multi-agent/oneapi"
	"multi-agent/tools"
	"strings"
	"sync"
	"time"
)

// ExpertAgent 专家型Agent
type ExpertAgent struct {
	*BaseAgent
	expertise   string                // 专业领域
	description string                // 专家描述
	useStream   bool                  // 是否使用流式输出
	callback    OutputCallback        // 回调函数
	tools       map[string]tools.Tool // 添加工具映射
	mu          sync.Mutex            // 添加互斥锁来保护通道操作
	Model       string                // 模型名称
	selector    AgentSelector         // 添加选择器
}

// NewExpertAgent 创建新的专家Agent
func NewSelectorAgent(name string, expertise string, description string) *ExpertAgent {
	return newDefaultAgent(name, expertise, description, "", NewDefaultSelector())
}
func NewSelectorModelAgent(name string, expertise string, description string, model string) *ExpertAgent {
	return newDefaultAgent(name, expertise, description, model, NewDefaultSelector())
}
func NewAgent(name string, expertise string, description string) *ExpertAgent {
	return newDefaultAgent(name, expertise, description, "", nil)
}
func NewModelAgent(name string, expertise string, description string, model string) *ExpertAgent {
	return newDefaultAgent(name, expertise, description, model, nil)
}
func newDefaultAgent(name string, expertise string, description string, model string, selector AgentSelector) *ExpertAgent {
	capabilities := []string{expertise}
	client := oneapi.NewClient()
	if client == nil {
		log.Fatal("创建AI客户端失败")
	}
	return &ExpertAgent{
		BaseAgent:   NewBaseAgent(name, capabilities, client, globalMemoryManager, getOrCreateTaskID()),
		expertise:   expertise,
		description: description,
		useStream:   false,
		tools:       make(map[string]tools.Tool),
		Model:       model,
		selector:    selector,
	}
}

// 全局的MemoryManager实例
var globalMemoryManager = NewMemoryManager()

// 全局任务ID
var currentTaskID string
var taskMutex sync.Mutex

// 获取或创建任务ID
func getOrCreateTaskID() string {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	if currentTaskID == "" {
		currentTaskID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}
	return currentTaskID
}

// ResetTaskID 重置任务ID（在新任务开始时调用）
func ResetTaskID() {
	taskMutex.Lock()
	defer taskMutex.Unlock()
	currentTaskID = ""
}

// SetStreamOutput 设置是否使用流式输出
func (e *ExpertAgent) SetStreamOutput(useStream bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.useStream = useStream
}

// handleStreamOutput 处理流式输出
func (e *ExpertAgent) handleStreamOutput(content string) {
	if e.callback != nil {
		e.callback.OnContent(e.Name(), content)
	}
}

// AddTool 添加工具
func (e *ExpertAgent) AddTool(tool tools.Tool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tools[tool.GetName()] = tool
}

// Execute 执行专家分析
func (e *ExpertAgent) Execute(ctx context.Context, input string) (string, error) {
	defer func() {
		// 确保在函数返回前调用OnComplete
		if !e.useStream && e.callback != nil {
			e.callback.OnComplete(e.Name())
		}
	}()
	if e.callback != nil {
		e.callback.OnStart(e.Name())
	}

	// 构建系统提示词
	systemPrompt := e.buildSystemPrompt()

	messages := []oneapi.ChatMessage{
		{Role: "system", Content: systemPrompt},
	}
	// 添加历史记录
	messages = append(messages, e.memory.GetHistory()...)
	// 添加当前输入
	messages = append(messages, oneapi.ChatMessage{
		Role:    "user",
		Content: input,
	})
	var finalResponse strings.Builder
	// 保存历史会话
	e.memory.AddMessage(oneapi.ChatMessage{
		Role:    "user",
		Content: input,
	})

	for {
		req := oneapi.ChatCompletionRequest{
			Messages:  messages,
			Stream:    e.useStream,
			Tools:     e.buildToolDefs(), // 添加工具定义
			MaxTokens: 1024,
			Model:     e.Model,
		}
		// 创建本地回调函数，确保其在范围内访问callback
		var streamCallback func(string)
		if e.useStream && e.callback != nil {
			// 预先通知开始
			e.callback.OnStart(e.Name())
			streamCallback = func(content string) {
				e.callback.OnContent(e.Name(), content)
			}
		}

		resp, err := e.client.ChatCompletion(ctx, req, streamCallback)
		if err != nil {
			return "", fmt.Errorf("专家分析失败: %w", err)
		}
		message := resp.Choices[0].Message
		// 处理工具调用
		if len(message.ToolCalls) > 0 {
			if e.useStream && e.callback != nil {
				e.callback.OnContent(e.Name(), "\n\n正在调用工具...\n")
			}

			// 处理每个工具调用
			for _, toolCall := range message.ToolCalls {
				// 验证工具调用
				if toolCall.Function.Name == "" || toolCall.Function.Arguments == "" {
					continue
				}

				// 验证参数
				var params map[string]interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
					continue
				}

				// 执行工具调用
				toolResult, err := e.executeToolCall(ctx, toolCall)
				if err != nil {
					if e.useStream && e.callback != nil {
						e.callback.OnContent(e.Name(), fmt.Sprintf("\n工具调用失败: %v\n", err))
					}
					return "", fmt.Errorf("工具调用失败: %w", err)
				}

				// 添加到消息历史
				messages = append(messages, oneapi.ChatMessage{
					Role:      "assistant",
					ToolCalls: []oneapi.ToolCall{toolCall},
				})
				messages = append(messages, oneapi.ChatMessage{
					Role:       "tool",
					ToolCallID: toolCall.ID,
					Content:    toolResult,
				})

				resultStr := fmt.Sprintf("\n工具 %s 执行结果：\n%s\n",
					toolCall.Function.Name, toolResult)

				if e.useStream && e.callback != nil {
					e.callback.OnContent(e.Name(), resultStr)
				}

				finalResponse.WriteString(resultStr)

				// 存储对话历史
				e.memory.AddMessage(oneapi.ChatMessage{
					Role:    "assistant",
					Content: finalResponse.String(),
				})

				// 添加短暂延迟，确保输出顺序
				time.Sleep(100 * time.Millisecond)
			}

			// 继续对话以处理工具结果
			continue
		}

		// 处理正常响应
		if resp.Choices[0].Message.Content != "" {
			if !e.useStream && e.callback != nil {
				e.callback.OnContent(e.Name(), resp.Choices[0].Message.Content)
			}
			finalResponse.WriteString(resp.Choices[0].Message.Content)
			// 存储对话历史
			e.memory.AddMessage(oneapi.ChatMessage{
				Role:    "assistant",
				Content: finalResponse.String(),
			})
		}
		break
	}

	// 如果是流式输出，在这里调用完成回调
	if e.useStream && e.callback != nil {
		e.callback.OnComplete(e.Name())
	} else {
		// 对于非流式输出，在这里发送内容
		e.callback.OnContent(e.Name(), finalResponse.String())
	}

	return finalResponse.String(), nil
}

func (e *ExpertAgent) executeToolCall(ctx context.Context, toolCall oneapi.ToolCall) (string, error) {
	tool, exists := e.tools[toolCall.Function.Name]
	if !exists {
		return "", fmt.Errorf("未找到工具: %s", toolCall.Function.Name)
	}

	// 解析参数
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	// 执行工具
	result, err := tool.Execute(ctx, params)
	if err != nil {
		return "", fmt.Errorf("工具执行失败: %w", err)
	}

	// 将结果转换为字符串
	// 转换结果
	var resultStr string
	switch v := result.(type) {
	case string:
		resultStr = v
	case []byte:
		resultStr = string(v)
	default:
		resultBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("结果格式化失败: %w", err)
		}
		resultStr = string(resultBytes)
	}
	return resultStr, nil
}

// SetCallback 设置输出回调
func (e *ExpertAgent) SetCallback(callback OutputCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.callback = callback
}

// buildToolDefs 构建工具定义
func (e *ExpertAgent) buildToolDefs() []oneapi.ToolDef {
	var defs []oneapi.ToolDef
	for _, tool := range e.tools {
		params := make(map[string]oneapi.Property)
		for name, spec := range tool.GetParameters() {
			params[name] = oneapi.Property{
				Type:        spec.Type,
				Description: spec.Description,
				Enum:        spec.Enum,
			}
		}

		defs = append(defs, oneapi.ToolDef{
			Type: "function",
			Function: oneapi.Tool{
				Name:        tool.GetName(),
				Description: tool.GetDescription(),
				Parameters: oneapi.Parameters{
					Type:       "object",
					Properties: params,
					Required:   getRequiredParams(tool.GetParameters()),
				},
			},
		})
	}
	return defs
}

// handleToolCalls 处理工具调用
func (e *ExpertAgent) handleToolCalls(ctx context.Context, toolCalls []oneapi.ToolCall) (string, error) {
	var results []string

	for _, call := range toolCalls {
		tool, exists := e.tools[call.Function.Name]
		if !exists {
			return "", fmt.Errorf("tool not found: %s", call.Function.Name)
		}

		// 解析参数
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(call.Function.Arguments), &params); err != nil {
			return "", fmt.Errorf("invalid tool arguments: %w", err)
		}

		// 执行工具调用
		result, err := tool.Execute(ctx, params)
		if err != nil {
			return "", fmt.Errorf("tool execution failed: %w", err)
		}

		// 转换结果为字符串
		resultStr := fmt.Sprintf("工具 %s 的执行结果：\n%v", call.Function.Name, result)
		results = append(results, resultStr)
	}

	return strings.Join(results, "\n\n"), nil
}

// getRequiredParams 获取必需参数列表
func getRequiredParams(params map[string]tools.ParameterSpec) []string {
	var required []string
	for name, param := range params {
		if param.Required {
			required = append(required, name)
		}
	}
	return required
}

func (e *ExpertAgent) buildSystemPrompt() string {
	var prompt strings.Builder
	prompt.WriteString(fmt.Sprintf(`你是一位%s领域的专家。%s。

  你的职责是：
  1. 基于你的专业知识，对讨论主题提供专业的见解
  2. 确保你的回应与之前的讨论保持连贯性
  3. 提供具体的、可操作的建议
  4. 适当质疑或补充其他专家的观点
  5. 基于你的专业领域提供独特的视角
  
  请记住你是%s领域的专家，所有回答都应该体现你的专业性。`,
		e.expertise,
		e.description,
		e.expertise,
	))
	prompt.WriteString("你可以使用以下工具来完成任务。使用工具时，请严格按照以下格式提供参数：\n\n")

	for _, tool := range e.tools {
		prompt.WriteString(fmt.Sprintf("工具名称：%s\n", tool.GetName()))
		prompt.WriteString(fmt.Sprintf("描述：%s\n", tool.GetDescription()))
		prompt.WriteString("参数：\n")

		for name, param := range tool.GetParameters() {
			required := param.Required
			prompt.WriteString(fmt.Sprintf("- %s (%s)", name, param.Type))
			if required {
				prompt.WriteString("【必需】")
			}
			prompt.WriteString(fmt.Sprintf(": %s\n", param.Description))
			if len(param.Enum) > 0 {
				prompt.WriteString(fmt.Sprintf("  可选值: %s\n", strings.Join(param.Enum, ", ")))
			}
		}
	}

	prompt.WriteString("\n注意：\n")
	prompt.WriteString("1. 调用工具时必须提供完整的参数\n")
	prompt.WriteString("2. 参数必须符合指定的类型\n")
	prompt.WriteString("3. 必需参数不能省略\n")

	return prompt.String()
}
