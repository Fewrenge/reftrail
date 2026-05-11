package v1

import (
	"net/http"
	"reftrail/internal/domain"

	echo "github.com/labstack/echo/v5"
)

// ListReferralLogsHandler handles GET /api/v1/referrals/:id/logs
func (s *APIV1Service) ListReferralLogsHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// 1. Extract the Patient ID from the URL path
	// Example: /api/v1/referrals/1/logs -> id = 1
	idStr := c.Param("id")
	refID := domain.ReferralID(idStr)

	// 2. Ask the Manager (Store) for the history
	logs, err := s.Store.ListReferralLogs(ctx, refID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// 3. Return the "Timeline" to the user
	return c.JSON(http.StatusOK, logs)
}
