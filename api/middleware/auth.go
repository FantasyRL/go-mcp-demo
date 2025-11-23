package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

const CtxKeyUserID = "user_id"

func Auth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		auth := string(c.GetHeader("Authorization"))
		if auth == "" {
			auth = string(c.GetHeader("X-Token"))
		}
		if auth == "" {
			c.AbortWithStatusJSON(401, map[string]any{"code": 401, "message": "unauthorized"})
			return
		}
		token := auth
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			token = strings.TrimSpace(auth[7:])
		}

		uid, err := parseUserIDFromToken(token)
		if err != nil || uid <= 0 {
			c.AbortWithStatusJSON(401, map[string]any{"code": 401, "message": "invalid token"})
			return
		}
		c.Set(CtxKeyUserID, uid)
		c.Next(ctx)
	}
}


func parseUserIDFromToken(token string) (int64, error) {

	return 1, nil
}
