package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	headerDataPrefix = []byte("data: ")
	errorPrefixJSON  = []byte(`{"error":`)
	bbDataDone       = []byte("[DONE]")
	clientMu         sync.Mutex
	clientCache      *http.Client
)

func getHTTPClient(cfg *Config, opts ChatOptions) *http.Client {
	if opts.HTTPClient() != nil {
		return opts.HTTPClient()
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		return http.DefaultClient
	}
	clientMu.Lock()
	defer clientMu.Unlock()
	if clientCache != nil {
		return clientCache
	}
	cloned := http.DefaultTransport.(*http.Transport).Clone()
	cloned.ResponseHeaderTimeout = timeout
	clientCache = &http.Client{Transport: cloned}
	return clientCache
}

func effectiveEndpoint(cfg *Config, opts ChatOptions) string {
	if opts.Endpoint != "" {
		return opts.Endpoint
	}
	if cfg.Endpoint != "" {
		return cfg.Endpoint
	}
	if v := os.Getenv("LLM_ENDPOINT"); v != "" {
		return v
	}
	return "https://api.deepseek.com/chat/completions"
}

func effectiveAPIKey(cfg *Config, opts ChatOptions) string {
	if opts.APIKey != "" {
		return opts.APIKey
	}
	if cfg.APIKey != "" {
		return cfg.APIKey
	}
	if v := os.Getenv("LLM_API_KEY"); v != "" {
		return v
	}
	return os.Getenv("LLM_KEY")
}

func effectiveModel(cfg *Config) string {
	if cfg.Model != "" {
		return cfg.Model
	}
	if v := os.Getenv("LLM_MODEL"); v != "" {
		return v
	}
	return "deepseek-chat"
}

// ======== SSE 流式读取器 ========

type chatStreamReader struct {
	finished bool
	reader   *bufio.Reader
	resp     *http.Response
	errAcc   bytes.Buffer
}

func newChatStreamReader(resp *http.Response) *chatStreamReader {
	return &chatStreamReader{reader: bufio.NewReader(resp.Body), resp: resp}
}

func (r *chatStreamReader) Recv() (*llmStreamChunk, error) {
	raw, err := r.recvRaw()
	if err != nil {
		return nil, err
	}
	var chunk llmStreamChunk
	if err := json.Unmarshal(raw, &chunk); err != nil {
		return nil, fmt.Errorf("解析chunk失败: %w", err)
	}
	return &chunk, nil
}

func (r *chatStreamReader) recvRaw() ([]byte, error) {
	if r.finished {
		return nil, io.EOF
	}
	var hasError bool
	for {
		rawLine, readErr := r.reader.ReadBytes('\n')
		if readErr != nil {
			if len(rawLine) > 0 && readErr == io.EOF && !hasError {
				noSpace := bytes.TrimSpace(rawLine)
				noPrefix := bytes.TrimPrefix(noSpace, headerDataPrefix)
				if bytes.HasPrefix(noPrefix, errorPrefixJSON) {
					hasError = true
				}
				if bytes.HasPrefix(noSpace, headerDataPrefix) {
					if bytes.Equal(noPrefix, bbDataDone) {
						r.finished = true
						return nil, io.EOF
					}
					return bytes.Clone(noPrefix), nil
				}
			}
			if err := r.unmarshalError(); err != nil {
				return nil, err
			}
			if readErr == io.EOF {
				return nil, io.EOF
			}
			return nil, readErr
		}
		if hasError {
			r.errAcc.Write(rawLine)
			continue
		}
		noSpace := bytes.TrimSpace(rawLine)
		noPrefix := bytes.TrimPrefix(noSpace, headerDataPrefix)
		if bytes.HasPrefix(noPrefix, errorPrefixJSON) {
			hasError = true
		}
		if !bytes.HasPrefix(noSpace, headerDataPrefix) || hasError {
			if hasError {
				r.errAcc.Write(noPrefix)
			}
			continue
		}
		if bytes.Equal(noPrefix, bbDataDone) {
			r.finished = true
			return nil, io.EOF
		}
		return bytes.Clone(noPrefix), nil
	}
}

