package playcamp

import "time"

// PlaytimeSession represents a recorded playtime session response.
type PlaytimeSession struct {
	SessionID       string `json:"sessionId"`
	UserID          string `json:"userId"`
	DurationSeconds int    `json:"durationSeconds"`
	// Recorded is false for an idempotent re-send (the existing row is kept).
	Recorded  bool   `json:"recorded"`
	CreatedAt string `json:"createdAt"`
}

// CreatePlaytimeSessionParams specifies parameters for reporting a session.
type CreatePlaytimeSessionParams struct {
	SessionID       string                 `json:"sessionId"`
	UserID          string                 `json:"userId"`
	DurationSeconds int                    `json:"durationSeconds"`
	StartedAt       time.Time              `json:"startedAt"`
	EndedAt         time.Time              `json:"endedAt"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CallbackID      string                 `json:"callbackId,omitempty"`
	IsTest          *bool                  `json:"isTest,omitempty"`
}

// CreateBulkPlaytimeSessionParams specifies parameters for bulk reporting.
type CreateBulkPlaytimeSessionParams struct {
	Sessions   []CreatePlaytimeSessionParams `json:"sessions"`
	CallbackID string                        `json:"callbackId,omitempty"`
	IsTest     *bool                         `json:"isTest,omitempty"`
}

// BulkPlaytimeSessionResultItem represents one session result within a bulk op.
//
// Note: unlike bulk payments, the item has no Data field.
type BulkPlaytimeSessionResultItem struct {
	SessionID string  `json:"sessionId"`
	Status    string  `json:"status"` // "SUCCESS", "SKIPPED", or "FAILED"
	Error     *string `json:"error,omitempty"`
}

// BulkPlaytimeSessionResult represents the bulk response.
type BulkPlaytimeSessionResult struct {
	TotalRequested int                             `json:"totalRequested"`
	Successful     int                             `json:"successful"`
	Failed         int                             `json:"failed"`
	Skipped        int                             `json:"skipped"`
	Results        []BulkPlaytimeSessionResultItem `json:"results"`
}
