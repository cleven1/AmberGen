package tools

import (
	"context"
	"fmt"
)

// Tool 定义工具接口
type Tool interface {
	// Execute 执行工具功能
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
	// GetDescription 获取工具描述
	GetDescription() string
	// GetName 获取工具名称
	GetName() string
	// GetParameters 获取参数说明
	GetParameters() map[string]ParameterSpec
}

// ParameterSpec 定义参数规格
type ParameterSpec struct {
	Type        string   `json:"type"`        // 参数类型: string, number, boolean等
	Description string   `json:"description"` // 参数描述
	Required    bool     `json:"required"`    // 是否必需
	Enum        []string `json:"enum"`        // 可选值列表(如果适用)
}

// BaseTool 提供基础工具实现
type BaseTool struct {
	name        string
	description string
	parameters  map[string]ParameterSpec
}

func NewBaseTool(name, description string) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		parameters:  make(map[string]ParameterSpec),
	}
}

// GetName 获取工具名称
func (b *BaseTool) GetName() string {
	return b.name
}

// GetDescription 获取工具描述
func (b *BaseTool) GetDescription() string {
	return b.description
}

// GetParameters 获取参数定义
func (b *BaseTool) GetParameters() map[string]ParameterSpec {
	return b.parameters
}

// AddParameter 添加参数定义
func (b *BaseTool) AddParameter(name string, spec ParameterSpec) {
	b.parameters[name] = spec
}

// ValidateParams 验证参数
func (b *BaseTool) ValidateParams(params map[string]interface{}) error {
	for name, spec := range b.parameters {
		if spec.Required {
			if _, exists := params[name]; !exists {
				return fmt.Errorf("missing required parameter: %s", name)
			}
		}
	}
	return nil
}
