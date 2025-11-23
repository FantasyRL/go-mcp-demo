package repository

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"
)

//  用于持久化的简化消息结构
type AIMessageDTO struct {
	Role     string   `json:"role"`
	Content  string   `json:"content"`
	ToolName string   `json:"tool_name,omitempty"`
	Images   []string `json:"images,omitempty"`
}

//定义会话历史的存取接口
type ConversationRepository interface {
	UpsertHistory(ctx context.Context, conversationID string, userID int64, messages []AIMessageDTO) error
	GetHistory(ctx context.Context, conversationID string) (messages []AIMessageDTO, userID int64, updatedAtMs int64, err error)
}

// 实现
type conversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) UpsertHistory(ctx context.Context, conversationID string, userID int64, messages []AIMessageDTO) error {
	raw, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	var conv Conversation
	err = r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).First(&conv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			conv = Conversation{
				ConversationID: conversationID,
				UserID:         userID,
				HistoryJSON:    raw,
				MessageCount:   len(messages),
			}
			return r.db.WithContext(ctx).Create(&conv).Error
		}
		return err
	}

	conv.HistoryJSON = raw
	conv.MessageCount = len(messages)
	return r.db.WithContext(ctx).Save(&conv).Error
}

func (r *conversationRepository) GetHistory(ctx context.Context, conversationID string) (messages []AIMessageDTO, userID int64, updatedAtMs int64, err error) {
	var conv Conversation
	err = r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).First(&conv).Error
	if err != nil {
		return nil, 0, 0, err
	}

	userID = conv.UserID
	if err = json.Unmarshal(conv.HistoryJSON, &messages); err != nil {
		return nil, 0, 0, err
	}
	updatedAtMs = conv.UpdatedAt.UnixMilli()
	return
}
