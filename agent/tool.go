package agent

import "context"

type Tool interface {
  Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
  GetDescription() string
}

type URLTool struct {
  URL         string
  Method      string
  Description string
}

func (t *URLTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
  // Implement HTTP call logic
  return nil, nil
}

func (t *URLTool) GetDescription() string {
  return t.Description
}
