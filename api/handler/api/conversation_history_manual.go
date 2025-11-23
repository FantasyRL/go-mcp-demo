package api

import (
	"context"

	"github.com/FantasyRL/go-mcp-demo/api/pack"
	"github.com/FantasyRL/go-mcp-demo/internal/host/repository"
	"github.com/cloudwego/hertz/pkg/app"
	consts "github.com/cloudwego/hertz/pkg/protocol/consts"
)

func GetConversationHistory(ctx context.Context, c *app.RequestContext) {
	convID := c.Query("conversation_id")
	if convID == "" {
		c.String(consts.StatusBadRequest, "conversation_id required")
		return
	}

	repo := repository.NewConversationRepository(clientSet.ActualDB)
	msgs, _, updatedAt, err := repo.GetHistory(ctx, convID)
	if err != nil {
		pack.RespError(c, err)
		return
	}

	out := make([]map[string]any, 0, len(msgs))
	for _, m := range msgs {
		item := map[string]any{
			"role":    m.Role,
			"content": m.Content,
			"images":  m.Images,
		}
		if m.ToolName != "" {
			item["tool_name"] = m.ToolName
		}
		out = append(out, item)
	}

	resp := map[string]any{
		"conversation_id": convID,
		"messages":        out,
		"total":           len(out),
		"updated_at_ms":   updatedAt,
	}
	pack.RespData(c, resp)
}
