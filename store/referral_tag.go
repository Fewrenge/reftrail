package store

import (
	"context"
	"errors"
	"reftrail/internal/domain"
)

type ReferralTag struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedTS   string `json:"createdTs"`
}

type CreateReferralTag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Store) CreateReferralTag(ctx context.Context, create *CreateReferralTag) (*ReferralTag, error) {
	user, ok := domain.GetUserContext(ctx)
	if !ok || user.Role != "REFTRAIL_ADMIN" {
		return nil, errors.New("forbidden: only admins can create system tags")
	}
	return s.driver.CreateReferralTag(ctx, create)
}

func (s *Store) ListReferralTags(ctx context.Context) ([]*ReferralTag, error) {
	return s.driver.ListReferralTags(ctx)
}
