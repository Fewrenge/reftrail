package v1

import (
	"log/slog"
	"net/http"
	"reftrail/store"
	"strconv"
	"strings"

	echo "github.com/labstack/echo/v5"
)

// GET /api/v1/physicians (Protected)
func (s *APIV1Service) ListReferralPhysiciansHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	find := &store.FindReferralPhysician{}

	if err := c.Bind(find); err != nil {
		slog.Warn("Failed parsing list query parameters for referral physicians", "error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid query filter parameters"})
	}

	// Map optional query parameters from the request URL
	if id := c.QueryParam("id"); id != "" {
		find.ID = &id
	}
	if cpso := c.QueryParam("cpsoNumber"); cpso != "" {
		find.CPSONumber = &cpso
	}
	if firstName := c.QueryParam("firstName"); firstName != "" {
		find.FirstName = &firstName
	}
	if lastName := c.QueryParam("lastName"); lastName != "" {
		find.LastName = &lastName
	}
	if emrID := c.QueryParam("emrPhysicianId"); emrID != "" {
		find.EMRPhysicianID = &emrID
	}
	if search := c.QueryParam("generalTerm"); search != "" {
		find.GeneralTerm = &search
	}

	if limitQuery := c.QueryParam("limit"); limitQuery != "" {
		if val, err := strconv.Atoi(limitQuery); err == nil {
			limitVal := val // Create a unique local copy
			find.Limit = &limitVal
		}
	}
	if offsetQuery := c.QueryParam("offset"); offsetQuery != "" {
		if val, err := strconv.Atoi(offsetQuery); err == nil {
			offsetVal := val // Create a unique local copy
			find.Offset = &offsetVal
		}
	}

	// Apply business defaults if the frontend didn't pass pagination bounds
	if find.Limit == nil {
		defaultLimit := 10
		find.Limit = &defaultLimit
	}
	if find.Offset == nil {
		defaultOffset := 0
		find.Offset = &defaultOffset
	}

	paginated, err := s.Store.ListReferralPhysicians(ctx, find)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to look up physician list"})
	}

	// Always return an empty array instead of null if the database has no records
	if paginated == nil {
		paginated = &store.PaginatedReferralPhysicians{
			ReferralPhysicians: []*store.ReferralPhysician{},
			TotalCount:         0,
		}
	}

	return c.JSON(http.StatusOK, paginated)
}

// GET /api/v1/physicians/:id (Protected)
func (s *APIV1Service) GetReferralPhysicianByIDHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if strings.TrimSpace(id) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Physician tracking ID cannot be empty"})
	}

	physician, err := s.Store.GetReferralPhysicianByID(ctx, id)
	if err != nil {
		// Differentiate between missing records and deep server errors
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database lookup failure"})
	}

	return c.JSON(http.StatusOK, physician)
}

// POST /api/v1/physicians (Admin Only)
func (s *APIV1Service) CreateReferralPhysicianHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	var payload store.CreateReferralPhysician

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Malformed JSON payload payload data"})
	}

	created, err := s.Store.CreateReferralPhysician(ctx, &payload)
	if err != nil {
		if strings.Contains(err.Error(), "validation failed") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to register physician record"})
	}

	return c.JSON(http.StatusCreated, created)
}

// 4. PATCH /api/v1/physicians/:id (Admin Only)
func (s *APIV1Service) UpdateReferralPhysicianHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var payload store.UpdateReferralPhysician
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Malformed JSON payload details"})
	}

	// Route variable overrides JSON block field property for transactional integrity safety
	payload.ID = id

	err := s.Store.UpdateReferralPhysician(ctx, &payload)
	if err != nil {
		if strings.Contains(err.Error(), "validation failed") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to apply profile changes"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Physician profile updated successfully"})
}

// DELETE /api/v1/physicians/:id (Admin Only)
func (s *APIV1Service) DeleteReferralPhysicianHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	payload := &store.DeleteReferralPhysician{ID: id}

	err := s.Store.DeleteReferralPhysician(ctx, payload)
	if err != nil {
		// Return HTTP 409 Conflict if SQLite foreign key constraint blocks deletion
		if strings.Contains(err.Error(), "linked to ongoing patient referral") {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal removal execution failure"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Physician record deleted successfully"})
}
