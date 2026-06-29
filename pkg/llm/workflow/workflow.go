package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"rest_demo/pkg/llm"
)

const defaultMaxRetry = 10

const (
	keyAgentTools = "_agent_tools"
	keyViewTag    = "_view_tag"
	keySessionID  = "_session_id"
	keyChatID     = "_chat_id"
)

type NotifyFunc func(full, chunk string, msgType llm.MessageType, files []FileItem) error

// ======== Store ========

type Store struct {
	mu   sync.RWMutex
	data map[string]any
}

func NewStore() *Store {
	return &Store{data: make(map[string]any)}
}

func (s *Store) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

func GetFromStore[T any](s *Store, key string) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	if !ok {
		var zero T
		return zero, false
	}
	typed, ok := v.(T)
	return typed, ok
}

// ======== StepCtx ========

type StepCtx struct {
	Context    context.Context
	Store      *Store
	Notify     NotifyFunc
	CloseThink func()
	Messages   []llm.LLMMsg
	Opts       llm.ChatOptions
	StepIndex  int
	LLMClient  *llm.Client
}

func (sc *StepCtx) AddMessage(role, content string) {
	sc.Messages = append(sc.Messages, llm.LLMMsg{Role: role, Content: content})
}

// ======== Step ========

type Step interface {
	Execute(sc *StepCtx) error
}

// ======== Runner ========

type Runner interface {
	Run(sc *StepCtx) error
}

type SingleFlow []Step

func (sf SingleFlow) Run(sc *StepCtx) error {
	for i, st := range sf {
		sc.StepIndex = i
		if err := st.Execute(sc); err != nil {
			return fmt.Errorf("步骤%d失败: %w", i, err)
		}
	}
	return nil
}

// ======== LLMStep ========

type LLMStep struct {
	systemPrompt string
	exitTool     string
	maxRetry     int
	promptKey    string
	chatOpts     *llm.ChatOptions
	onlyLastUser bool
	noExit       bool
	silent       bool
}

func NewLLMStep(systemPrompt string, maxRetry int) *LLMStep {
	if maxRetry <= 0 {
		maxRetry = defaultMaxRetry
	}
	return &LLMStep{systemPrompt: systemPrompt, maxRetry: maxRetry}
}

func (s *LLMStep) WithExitTool(name string) *LLMStep          { s.exitTool = name; return s }
func (s *LLMStep) WithPromptKey(key string) *LLMStep          { s.promptKey = key; return s }
func (s *LLMStep) WithOptions(opts *llm.ChatOptions) *LLMStep { s.chatOpts = opts; return s }
func (s *LLMStep) WithOnlyLastUser() *LLMStep                 { s.onlyLastUser = true; return s }
func (s *LLMStep) WithoutExitTool() *LLMStep                  { s.noExit = true; return s }
func (s *LLMStep) WithSilent() *LLMStep                       { s.silent = true; return s }

func (s *LLMStep) Execute(sc *StepCtx) error {
	prompt := s.systemPrompt
	if s.promptKey != "" {
		if raw, ok := sc.Store.Get(s.promptKey); ok {
			if p, ok2 := raw.(string); ok2 && p != "" {
				prompt = p
			}
		}
	}

	for retry := 0; retry < s.maxRetry; retry++ {
		var allTools []llm.Tool
		exit := &StepExitTool{}
		if !s.noExit {
			allTools = append(allTools, exit)
		}
		if raw, ok := sc.Store.Get(keyAgentTools); ok {
			if tl, ok := raw.([]llm.Tool); ok {
				allTools = append(allTools, tl...)
			}
		}

		var stepMsgs []llm.LLMMsg
		if s.onlyLastUser {
			for i := len(sc.Messages) - 1; i >= 0; i-- {
				if sc.Messages[i].Role == "user" && sc.Messages[i].Content != "" {
					stepMsgs = []llm.LLMMsg{
						{Role: "system", Content: prompt},
						{Role: "user", Content: sc.Messages[i].Content},
					}
					break
				}
			}
		}
		if stepMsgs == nil {
			stepMsgs = append([]llm.LLMMsg{{Role: "system", Content: prompt}}, sc.Messages...)
		}

		opts := sc.Opts
		if s.chatOpts != nil {
			opts = *s.chatOpts
		}

		var noExitContent strings.Builder
		err := sc.LLMClient.StreamCallback(sc.Context, stepMsgs, allTools, opts,
			func(full string, chunk string, msgType llm.MessageType) error {
				if (s.noExit || s.silent) && msgType == llm.MessageTypeContent {
					noExitContent.WriteString(chunk)
				}
				if s.silent {
					return sc.Notify(full, chunk, llm.MessageTypeContent, nil)
				}
				return sc.Notify(full, chunk, msgType, nil)
			})

		if sc.CloseThink != nil && !s.silent {
			sc.CloseThink()
		}

		if err != nil {
			if sc.Context.Err() != nil {
				return err
			}
			continue
		}

		if !s.noExit && exit.Called {
			answer, _ := sc.LLMClient.Chat(sc.Context, stepMsgs, allTools, opts)
			if answer != "" {
				sc.Messages = append(sc.Messages, llm.LLMMsg{Role: "assistant", Content: answer})
				if !s.silent {
					sc.Notify(answer, answer, llm.MessageTypeContent, nil)
				}
			}
			return nil
		}

		if s.noExit {
			if noExitContent.Len() > 0 {
				sc.Messages = append(sc.Messages, llm.LLMMsg{Role: "assistant", Content: noExitContent.String()})
			}
			return nil
		}

		remind := "请完成任务目标后调用 step_complete 工具。"
		if s.exitTool != "" {
			remind = "请完成任务目标后调用 " + s.exitTool + " 工具。"
		}
		sc.Messages = append(sc.Messages, llm.LLMMsg{Role: "user", Content: remind})
	}
	return nil
}

// ======== StepExitTool ========

type StepExitTool struct {
	Called bool
}

func (t *StepExitTool) Name() string { return "step_complete" }

func (t *StepExitTool) Schema() llm.ToolSchema {
	return llm.ToolSchema{Type: "function", Description: "标记当前步骤已完成，允许进入下一步骤。"}
}

func (t *StepExitTool) Handle(args string) (string, error) {
	t.Called = true
	return "步骤完成确认", nil
}

// ======== ActionFn / ActionStep ========

type ActionFn func(sc *StepCtx) error

type actionStep struct {
	name string
	fn   ActionFn
}

func NewActionStep(name string, fn ActionFn) *actionStep {
	return &actionStep{name: name, fn: fn}
}

func (s *actionStep) Execute(sc *StepCtx) error {
	return s.fn(sc)
}

// ======== Registry ========

var registry = map[string]Runner{
	"analyze": SingleFlow{
		NewLLMStep("你是一个安全分析专家，请分析用户提供的信息，识别潜在的安全威胁。", 3),
		NewLLMStep("基于上一步的分析结果，给出具体的安全建议和补救措施。", 3),
		NewLLMStep("总结以上分析和建议，生成最终的安全分析报告。", 3),
	},
	"code-review": SingleFlow{
		NewLLMStep("你是一个代码审查专家，请审查用户提供的代码，找出潜在的问题。", 3),
		NewLLMStep("基于上一步发现的问题，提供具体的改进建议和代码示例。", 3),
	},
}

func Register(topic string, runner Runner) {
	registry[topic] = runner
}

func GetWorkflow(topic string) Runner {
	return registry[topic]
}

func GetAllTopics() []string {
	keys := make([]string, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}
	return keys
}
