package llm

import (
	"encoding/json"
	"fmt"
)

type Tool interface {
	Name() string
	Schema() ToolSchema
	Handle(args string) (string, error)
}

type ToolSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

type ToolParam struct {
	Type       string              `json:"type,omitempty"`
	Properties map[string]ParamDef `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type ParamDef struct {
	Type        string              `json:"type,omitempty"`
	Description string              `json:"description,omitempty"`
	Enum        []string            `json:"enum,omitempty"`
	Items       *ParamDef           `json:"items,omitempty"`
	Properties  map[string]ParamDef `json:"properties,omitempty"`
	Required    []string            `json:"required,omitempty"`
}

type genericTool[T1 any, T2 any] struct {
	name   string
	schema ToolSchema
	handle func(T1) (T2, error)
}

func (gt *genericTool[T1, T2]) Name() string       { return gt.name }
func (gt *genericTool[T1, T2]) Schema() ToolSchema { return gt.schema }
func (gt *genericTool[T1, T2]) Handle(args string) (string, error) {
	var in T1
	if _, ok := any(in).(string); ok {
		in = any(args).(T1)
	} else if err := json.Unmarshal([]byte(args), &in); err != nil {
		return "", fmt.Errorf("工具参数解析失败: %w", err)
	}
	out, err := gt.handle(in)
	if err != nil {
		return "", err
	}
	if s, ok := any(out).(string); ok {
		return s, nil
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("工具结果序列化失败: %w", err)
	}
	return string(b), nil
}

func NewTool[T1 any, T2 any](name, desc string, schema ToolSchema, handler func(T1) (T2, error)) *genericTool[T1, T2] {
	if schema.Description == "" {
		schema.Description = desc
	}
	return &genericTool[T1, T2]{name: name, schema: schema, handle: handler}
}

func toAPITool(t Tool) map[string]any {
	s := t.Schema()
	s.Name = t.Name()
	if s.Type == "" {
		s.Type = "function"
	}
	return map[string]any{"type": s.Type, "function": s}
}

type trackingTool struct {
	inner  Tool
	called *bool
}

func (t *trackingTool) Name() string       { return t.inner.Name() }
func (t *trackingTool) Schema() ToolSchema { return t.inner.Schema() }
func (t *trackingTool) Handle(args string) (string, error) {
	*t.called = true
	return t.inner.Handle(args)
}

func WrapTools(tools []Tool) (wrapped []Tool, called func(name string) bool) {
	flags := make(map[string]*bool, len(tools))
	wrapped = make([]Tool, len(tools))
	for i, t := range tools {
		f := new(bool)
		flags[t.Name()] = f
		wrapped[i] = &trackingTool{inner: t, called: f}
	}
	return wrapped, func(name string) bool {
		if f, ok := flags[name]; ok {
			return *f
		}
		return false
	}
}
