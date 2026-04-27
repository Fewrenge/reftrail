package v1

import (
	"log"
	"net/http"
	"reftrail/store"
	"strconv"

	echo "github.com/labstack/echo/v5"
)

// GetWLEntryHandler handles GET /api/v1/waitlist/:id
func (s *APIV1Service) GetWLEntryHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// 1. Extract the "id" from the URL path parameter
	idStr := c.Param("id")

	log.Printf("Sniper Handler triggered with ID: [%s]", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid ID format"})
	}

	// 2. Ask the Manager (Store) to find this specific entry
	// We use our 'Find' blueprint here
	entry, err := s.Store.GetWLEntry(ctx, &store.FindWLEntry{
		ID: ptrInt32(int32(id)),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// 3. If no patient was found, return a 404
	if entry == nil {
		return c.JSON(http.StatusNotFound, map[string]any{"message": "Patient entry not found"})
	}

	// 4. Return the patient data as JSON
	return c.JSON(http.StatusOK, entry)
}

// Helper to handle those Blueprint pointers (*) we defined earlier
func ptrInt32(v int32) *int32 {
	return &v
}

// ListWLLogsHandler handles GET /api/v1/waitlist/:id/logs
func (s *APIV1Service) ListWLLogsHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// 1. Extract the Patient ID from the URL path
	// Example: /api/v1/waitlist/1/logs -> id = 1
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid patient ID"})
	}

	// 2. Ask the Manager (Store) for the history
	logs, err := s.Store.ListWLLogs(ctx, int32(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// 3. Return the "Timeline" to the user
	return c.JSON(http.StatusOK, logs)
}
