package infra

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/FantasyRL/go-mcp-demo/internal/host/repository"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/db"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/query"
	"github.com/openai/openai-go/v2"
	"gorm.io/gorm"
)

var _ repository.TemplateRepository = (*TemplateRepository)(nil)

type TemplateRepository struct {
	db *db.DB[*query.Query]
}

func NewTemplateRepository(db *db.DB[*query.Query]) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) CreateUserByIDAndName(ctx context.Context, id string, name string) (*model.Users, error) {
	d := r.db.Get(ctx)
	user := &model.Users{
		ID:   id,
		Name: name,
	}
	// 由于user是指针类型，Create方法会自动填充user的其他字段（如时间戳）
	err := d.WithContext(ctx).Users.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *TemplateRepository) GetUserByID(ctx context.Context, id string) (*model.Users, error) {
	d := r.db.Get(ctx)
	user, err := d.WithContext(ctx).Users.Where(d.Users.ID.Eq(id)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return user, nil
}

func (r *TemplateRepository) UpsertConversation(
	ctx context.Context,
	userID string,
	conversationID string,
	openaiMessages []openai.ChatCompletionMessageParamUnion,
) error {
	d := r.db.Get(ctx)

	// 先把本次要新增的消息序列化成 JSON
	newBytes, err := json.Marshal(openaiMessages)
	if err != nil {
		return err
	}

	// 查询是否已有这条会话
	q := d.WithContext(ctx)
	conv, err := q.Conversations.
		Where(d.Conversations.ID.Eq(conversationID)).
		Where(d.Conversations.UserID.Eq(userID)).
		First()

	if err != nil {
		// 不存在 => 直接创建
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newConv := &model.Conversations{
				ID:           conversationID,
				UserID:       userID,
				Messages:     string(newBytes), // 直接保存为 JSON 数组
				IsSummarized: 0,
			}
			return q.Conversations.Create(newConv)
		}
		// 其他错误直接返回
		return err
	}

	// 已存在 => 需要把 messages 做 append
	// 假设 messages 字段里始终是 JSON 数组
	var existingMsgs []json.RawMessage
	if len(conv.Messages) > 0 {
		if err := json.Unmarshal([]byte(conv.Messages), &existingMsgs); err != nil {
			// 如果原数据不是合法 JSON，就保守一点直接覆盖
			conv.Messages = string(newBytes)
			return q.Conversations.Save(conv)
		}
	}

	var newMsgs []json.RawMessage
	if err := json.Unmarshal(newBytes, &newMsgs); err != nil {
		return err
	}

	merged := append(existingMsgs, newMsgs...)
	mergedBytes, err := json.Marshal(merged)
	if err != nil {
		return err
	}

	conv.Messages = string(mergedBytes)

	return q.Conversations.Save(conv)
}

func (r *TemplateRepository) GetConversationByID(ctx context.Context, id string) (*model.Conversations, error) {
	d := r.db.Get(ctx)

	conv, err := d.
		WithContext(ctx).
		Conversations.
		Where(d.Conversations.ID.Eq(id)).
		First()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return conv, nil
}
