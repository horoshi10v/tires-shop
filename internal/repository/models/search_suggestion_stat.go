package models

import "time"

type SearchSuggestionStat struct {
	Base
	Scope                string    `gorm:"type:varchar(16);not null;uniqueIndex:idx_search_suggestion_scope_key"`
	Suggestion           string    `gorm:"type:varchar(160);not null"`
	NormalizedSuggestion string    `gorm:"type:varchar(160);not null;uniqueIndex:idx_search_suggestion_scope_key;index"`
	UsageCount           int       `gorm:"not null;default:1"`
	LastSelectedAt       time.Time `gorm:"not null;index"`
}
