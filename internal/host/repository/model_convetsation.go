package repository

import "time"


type Conversation struct {
	ID             int64  `gorm:"primaryKey"`
	ConversationID string `gorm:"uniqueIndex;size:64"`
	UserID         int64  `gorm:"index"`
	HistoryJSON    []byte `gorm:"type:jsonb"`
	MessageCount   int
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

func (Conversation) TableName() string { return "conversations" }
