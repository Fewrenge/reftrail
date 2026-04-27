package v1

import (
	"net/http"
	"strconv"

	echo "github.com/labstack/echo/v5"
)

// Helper to handle those Blueprint pointers (*) we defined earlier
func ptrInt32(v int32) *int32 {
	return &v
}

// ListReferralLogsHandler handles GET /api/v1/referrals/:id/logs
func (s *APIV1Service) ListReferralLogsHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// 1. Extract the Patient ID from the URL path
	// Example: /api/v1/referrals/1/logs -> id = 1
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid patient ID"})
	}

	// 2. Ask the Manager (Store) for the history
	logs, err := s.Store.ListReferralLogs(ctx, int32(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// 3. Return the "Timeline" to the user
	return c.JSON(http.StatusOK, logs)
}
