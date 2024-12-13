package oneapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"multi-agent/config"
	"net/http"
	"strings"
)

type Client struct {
	APIKey  string
	BaseURL string
	Model   string
	client  *http.Client
}

func NewClient() *Client {
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}
	return &Client{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
		client:  &http.Client{},
	}
}

// ChatCompletion 支持流式和非流式输出
func (c *Client) ChatCompletion(ctx context.Context, req ChatCompletionRequest, callback func(string)) (*ChatCompletionResponse, error) {
	if req.Stream {
		return c.streamChatCompletion(ctx, req, callback)
	}
	return c.normalChatCompletion(ctx, req)
}

// normalChatCompletion 非流式输出
func (c *Client) normalChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	url := fmt.Sprintf("%s/v1/chat/completions", c.BaseURL)
	// 确保stream为false
	req.Stream = false
	// 序列化请求
	if req.Model == "" {
		req.Model = c.Model
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	// 发送请求
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error: %s (type: %s, code: %s)",
			errorResp.Error.Message,
			errorResp.Error.Type,
			errorResp.Error.Code)
	}

	// 解析响应
	var response ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &response, nil
}

// streamChatCompletion 流式输出
func (c *Client) streamChatCompletion(ctx context.Context, req ChatCompletionRequest, callback func(string)) (*ChatCompletionResponse, error) {
	url := fmt.Sprintf("%s/v1/chat/completions", c.BaseURL)

	// 确保stream为true
	req.Stream = true
	if req.Model == "" {
		req.Model = c.Model
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	response := &ChatCompletionResponse{
		Choices: make([]Choice, 1),
	}

	// 用于累积工具调用的map
	toolCallStates := make(map[int]*toolCallState)
	var currentContent strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read stream failed: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		line = strings.TrimPrefix(line, "data: ")
		if line == "[DONE]" {
			break
		}

		var streamResp StreamResponse
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) > 0 {
			choice := streamResp.Choices[0]

			// 处理常规内容
			if choice.Delta.Content != "" {
				if callback != nil {
					callback(choice.Delta.Content)
				}
				currentContent.WriteString(choice.Delta.Content)
			}

			// 处理工具调用
			if len(choice.Delta.ToolCalls) > 0 {
				for _, tc := range choice.Delta.ToolCalls {
					state, exists := toolCallStates[tc.Index]
					if !exists {
						state = &toolCallState{
							ID:   tc.ID,
							Type: tc.Type,
						}
						toolCallStates[tc.Index] = state
					}

					// 更新工具名称
					if tc.Function.Name != "" {
						state.Name = tc.Function.Name
					}

					// 累积参数
					if tc.Function.Arguments != "" {
						state.Arguments.WriteString(tc.Function.Arguments)
					}
				}
			}

			// 检查完成状态
			if choice.FinishReason == "tool_calls" {
				var toolCalls []ToolCall
				for _, state := range toolCallStates {
					args := state.Arguments.String()

					// 验证参数完整性
					var testMap map[string]interface{}
					if err := json.Unmarshal([]byte(args), &testMap); err == nil {
						toolCalls = append(toolCalls, ToolCall{
							ID:   state.ID,
							Type: state.Type,
							Function: struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							}{
								Name:      state.Name,
								Arguments: args,
							},
						})
					}
				}

				response.Choices[0].Message = ChatMessage{
					Role:      "assistant",
					Content:   currentContent.String(),
					ToolCalls: toolCalls,
				}

				return response, nil
			}
		}
	}

	// 如果没有工具调用，返回普通响应
	response.Choices[0].Message = ChatMessage{
		Role:    "assistant",
		Content: currentContent.String(),
	}
	return response, nil
}
