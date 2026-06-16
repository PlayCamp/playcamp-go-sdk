package playcamp

import (
	"context"

	"github.com/playcamp/playcamp-go-sdk/internal/httpclient"
)

// PlaytimeSessionService provides access to playtime endpoints for the Server API.
type PlaytimeSessionService struct {
	client   *httpclient.Client
	basePath string
}

func newPlaytimeSessionService(client *httpclient.Client) *PlaytimeSessionService {
	return &PlaytimeSessionService{client: client, basePath: "/v1/server/playtime/sessions"}
}

// Create reports a single game session's playtime.
func (s *PlaytimeSessionService) Create(ctx context.Context, params CreatePlaytimeSessionParams) (*PlaytimeSession, error) {
	if err := requireNonEmpty("sessionId", params.SessionID); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("userId", params.UserID); err != nil {
		return nil, err
	}
	if params.DurationSeconds <= 0 {
		return nil, &InputValidationError{Field: "durationSeconds", Message: "must be a positive integer"}
	}
	if params.StartedAt.IsZero() || params.EndedAt.IsZero() {
		return nil, &InputValidationError{Field: "startedAt/endedAt", Message: "must be non-zero time.Time"}
	}
	if params.EndedAt.Before(params.StartedAt) {
		return nil, &InputValidationError{Field: "endedAt", Message: "must be on or after startedAt"}
	}
	params.StartedAt = params.StartedAt.UTC()
	params.EndedAt = params.EndedAt.UTC()
	var result PlaytimeSession
	if err := s.client.Post(ctx, s.basePath, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateBulk reports playtime sessions in bulk (up to 1000).
func (s *PlaytimeSessionService) CreateBulk(ctx context.Context, params CreateBulkPlaytimeSessionParams) (*BulkPlaytimeSessionResult, error) {
	if len(params.Sessions) == 0 {
		return nil, &InputValidationError{Field: "sessions", Message: "must contain at least one session"}
	}
	for i := range params.Sessions {
		params.Sessions[i].StartedAt = params.Sessions[i].StartedAt.UTC()
		params.Sessions[i].EndedAt = params.Sessions[i].EndedAt.UTC()
	}
	var result BulkPlaytimeSessionResult
	if err := s.client.Post(ctx, s.basePath+"/bulk", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
