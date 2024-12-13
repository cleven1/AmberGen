package config

import (
	"encoding/json"
	"os"
)

// Config 配置结构
type Config struct {
	APIKey  string `json:"api_key"`  // OpenAI API密钥
	BaseURL string `json:"base_url"` // API基础URL
	Model   string `json:"model"`    // 模型
}

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
