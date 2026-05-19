package v1

import (
	"log/slog"
	"net/http"
	"reftrail/internal/domain"
	"reftrail/store"

	echo "github.com/labstack/echo/v5"
)

// POST /api/v1/referrals/:id/logs
// For recording notes that don't correspond to a status change only.
// For status changes, the log is created automatically inside the UpdateReferralEntry function in the Manager.
func (s *APIV1Service) CreateReferralLogHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// 1. Extract the referral entry ID from the URL path
	idStr := c.Param("id")
	refID := domain.ReferralID(idStr)

	// 2. Extract the log details from the request body
	log := &store.ReferralLog{}
	if err := c.Bind(log); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	log.EntryID = refID

	currentStatus, err := s.Store.GetReferralEntryStatusByID(ctx, refID)
	if err != nil {
		slog.Error("failed to get referral entry status by ID", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to verify referral status"})
	}

	log.OldStatus = string(currentStatus)
	log.NewStatus = string(currentStatus)

	// 3. Ask the Manager (Store) to create the log
	createdLog, err := s.Store.CreateReferralLog(ctx, log)
	if err != nil {
		slog.Error("failed to create referral log", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create referral log"})
	}

	// 4. Return the created log to the user
	return c.JSON(http.StatusCreated, createdLog)
}

// GET /api/v1/referrals/:id/logs
func (s *APIV1Service) ListReferralLogsHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// 1. Extract the referral entry ID from the URL path
	idStr := c.Param("id")
	refID := domain.ReferralID(idStr)

	// 2. Ask the Manager (Store) for the history
	logs, err := s.Store.ListReferralLogs(ctx, refID)
	if err != nil {
		slog.Error("failed to list referral logs", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list referral logs"})
	}

	// 3. Return the "Timeline" to the user
	return c.JSON(http.StatusOK, logs)
}
