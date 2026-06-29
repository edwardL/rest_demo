package v1

import (
	"rest_demo/pkg/llm"
	"rest_demo/pkg/llm/workflow"
)

type ChatReq struct {
	Messages  []llm.LLMMsg `json:"messages" binding:"required"`
	Tools     []ToolDef    `json:"tools,omitempty"`
	Stream    bool         `json:"stream,omitempty"`
	SystemMsg string       `json:"system_msg,omitempty"`
	Options   ChatOptsReq  `json:"options,omitempty"`
}

type ToolDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	URL         string `json:"url,omitempty"`
}

type ChatOptsReq struct {
	Temperature *float32 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
	TopP        *float32 `json:"top_p,omitempty"`
}

type ChatRes struct {
	Content  string    `json:"content"`
	Think    string    `json:"think,omitempty"`
	Segments []SegItem `json:"segments,omitempty"`
}

type SegItem struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type WorkflowReq struct {
	Topic     string       `json:"topic" binding:"required"`
	Messages  []llm.LLMMsg `json:"messages" binding:"required"`
	Stream    bool         `json:"stream,omitempty"`
	SystemMsg string       `json:"system_msg,omitempty"`
	Options   ChatOptsReq  `json:"options,omitempty"`
}

type TopicItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type TopicsRes struct {
	Topics []TopicItem `json:"topics"`
}

type WorkflowRes struct {
	Content  string                `json:"content"`
	Think    string                `json:"think,omitempty"`
	Messages []workflow.SSEMessage `json:"messages,omitempty"`
}
