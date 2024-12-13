package agent

import "errors"

// 定义错误常量
var (
  ErrAgentNotFound    = errors.New("agent not found")
  ErrInvalidDependency = errors.New("invalid dependency")
  ErrCircularDependency = errors.New("circular dependency detected")
  ErrExecutionFailed   = errors.New("agent execution failed")
  ErrToolNotFound      = errors.New("tool not found")
  ErrInvalidParameters = errors.New("invalid parameters")
)
