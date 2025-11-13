package host

import (
	"context"
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/mcp_client"
	"github.com/openai/openai-go/v2"
)

// 简单的内存存储用户对话历史
var history = make(map[int64][]ai_provider.Message)
var historyOpenAI = make(map[int64][]openai.ChatCompletionMessageParamUnion)

type Host struct {
	ctx           context.Context
	mcpCli        mcp_client.ToolClient
	aiProviderCli *ai_provider.Client
}

func NewHost(ctx context.Context, clientSet *base.ClientSet) *Host {
	return &Host{
		ctx:           ctx,
		mcpCli:        clientSet.MCPCli,
		aiProviderCli: clientSet.AiProviderCli,
	}
}
