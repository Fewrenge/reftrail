package store

import (
	"context"
	"reftrail/internal/domain"
)

type ReferralTag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateReferralTag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateReferralTagDefinition struct {
	Name           string `json:"name"`
	NewDescription string `json:"newDescription"`
}

type DeleteReferralTag struct {
	Name string `json:"name"`
}

type LinkedReferralTagRow struct {
	ReferralID domain.ReferralID
	TagName    string
}

func (s *Store) CreateReferralTag(ctx context.Context, create *CreateReferralTag) (*ReferralTag, error) {
	user, ok := domain.GetUserContext(ctx)
	if !ok || user.Role != "REFTRAIL_ADMIN" {
		return nil, domain.ErrForbidden
	}
	return s.driver.CreateReferralTag(ctx, create)
}

func (s *Store) UpdateReferralTagDefinition(ctx context.Context, update *UpdateReferralTagDefinition) (*ReferralTag, error) {
	user, ok := domain.GetUserContext(ctx)
	if !ok || user.Role != "REFTRAIL_ADMIN" {
		return nil, domain.ErrForbidden
	}
	return s.driver.UpdateReferralTagDefinition(ctx, update)
}

func (s *Store) ListReferralTags(ctx context.Context) ([]*ReferralTag, error) {
	return s.driver.ListReferralTags(ctx)
}

func (s *Store) DeleteReferralTag(ctx context.Context, delete *DeleteReferralTag) error {
	user, ok := domain.GetUserContext(ctx)
	if !ok || user.Role != "REFTRAIL_ADMIN" {
		return domain.ErrForbidden
	}
	return s.driver.DeleteReferralTag(ctx, delete)
}

func (s *Store) AssignTagToReferral(ctx context.Context, referralID domain.ReferralID, tagName string) error {
	return s.driver.AssignTagToReferral(ctx, referralID, tagName)
}

func (s *Store) RemoveTagFromReferral(ctx context.Context, referralID domain.ReferralID, tagName string) error {
	return s.driver.RemoveTagFromReferral(ctx, referralID, tagName)
}
