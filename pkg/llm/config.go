package llm

import "time"

type Config struct {
	Endpoint string        `help:"LLM API端点" default:"https://api.deepseek.com/chat/completions"`
	APIKey   string        `help:"LLM API密钥" default:""`
	Model    string        `help:"LLM 模型名称" default:"deepseek-chat"`
	Timeout  time.Duration `help:"LLM 请求超时" default:"60s"`
}
