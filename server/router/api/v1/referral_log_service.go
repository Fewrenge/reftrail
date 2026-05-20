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

	// Define a presentation model local to this endpoint's response
	type LogResponse struct {
		ID            domain.ReferralLogID `json:"id"`
		EntryID       domain.ReferralID    `json:"entryId"`
		Username      string               `json:"username"`
		UserFirstName string               `json:"userFirstName"`
		UserLastName  string               `json:"userLastName"`
		OldStatus     string               `json:"oldStatus"`
		NewStatus     string               `json:"newStatus"`
		Note          string               `json:"note"`
		CreatedTs     string               `json:"createdTs"`
	}

	logPayload := make([]LogResponse, len(logs))
	for i, l := range logs {
		logPayload[i] = LogResponse{
			ID:            l.ID,
			EntryID:       l.EntryID,
			Username:      l.Username,
			UserFirstName: l.UserFirstName,
			UserLastName:  l.UserLastName,
			OldStatus:     l.OldStatus,
			NewStatus:     l.NewStatus,
			Note:          l.Note,
			CreatedTs:     l.CreatedTs,
			//l.UserID is deliberately excluded
		}
	}

	// 3. Return the "Timeline" to the user
	return c.JSON(http.StatusOK, logPayload)
}
