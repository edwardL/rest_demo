package service

import (
	"context"
	"rest_demo/pkg/llm"

	"go.uber.org/zap"
)

type LLMService struct {
	client *llm.Client
	logger *zap.Logger
}

func NewLLMService(cfg *llm.Config, logger *zap.Logger) *LLMService {
	return &LLMService{
		client: llm.NewClient(cfg),
		logger: logger,
	}
}

func (s *LLMService) Chat(ctx context.Context, messages []llm.LLMMsg, tools []llm.Tool, opts llm.ChatOptions) (string, error) {
	return s.client.Chat(ctx, messages, tools, opts)
}

func (s *LLMService) ChatWithThink(ctx context.Context, messages []llm.LLMMsg, tools []llm.Tool, opts llm.ChatOptions) (string, string, error) {
	return s.client.ChatWithThink(ctx, messages, tools, opts)
}

func (s *LLMService) StreamCallback(ctx context.Context, messages []llm.LLMMsg, tools []llm.Tool, opts llm.ChatOptions, cb func(full string, chunk string, msgType llm.MessageType) error) error {
	return s.client.StreamCallback(ctx, messages, tools, opts, cb)
}

func (s *LLMService) ChatOnce(ctx context.Context, messages []llm.LLMMsg, tools []llm.Tool, opts llm.ChatOptions, onContent func(delta string, full string), onThink func(delta string, full string), onDone func(full string)) error {
	return s.client.ChatOnce(ctx, messages, tools, opts, onContent, onThink, onDone)
}
