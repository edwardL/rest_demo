package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	v1 "rest_demo/api/v1"
	"rest_demo/internal/service"
	"rest_demo/pkg/llm"
	"rest_demo/pkg/llm/workflow"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LLMHandler struct {
	llmService *service.LLMService
	llmConfig  *llm.Config
	logger     *zap.Logger
}

func NewLLMHandler(service *service.LLMService, cfg *llm.Config, logger *zap.Logger) *LLMHandler {
	return &LLMHandler{
		llmService: service,
		llmConfig:  cfg,
		logger:     logger,
	}
}

func (h *LLMHandler) Chat(c *gin.Context) {
	var req v1.ChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.Error(c, err)
		return
	}

	if req.Stream {
		h.streamChat(c, &req)
		return
	}

	messages := prepMessages(req.SystemMsg, req.Messages)
	tools := prepTools(req.Tools)
	opts := prepOptions(req.Options)

	content, think, err := h.llmService.ChatWithThink(c.Request.Context(), messages, tools, opts)
	if err != nil {
		v1.Error(c, err)
		return
	}

	v1.Success(c, v1.ChatRes{
		Content: content,
		Think:   think,
	})
}

func (h *LLMHandler) streamChat(c *gin.Context, req *v1.ChatReq) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	messages := prepMessages(req.SystemMsg, req.Messages)
	tools := prepTools(req.Tools)
	opts := prepOptions(req.Options)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		v1.Error(c, fmt.Errorf("不支持流式响应"))
		return
	}

	var lastFull string

	sseWriter := func(event string, data any) error {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, string(b))
		if err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	_ = sseWriter("start", map[string]string{"msg": "开始生成"})

	err := h.llmService.StreamCallback(ctx, messages, tools, opts, func(full string, chunk string, msgType llm.MessageType) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if msgType == llm.MessageTypeContent {
			lastFull = full
			_ = sseWriter("content", map[string]string{
				"delta": chunk,
				"full":  full,
				"type":  string(msgType),
			})
		} else if msgType == llm.MessageTypeThink {
			_ = sseWriter("think", map[string]string{
				"delta": chunk,
				"full":  full,
				"type":  string(msgType),
			})
		}
		return nil
	})

	if err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			_ = sseWriter("error", map[string]string{"msg": "请求已取消或超时"})
		} else {
			_ = sseWriter("error", map[string]string{"msg": err.Error()})
		}
		return
	}

	_ = sseWriter("done", map[string]string{
		"full": lastFull,
	})
}

func prepMessages(systemMsg string, msgs []llm.LLMMsg) []llm.LLMMsg {
	var result []llm.LLMMsg
	if systemMsg != "" {
		result = append(result, llm.LLMMsg{Role: "system", Content: systemMsg})
	}
	result = append(result, msgs...)
	return result
}

func prepTools(toolDefs []v1.ToolDef) []llm.Tool {
	tools := make([]llm.Tool, 0, len(toolDefs))
	for _, td := range toolDefs {
		if td.URL != "" {
			t := newHTTPTool(td.Name, td.Description, td.URL)
			tools = append(tools, t)
		}
	}
	return tools
}

func prepOptions(opts v1.ChatOptsReq) llm.ChatOptions {
	o := llm.ChatOptions{}
	if opts.Temperature != nil {
		o.Temperature = *opts.Temperature
	}
	if opts.MaxTokens != nil {
		o.MaxTokens = *opts.MaxTokens
	}
	if opts.TopP != nil {
		o.TopP = *opts.TopP
	}
	return o
}

type httpTool struct {
	name string
	desc string
	url  string
}

func newHTTPTool(name, desc, url string) *httpTool {
	return &httpTool{name: name, desc: desc, url: url}
}

func (t *httpTool) Name() string { return t.name }

func (t *httpTool) Schema() llm.ToolSchema {
	return llm.ToolSchema{
		Type:        "function",
		Description: t.desc,
		Parameters: llm.ToolParam{
			Type: "object",
			Properties: map[string]llm.ParamDef{
				"query": {Type: "string", Description: "查询内容"},
			},
			Required: []string{"query"},
		},
	}
}