func (r *chatStreamReader) unmarshalError() error {
	errBytes := r.errAcc.Bytes()
	r.errAcc.Reset()
	if len(errBytes) == 0 {
		return nil
	}
	var errResp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(errBytes, &errResp) != nil {
		return nil
	}
	if errResp.Error.Message != "" {
		return fmt.Errorf("LLM API错误: %s", errResp.Error.Message)
	}
	return nil
}

func (r *chatStreamReader) Close() error { return r.resp.Body.Close() }

// ======== 高层 API (无配置, 需要传入 cfg) ========

type Client struct {
	cfg    *Config
	client *http.Client
}

func NewClient(cfg *Config) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) httpClient() *http.Client {
	if c.client != nil {
		return c.client
	}
	c.client = getHTTPClient(c.cfg, ChatOptions{})
	return c.client
}

// Chat 非流式对话。支持 tool call 自动循环和 response_format。
func (c *Client) Chat(ctx context.Context, messages []LLMMsg, tools []Tool, opts ChatOptions) (string, error) {
	content, _, err := c.ChatWithThink(ctx, messages, tools, opts)
	return content, err
}

// ChatWithThink 非流式对话，分别返回正文和思考内容。
func (c *Client) ChatWithThink(ctx context.Context, messages []LLMMsg, tools []Tool, opts ChatOptions) (string, string, error) {
	var contentBuf, thinkBuf strings.Builder
	err := c.StreamCallback(ctx, messages, tools, opts, func(full string, chunk string, msgType MessageType) error {
		switch msgType {
		case MessageTypeContent:
			contentBuf.WriteString(chunk)
		case MessageTypeThink:
			thinkBuf.WriteString(chunk)
		}
		return nil
	})
	return contentBuf.String(), thinkBuf.String(), err
}

// StreamCallback 高级流式对话。
// cb 的回调参数: full=累计全文, chunk=新增片段, msgType=正文(MessageTypeContent)/思考(MessageTypeThink)
// cb 返回 error 可终止流式。
func (c *Client) StreamCallback(ctx context.Context, messages []LLMMsg, tools []Tool, opts ChatOptions, cb func(full string, chunk string, msgType MessageType) error) error {
	if len(messages) == 0 {
		return fmt.Errorf("消息列表不能为空")
	}
	var fullText strings.Builder
	var fullThink strings.Builder
	byName := make(map[string]Tool, len(tools))
	for _, t := range tools {
		byName[t.Name()] = t
	}

	for {
		toolCalled := false
		err := c.llmChatStream(ctx, messages, tools, opts, func(chunk string) {
			fullText.WriteString(chunk)
			if cb != nil {
				cb(fullText.String(), chunk, MessageTypeContent)
			}
		}, func(chunk string) {
			fullThink.WriteString(chunk)
			if cb != nil {
				cb(fullThink.String(), chunk, MessageTypeThink)
			}
		}, func(name string, args string) {
			if name != "" {
				toolCalled = true
			}
		})
		if err != nil {
			return err
		}
		if !toolCalled {
			return nil
		}
		toolCalls, err := c.extractToolCalls(ctx, messages, tools, opts)
		if err != nil || len(toolCalls) == 0 {
			return fmt.Errorf("提取工具调用失败: %w", err)
		}
		messages = append(messages, LLMMsg{Role: "assistant", ToolCalls: toolCalls})
		for _, tc := range toolCalls {
			t := byName[tc.Function.Name]
			if t == nil {
				return fmt.Errorf("LLM请求工具 [%s] 但未注册", tc.Function.Name)
			}
			result, tErr := t.Handle(tc.Function.Arguments)
			if tErr != nil {
				result = fmt.Sprintf("工具执行失败: %s", tErr.Error())
			}
			messages = append(messages, LLMMsg{Role: "tool", Content: result, ToolCallID: tc.ID})
		}
		fullText.Reset()
		fullThink.Reset()
	}
}

