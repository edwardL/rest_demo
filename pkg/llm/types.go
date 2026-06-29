package llm

type MessageType string

const (
	MessageTypeContent MessageType = "content"
	MessageTypeThink   MessageType = "think"
)

type LLMMsg struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolCallFunc `json:"function"`
}

type ToolCallFunc struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type llmStreamChunk struct {
	Choices []struct {
		Delta struct {
			Role             string     `json:"role,omitempty"`
			Content          string     `json:"content,omitempty"`
			ReasoningContent string     `json:"reasoning_content,omitempty"`
			ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

type ResponseFormatType string

const (
	ResponseFormatTypeText       ResponseFormatType = "text"
	ResponseFormatTypeJSONObject ResponseFormatType = "json_object"
	ResponseFormatTypeJSONSchema ResponseFormatType = "json_schema"
)

type ResponseFormat struct {
	Type       ResponseFormatType        `json:"type"`
	JSONSchema *ResponseFormatJSONSchema `json:"json_schema,omitempty"`
}

type ResponseFormatJSONSchema struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Schema      any    `json:"schema"`
	Strict      bool   `json:"strict,omitempty"`
}

type ChatOptions struct {
	Temperature    float32
	MaxTokens      int
	TopP           float32
	StopWords      []string
	ResponseFormat *ResponseFormat
	ExtraFields    map[string]any
	Endpoint       string
	APIKey         string
}

func (o ChatOptions) Apply(payload map[string]any) {
	if o.Temperature != 0 {
		payload["temperature"] = o.Temperature
	}
	if o.MaxTokens != 0 {
		payload["max_tokens"] = o.MaxTokens
	}
	if o.TopP != 0 {
		payload["top_p"] = o.TopP
	}
	if len(o.StopWords) > 0 {
		payload["stop"] = o.StopWords
	}
	if o.ResponseFormat != nil {
		payload["response_format"] = o.ResponseFormat
	}
	for k, v := range o.ExtraFields {
		payload[k] = v
	}
}