func (t *httpTool) Handle(args string) (string, error) {
	var params map[string]string
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}
	query := params["query"]
	if query == "" {
		query = args
	}

	reqURL := strings.ReplaceAll(t.url, "{query}", query)
	resp, err := http.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}
	return string(body), nil
}

// ======== Workflow Handlers ========

func (h *LLMHandler) Workflow(c *gin.Context) {
	var req v1.WorkflowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.Error(c, err)
		return
	}

	runner := workflow.GetWorkflow(req.Topic)
	if runner == nil {
		v1.Error(c, fmt.Errorf("未找到工作流: %s", req.Topic))
		return
	}

	if req.Stream {
		h.streamWorkflow(c, &req, runner)
		return
	}

	messages := prepMessages(req.SystemMsg, req.Messages)
	opts := prepOptions(req.Options)

	store := workflow.NewStore()
	var allMsgs []workflow.SSEMessage
	var mu sync.Mutex

	nw := workflow.NewNotifyWriter(&memoryStore{}, 0, 0,
		func(event string, msg *workflow.SSEMessage) {
			mu.Lock()
			allMsgs = append(allMsgs, *msg)
			mu.Unlock()
		})

	client := llm.NewClient(h.llmConfig)
	sc := &workflow.StepCtx{
		Context:   c.Request.Context(),
		Store:     store,
		Messages:  messages,
		Opts:      opts,
		Notify:    nw.NotifyFunc(),
		LLMClient: client,
	}
	sc.CloseThink = nw.CloseThink

	if err := runner.Run(sc); err != nil {
		nw.Close()
		v1.Error(c, err)
		return
	}
	nw.Close()

	var content strings.Builder
	for _, m := range allMsgs {
		if m.MessageType == string(llm.MessageTypeContent) && !m.Done {
			content.WriteString(m.Message)
		}
	}

	v1.Success(c, v1.WorkflowRes{
		Content:  content.String(),
		Messages: allMsgs,
	})
}

func (h *LLMHandler) streamWorkflow(c *gin.Context, req *v1.WorkflowReq, runner workflow.Runner) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		v1.Error(c, fmt.Errorf("不支持流式响应"))
		return
	}

	sseWriter := func(event string, data any) error {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, string(b))
		if err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	_ = sseWriter("start", map[string]any{"topic": req.Topic})

	messages := prepMessages(req.SystemMsg, req.Messages)
	opts := prepOptions(req.Options)

	store := workflow.NewStore()
	nw := workflow.NewNotifyWriter(&memoryStore{}, 0, 0,
		func(event string, msg *workflow.SSEMessage) {
			_ = sseWriter(string(msg.MessageType), msg)
		})

	client := llm.NewClient(h.llmConfig)
	sc := &workflow.StepCtx{
		Context:   ctx,
		Store:     store,
		Messages:  messages,
		Opts:      opts,
		Notify:    nw.NotifyFunc(),
		LLMClient: client,
	}
	sc.CloseThink = nw.CloseThink
	store.Set("_session_id", int64(0))
	store.Set("_chat_id", int64(0))

	err := runner.Run(sc)
	nw.Close()

	if err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			_ = sseWriter("error", map[string]string{"msg": "请求已取消或超时"})
		} else {
			_ = sseWriter("error", map[string]string{"msg": err.Error()})
		}
		return
	}

	_ = sseWriter("done", map[string]string{"topic": req.Topic})
}

func (h *LLMHandler) GetTopics(c *gin.Context) {
	topics := workflow.GetAllTopics()
	var items []v1.TopicItem
	for _, t := range topics {
		items = append(items, v1.TopicItem{Key: t, Name: t})
	}
	v1.Success(c, v1.TopicsRes{Topics: items})
}

type memoryStore struct {
	mu       sync.Mutex
	messages []workflow.SessionMessage
	nextID   int64
}

func (s *memoryStore) Create(msg *workflow.SessionMessage) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	msg.ID = s.nextID
	s.messages = append(s.messages, *msg)
	return msg.ID, nil
}

func (s *memoryStore) Update(id int64, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.messages {
		if s.messages[i].ID == id {
			s.messages[i].Content = content
			return nil
		}
	}
	return fmt.Errorf("message not found: %d", id)
}