// StreamChat 流式对话，通过回调持续输出（不累积, 不处理 think）。
// onContent: 每次收到新文本片段时回调（传入增量片段和累计全文）。
// onDone: 流式结束时回调，传入最终全文。
func (c *Client) StreamChat(ctx context.Context, messages []LLMMsg, tools []Tool, opts ChatOptions, onContent func(delta string, full string), onDone func(full string)) error {
	var fullText strings.Builder
	return c.StreamCallback(ctx, messages, tools, opts, func(full string, chunk string, msgType MessageType) error {
		if msgType == MessageTypeContent {
			fullText.WriteString(chunk)
			if onContent != nil {
				onContent(chunk, full)
			}
		}
		return nil
	})
}

// ChatOnce 单次调用 ChatGPT 流式 API, 不带 tool loop。
// onContent/onThink: 流式文本回调; onDone: 流结束时回调。
func (c *Client) ChatOnce(ctx context.Context, messages []LLMMsg, tools []Tool, opts ChatOptions, onContent func(delta string, full string), onThink func(delta string, full string), onDone func(full string)) error {
	var fullText strings.Builder
	var fullThink strings.Builder
	err := c.llmChatStream(ctx, messages, tools, opts, func(chunk string) {
		fullText.WriteString(chunk)
		if onContent != nil {
			onContent(chunk, fullText.String())
		}
	}, func(chunk string) {
		fullThink.WriteString(chunk)
		if onThink != nil {
			onThink(chunk, fullThink.String())
		}
	}, nil)
	if onDone != nil {
		onDone(fullText.String())
	}
	return err
}

// ======== 底层调用 ========

func (c *Client) llmChatStream(ctx context.Context, messages []LLMMsg, tools []Tool, opts ChatOptions, onContent func(chunk string), onReasoning func(chunk string), onToolCall func(name string, args string)) error {
	endpoint := effectiveEndpoint(c.cfg, opts)
	key := effectiveAPIKey(c.cfg, opts)
	model := effectiveModel(c.cfg)

	payload := map[string]any{"model": model, "messages": messages, "stream": true}
	opts.Apply(payload)
	if len(tools) > 0 {
		var apiTools []map[string]any
		for _, t := range tools {
			apiTools = append(apiTools, toAPITool(t))
		}
		payload["tools"] = apiTools
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LLM状态码 %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	reader := newChatStreamReader(resp)
	for {
		chunk, recvErr := reader.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				break
			}
			reader.Close()
			return recvErr
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" && onContent != nil {
				onContent(choice.Delta.Content)
			}
			if choice.Delta.ReasoningContent != "" && onReasoning != nil {
				onReasoning(choice.Delta.ReasoningContent)
			}
			for _, tc := range choice.Delta.ToolCalls {
				if onToolCall != nil {
					onToolCall(tc.Function.Name, tc.Function.Arguments)
				}
			}
		}
	}
	return reader.Close()
}

func (c *Client) extractToolCalls(ctx context.Context, messages []LLMMsg, tools []Tool, opts ChatOptions) ([]ToolCall, error) {
	endpoint := effectiveEndpoint(c.cfg, opts)
	key := effectiveAPIKey(c.cfg, opts)
	model := effectiveModel(c.cfg)

	payload := map[string]any{"model": model, "messages": messages}
	opts.Apply(payload)
	if len(tools) > 0 {
		var apiTools []map[string]any
		for _, t := range tools {
			apiTools = append(apiTools, toAPITool(t))
		}
		payload["tools"] = apiTools
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM状态码 %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	var result struct {
		Choices []struct {
			Message struct {
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("LLM未返回有效响应")
	}
	return result.Choices[0].Message.ToolCalls, nil
}

// ChatOptions helper to return optional custom HTTP client
func (o ChatOptions) HTTPClient() *http.Client { return nil }
