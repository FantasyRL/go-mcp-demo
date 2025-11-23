package application

import (
	"context"
	"encoding/base64"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/internal/host/repository"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
)

func (h *Host) Chat(id int64, msg string, imageData []byte) (string, error) {
	return h.ChatWithConversation(id, "", msg, imageData)
}

func (h *Host) ChatWithConversation(id int64, conversationID string, msg string, imageData []byte) (string, error) {
	if config.AiProvider.Mode == constant.AiProviderModeRemote {
		return h.ChatOpenAI(id, msg, imageData)
	}

	var userHistory []ai_provider.Message
	if conversationID != "" {
		userHistory = h.historyStore.Get(conversationID)
	} else {
		userHistory = history[id]
	}
	if userHistory == nil {
		userHistory = []ai_provider.Message{}
	}

	userMsg := ai_provider.Message{Role: "user", Content: msg}
	if len(imageData) > 0 {
		userMsg.Images = []string{base64.StdEncoding.EncodeToString(imageData)}
	}
	userHistory = append(userHistory, userMsg)

	ollamaTools := h.mcpCli.ConvertToolsToOllama()
	ollamaOptions := ai_provider.BuildOptions()

	// 第一次调用
	resp, err := h.aiProviderCli.Chat(h.ctx, ai_provider.ChatRequest{
		Model:     config.AiProvider.Model,
		Messages:  userHistory,
		Options:   ollamaOptions,
		Tools:     ollamaTools,
		KeepAlive: config.AiProvider.Options.KeepAlive,
	})
	if err != nil {
		return "", err
	}
	userHistory = append(userHistory, ai_provider.Message{Role: "assistant", Content: resp.Message.Content})

	// 工具调用路径
	if len(resp.Message.ToolCalls) > 0 {
		for _, c := range resp.Message.ToolCalls {
			args, parseErr := ai_provider.ParseToolArguments(c.Function.Arguments)
			if parseErr != nil {
				args = map[string]any{"_error": parseErr.Error()}
			}
			out, callErr := h.mcpCli.CallTool(context.Background(), c.Function.Name, args)
			if callErr != nil {
				out = "tool error: " + callErr.Error()
			}
			userHistory = append(userHistory, ai_provider.Message{
				Role:     "tool",
				ToolName: c.Function.Name,
				Content:  out,
			})
			logger.Infof("[tool] %s executed", c.Function.Name)
		}

		resp2, err2 := h.aiProviderCli.Chat(h.ctx, ai_provider.ChatRequest{
			Model:    config.AiProvider.Model,
			Messages: userHistory,
			Options:  ollamaOptions,
			Tools:    ollamaTools,
		})
		if err2 != nil {
			return "", err2
		}
		userHistory = append(userHistory, ai_provider.Message{Role: "assistant", Content: resp2.Message.Content})
		h.saveHistory(conversationID, id, userHistory)
		return resp2.Message.Content, nil
	}

	// 无工具调用
	h.saveHistory(conversationID, id, userHistory)
	return resp.Message.Content, nil
}

func (h *Host) saveHistory(conversationID string, userID int64, msgs []ai_provider.Message) {
	if conversationID == "" {
		history[userID] = msgs
		return
	}
	h.historyStore.Set(conversationID, msgs)
	dto := make([]repository.AIMessageDTO, 0, len(msgs))
	for _, m := range msgs {
		dto = append(dto, repository.AIMessageDTO{
			Role:     m.Role,
			Content:  m.Content,
			ToolName: m.ToolName,
			Images:   m.Images,
		})
	}
	if err := h.conversationRepo.UpsertHistory(context.Background(), conversationID, userID, dto); err != nil {
		logger.Errorf("persist conversation %s failed: %v", conversationID, err)
	}
}

// 旧流式入口：无会话ID
func (h *Host) StreamChat(
	ctx context.Context,
	id int64,
	userMsg string,
	emit func(event string, v any) error,
) error {
	return h.StreamChatWithConversation(ctx, id, "", userMsg, emit)
}

// 新流式封装：支持会话ID（当前不持久化流式历史，保持原策略）
func (h *Host) StreamChatWithConversation(
	ctx context.Context,
	id int64,
	conversationID string,
	userMsg string,
	emit func(event string, v any) error,
) error {
	hist := func() []ai_provider.Message {
		if conversationID != "" {
			return h.historyStore.Get(conversationID)
		}
		return history[id]
	}()
	if hist == nil {
		hist = []ai_provider.Message{}
	}
	hist = append(hist, ai_provider.Message{Role: "user", Content: userMsg})

	tools := h.mcpCli.ConvertToolsToOllama()
	opts := ai_provider.BuildOptions()
	var assistantBuf string
	var toolCalls []ai_provider.ToolCall

	// 首次流式
	err := h.aiProviderCli.ChatStream(ctx, ai_provider.ChatRequest{
		Model:     config.AiProvider.Model,
		Messages:  hist,
		Tools:     tools,
		Options:   opts,
		KeepAlive: config.AiProvider.Options.KeepAlive,
	}, func(chunk *ai_provider.ChatResponse) error {
		if s := chunk.Message.Content; s != "" {
			assistantBuf += s
			_ = emit(constant.SSEEventDelta, map[string]any{"text": s})
		}
		if len(chunk.Message.ToolCalls) > 0 {
			toolCalls = append(toolCalls, chunk.Message.ToolCalls...)
			_ = emit(constant.SSEEventStartToolCall, map[string]any{"tool_calls": chunk.Message.ToolCalls})
			return errno.OllamaInternalStopStream
		}
		return nil
	})
	if err != nil {
		return err
	}
	if assistantBuf != "" {
		hist = append(hist, ai_provider.Message{Role: "assistant", Content: assistantBuf})
	}
	if len(toolCalls) == 0 {
		_ = emit(constant.SSEEventDone, map[string]any{"reason": "no_tool"})
		return nil
	}

	// 执行工具
	for _, tc := range toolCalls {
		args, parseErr := ai_provider.ParseToolArguments(tc.Function.Arguments)
		if parseErr != nil {
			args = map[string]any{"_error": parseErr.Error()}
		}
		_ = emit(constant.SSEEventToolCall, map[string]any{"name": tc.Function.Name, "args": args})
		out, callErr := h.mcpCli.CallTool(ctx, tc.Function.Name, args)
		if callErr != nil {
			out = "tool error: " + callErr.Error()
		}
		_ = emit(constant.SSEEventToolResult, map[string]any{"name": tc.Function.Name, "result": out})
		hist = append(hist, ai_provider.Message{Role: "tool", ToolName: tc.Function.Name, Content: out})
	}

	// 二次流式
	var finalBuf string
	err = h.aiProviderCli.ChatStream(ctx, ai_provider.ChatRequest{
		Model:     config.AiProvider.Model,
		Messages:  hist,
		Tools:     tools,
		Options:   opts,
		KeepAlive: config.AiProvider.Options.KeepAlive,
	}, func(chunk *ai_provider.ChatResponse) error {
		if s := chunk.Message.Content; s != "" {
			finalBuf += s
			_ = emit(constant.SSEEventDelta, map[string]any{"text": s})
		}
		return nil
	})
	if err != nil {
		return err
	}
	if finalBuf != "" {
		hist = append(hist, ai_provider.Message{Role: "assistant", Content: finalBuf})
	}
	_ = emit(constant.SSEEventDone, map[string]any{"reason": "completed"})
	return nil
}
