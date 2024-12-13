package tools

import (
	"context"
	"fmt"
)

// Calculator 计算器工具
type Calculator struct {
	*BaseTool
}

func NewCalculator() *Calculator {
	tool := &Calculator{
		BaseTool: NewBaseTool("calculator", "执行基础的数学计算"),
	}

	// 添加参数规格
	tool.AddParameter("a", ParameterSpec{
		Type:        "number",
		Description: "第一个数字",
		Required:    true,
	})
	tool.AddParameter("b", ParameterSpec{
		Type:        "number",
		Description: "第二个数字",
		Required:    true,
	})
	tool.AddParameter("operation", ParameterSpec{
		Type:        "string",
		Description: "运算类型 (add/subtract/multiply/divide)",
		Required:    true,
		Enum:        []string{"add", "subtract", "multiply", "divide"},
	})

	return tool
}

func (c *Calculator) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// 参数提取和类型转换
	a, ok := params["a"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid parameter 'a': must be a number")
	}

	b, ok := params["b"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid parameter 'b': must be a number")
	}

	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid parameter 'operation': must be a string")
	}

	// 执行计算
	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	return result, nil
}
