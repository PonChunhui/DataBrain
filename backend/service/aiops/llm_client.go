package aiops

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"devops-backend/model"
)

type LLMResponse struct {
	Content      string
	TokensUsed   int
	Model        string
	FinishReason string
}

func NewLLMClient(config *model.LLMConfig) (llms.Model, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	opts := []openai.Option{
		openai.WithToken(config.APIKey),
		openai.WithModel(config.Model),
	}

	if config.BaseURL != "" {
		opts = append(opts, openai.WithBaseURL(config.BaseURL))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %v", err)
	}

	return llm, nil
}

func CallLLM(ctx context.Context, config *model.LLMConfig, prompt string) (*LLMResponse, error) {
	llm, err := NewLLMClient(config)
	if err != nil {
		return nil, err
	}

	resp, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt,
		llms.WithTemperature(config.Temperature),
		llms.WithMaxTokens(config.MaxTokens),
	)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %v", err)
	}

	return &LLMResponse{
		Content: resp,
		Model:   config.Model,
	}, nil
}

func CallLLMWithSystemPrompt(ctx context.Context, config *model.LLMConfig, systemPrompt, userPrompt string) (*LLMResponse, error) {
	llm, err := NewLLMClient(config)
	if err != nil {
		return nil, err
	}

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userPrompt),
	}

	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithTemperature(config.Temperature),
		llms.WithMaxTokens(config.MaxTokens),
	)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %v", err)
	}

	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Content
	}

	return &LLMResponse{
		Content: content,
		Model:   config.Model,
	}, nil
}

func StreamLLM(ctx context.Context, config *model.LLMConfig, messages []llms.MessageContent, callback func(string)) error {
	llm, err := NewLLMClient(config)
	if err != nil {
		return err
	}

	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithTemperature(config.Temperature),
		llms.WithMaxTokens(config.MaxTokens),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			callback(string(chunk))
			return nil
		}),
	)
	if err != nil {
		return fmt.Errorf("LLM stream failed: %v", err)
	}

	if len(resp.Choices) > 0 && resp.Choices[0].Content != "" {
		callback(resp.Choices[0].Content)
	}

	return nil
}
