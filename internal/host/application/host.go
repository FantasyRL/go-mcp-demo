package application

import (
	"context"
	"errors"

	"github.com/FantasyRL/go-mcp-demo/internal/host/infra"
	"github.com/FantasyRL/go-mcp-demo/internal/host/repository"
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/db"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/mcp_client"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/query"
	"github.com/openai/openai-go/v2"
)

var history = make(map[int64][]ai_provider.Message)
var historyOpenAI = make(map[int64][]openai.ChatCompletionMessageParamUnion)

type Host struct {
	ctx                context.Context
	mcpCli             mcp_client.ToolClient
	aiProviderCli      *ai_provider.Client
	templateRepository repository.TemplateRepository

	// 会话持久化与内存缓存
	conversationRepo repository.ConversationRepository
	historyStore     *MemoryHistoryStore
}

func NewHost(ctx context.Context, clientSet *base.ClientSet) *Host {
	return &Host{
		ctx:                ctx,
		mcpCli:             clientSet.MCPCli,
		aiProviderCli:      clientSet.AiProviderCli,
		templateRepository: infra.NewTemplateRepository(db.NewDBWithQuery(clientSet.ActualDB, query.Use)),
		conversationRepo:   repository.NewConversationRepository(clientSet.ActualDB),
		historyStore:       NewMemoryHistoryStore(),
	}
}

func (h *Host) SummarizeConversation(conversationID string) (*SummarizeResult, error) {
	if h == nil {
		return nil, errors.New("host is nil")
	}
	if conversationID == "" {
		return nil, errors.New("conversation_id is required")
	}
	return h.summarizeConversation(conversationID)
}

func (h *Host) ConversationRepository() repository.ConversationRepository {
	return h.conversationRepo
}
