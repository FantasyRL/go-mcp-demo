package repository

import (
	"context"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
	"github.com/openai/openai-go/v2"
)

// TemplateRepository 根据实际需求定义上层访问的接口，在下层infra做具体方法的实现
type TemplateRepository interface {
	// CreateUserByIDAndName 这里定义一些示例方法，这样就可以先编排逻辑了，后面再去下层实现接口
	CreateUserByIDAndName(ctx context.Context, id string, name string) (*model.Users, error)
	// GetUserByID 通过ID获取用户信息
	GetUserByID(ctx context.Context, id string) (*model.Users, error)
	// UpsertConversation 插入或更新对话记录
	UpsertConversation(ctx context.Context, userID string, conversationID string, openaiMessages []openai.ChatCompletionMessageParamUnion) error
	// GetConversationByID 通过ID获取对话记录
	GetConversationByID(ctx context.Context, id string) (*model.Conversations, error)
}
