package v1

import (
	"log/slog"
	"net/http"
	"reftrail/store" // Adjust this to match your exact module path
	"time"

	echo "github.com/labstack/echo/v5"
)

// GetUrgencyAnalyticsHandler extracts filters and returns pie-chart metric data
// GET /api/v1/analytics/urgency-distribution
func (s *APIV1Service) GetUrgencyAnalyticsHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	find := &store.FindReferralEntry{}

	// 1. Automatically parse query strings into the filter struct
	// e.g., /api/v1/analytics/urgency-distribution?referralDateFrom=2026-01-01&referralDateTo=2026-03-31
	if err := c.Bind(find); err != nil {
		slog.Warn("Failed parsing analytics query parameters", "error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid query filter parameters"})
	}

	// 2. Optional Validation: Ensure date formats are valid if they are supplied
	if find.ReferralDateFrom != nil && *find.ReferralDateFrom != "" {
		if _, err := time.Parse("2006-01-02", *find.ReferralDateFrom); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid referralDateFrom format. Use YYYY-MM-DD"})
		}
	}
	if find.ReferralDateTo != nil && *find.ReferralDateTo != "" {
		if _, err := time.Parse("2006-01-02", *find.ReferralDateTo); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid referralDateTo format. Use YYYY-MM-DD"})
		}
	}

	// 3. Call your SQLite driver layer execution function
	response, err := s.Store.GetUrgencyDistribution(ctx, find)
	if err != nil {
		slog.Error("Database execution error during analytics calculation", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to compile metric calculations"})
	}

	// 4. Return the aggregated data to the frontend chart component
	return c.JSON(http.StatusOK, response)
}
