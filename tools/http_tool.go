package tools

import (
  "bytes"
  "context"
  "encoding/json"
  "fmt"
  "net/http"
)

// HTTPTool 实现HTTP调用的工具
type HTTPTool struct {
  *BaseTool
  URL     string
  Method  string
  Headers map[string]string
  client  *http.Client
}

// NewHTTPTool 创建新的HTTP工具
func NewHTTPTool(name, url, method string, headers map[string]string) *HTTPTool {
  tool := &HTTPTool{
    BaseTool: NewBaseTool(
      name,
      fmt.Sprintf("HTTP %s tool for %s", method, url),
    ),
    URL:     url,
    Method:  method,
    Headers: headers,
    client:  &http.Client{},
  }

  // 添加基础参数规格
  tool.AddParameter("body", ParameterSpec{
    Type:        "object",
    Description: "Request body",
    Required:    false,
  })

  tool.AddParameter("queryParams", ParameterSpec{
    Type:        "object",
    Description: "URL query parameters",
    Required:    false,
  })

  return tool
}

func (t *HTTPTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
  // 处理请求体
  var body []byte
  if bodyData, ok := params["body"]; ok {
    var err error
    body, err = json.Marshal(bodyData)
    if err != nil {
      return nil, fmt.Errorf("failed to marshal request body: %w", err)
    }
  }

  // 创建请求
  req, err := http.NewRequestWithContext(ctx, t.Method, t.URL, bytes.NewBuffer(body))
  if err != nil {
    return nil, fmt.Errorf("failed to create request: %w", err)
  }

  // 设置请求头
  for k, v := range t.Headers {
    req.Header.Set(k, v)
  }

  // 发送请求
  resp, err := t.client.Do(req)
  if err != nil {
    return nil, fmt.Errorf("request failed: %w", err)
  }
  defer resp.Body.Close()

  // 处理响应
  var result interface{}
  if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
    return nil, fmt.Errorf("failed to decode response: %w", err)
  }

  return result, nil
}
