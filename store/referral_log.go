package store

import (
	"context"
	"fmt"
	"reftrail/internal/domain"
)

type ReferralLog struct {
	ID              domain.ReferralLogID   `json:"id"`
	ReferralID      domain.ReferralID      `json:"referralId"`
	CreatorUsername domain.Username        `json:"-"`
	OldStatus       *domain.ReferralStatus `json:"oldStatus,omitempty"`
	NewStatus       domain.ReferralStatus  `json:"newStatus"`
	Note            string                 `json:"note"`
	CreatedTs       string                 `json:"createdTs"`
}

type ReferralLogWithUser struct {
	ReferralLog
	UserPublicInfo
}

func (s *Store) CreateReferralLog(ctx context.Context, create *ReferralLog) (*ReferralLog, error) {
	var logPayload *ReferralLog
	err := s.driver.RunInTransaction(ctx, func(txCtx context.Context) error {

		// 2. Extract user context to verify permissions and ownership
		user, ok := domain.GetUserContext(txCtx)
		if !ok {
			return domain.ErrUnauthorized
		}
		create.CreatorUsername = domain.Username(user.Username)

		// 3. Fetch the stable current status from inside the transaction
		currentStatus, err := s.driver.GetReferralEntryStatusByID(txCtx, create.ReferralID)
		if err != nil {
			return fmt.Errorf("failed to fetch status for log: %w", err)
		}

		// 4. Standalone notes do not alter state: Old == New
		create.OldStatus = &currentStatus
		create.NewStatus = currentStatus

		// 5. Hand the work to the driver safely
		res, err := s.driver.CreateReferralLog(txCtx, create)
		if err != nil {
			return fmt.Errorf("driver failed to create standalone log: %w", err)
		}

		logPayload = res
		return nil
	})

	return logPayload, err
}

// Manager Logic: Notice we don't use a "Find" struct here.
// We just ask for the ID of the referral we care about.
func (s *Store) ListReferralLogs(ctx context.Context, referralID domain.ReferralID) ([]*ReferralLogWithUser, error) {
	return s.driver.ListReferralLogs(ctx, referralID)
}
