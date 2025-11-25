package application

import (
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
)

func (h *Host) GetConversation(
	userID string,
	conversationID string,
) (*model.Conversations, error) {
	conversation, err := h.templateRepository.GetConversationByID(h.ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		// 未找到
		return nil, errno.ParamError
	}
	if conversation.UserID != userID {
		return nil, errno.AuthError
	}

	return conversation, nil
}
