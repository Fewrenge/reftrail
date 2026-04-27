package v1

import (
	"log"
	"net/http"
	"reftrail/internal/domain"
	"reftrail/store"
	"strconv"

	echo "github.com/labstack/echo/v5"
)

// Get all referrals
func (s *APIV1Service) GetReferralsHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	list, err := s.Store.ListReferralEntries(ctx, &store.FindReferralEntry{})
	if err != nil {
		log.Printf("Database error: %v", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, list)
}

// GetReferralEntryHandler handles GET /api/v1/referrals/:id
// Focuses on one referral
func (s *APIV1Service) GetReferralEntryHandler(c *echo.Context) error {
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
	entry, err := s.Store.GetReferralEntry(ctx, &store.FindReferralEntry{
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

func (s *APIV1Service) CreateReferralEntryHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	create := &store.CreateReferralEntry{}

	if err := c.Bind(create); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	entry, err := s.Store.CreateReferralEntry(ctx, create)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, entry)
}

func (s *APIV1Service) UpdateReferralEntryHandler(c *echo.Context) error {
	// 1. Get the ID from the URL (e.g., /api/v1/referrals/1)
	id, _ := strconv.Atoi(c.Param("id"))

	update := &store.UpdateReferralEntry{ID: int32(id)}
	if err := c.Bind(update); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := s.Store.UpdateReferralEntry(c.Request().Context(), update); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, true)
}

func (s *APIV1Service) UpdateReferralEntryStatusHandler(c *echo.Context) error {
	// 1. Get ID from URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid referral ID")
	}

	// 2. Bind Request (Only the status)
	var req struct {
		Status domain.ReferralStatus `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	// 3. Get User Context (The "Who")
	user, ok := domain.GetUserContext(c.Request().Context())
	if !ok {
		return c.JSON(http.StatusUnauthorized, "User context missing")
	}

	// 4. Fetch CURRENT status from DB
	// We need to know where we are to know if we can move
	currentStatus, err := s.Store.GetReferralEntryStatusByID(c.Request().Context(), int32(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, "Referral not found")
	}

	// 5. Check State Machine & God Mode
	if !domain.CanTransition(currentStatus, req.Status, user.Role) {
		return c.JSON(http.StatusForbidden, "Illegal status transition for your role")
	}

	// 6. Update the DB
	// We pass user.ID so the Store can eventually handle the "Who" in the logs
	err = s.Store.UpdateReferralEntryStatus(c.Request().Context(), int32(id), req.Status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to update status")
	}

	return c.JSON(http.StatusOK, true)
}

func (s *APIV1Service) DeleteReferralEntryHandler(c *echo.Context) error {
	// 1. Get the ID from the URL (/api/v1/referrals/15 -> 15)
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid ID format")
	}

	// 2. Call the "Janitor" (Store.DeleteReferralEntry)
	// We wrap the ID into the struct your store expects
	err = s.Store.DeleteReferralEntry(c.Request().Context(), &store.DeleteReferralEntry{
		ID: int32(id),
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// 3. Return "No Content" (Status 204) to say "It's gone!"
	return c.NoContent(http.StatusNoContent)
}
