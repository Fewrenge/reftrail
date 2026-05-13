package v1

import (
	"log/slog"
	"net/http"
	"reftrail/internal/domain"
	"reftrail/store"

	echo "github.com/labstack/echo/v5"
)

// Get all referrals
func (s *APIV1Service) ListReferralEntriesHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	list, err := s.Store.ListReferralEntries(ctx, &store.FindReferralEntry{})
	if err != nil {
		slog.Error("failed to get referral entries list", "error", err.Error())

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve referral records",
		})
	}
	if list == nil {
		list = []*store.ReferralEntry{}
	}

	return c.JSON(http.StatusOK, list)
}

// GetReferralEntryHandler handles GET /api/v1/referrals/:id
// Focuses on one referral
func (s *APIV1Service) GetReferralEntryHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// 1. Extract the "id" from the URL path parameter
	idStr := c.Param("id")
	refID := domain.ReferralID(idStr)

	// 2. Ask the Manager (Store) to find this specific entry
	// We use our 'Find' blueprint here
	entry, err := s.Store.GetReferralEntry(ctx, &store.FindReferralEntry{
		ID: &refID,
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
		slog.Warn("Failed to bind malformed JSON request body", "error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload syntax"})
	}

	if err := domain.ValidateStruct(create); err != nil {
		slog.Warn("Referral payload structural validation failed",
			"patient_last_name", create.PatientLastName,
			"patient_first_name", create.PatientFirstName,
			"error", err.Error(),
		)

		return c.JSON(http.StatusUnprocessableEntity, map[string]string{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	entry, err := s.Store.CreateReferralEntry(ctx, create)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create referral"})
	}
	return c.JSON(http.StatusOK, entry)
}

func (s *APIV1Service) BatchCreateReferralEntriesHandler(c *echo.Context) error {
	var req store.BatchCreateReferralEntries

	// 1. Bind the JSON body to our struct
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid batch data format")
	}

	// 2. Optional: Basic validation (e.g., check if empty)
	if len(req.ReferralEntries) == 0 {
		return c.JSON(http.StatusBadRequest, "No entries provided")
	}

	// 3. Call the Store
	err := s.Store.BatchCreateReferralEntries(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Batch import successful"})
}

func (s *APIV1Service) UpdateReferralEntryHandler(c *echo.Context) error {
	// 1. Get the ID from the URL
	idStr := c.Param("id")
	refID := domain.ReferralID(idStr)

	update := &store.UpdateReferralEntry{ID: refID}
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
	idStr := c.Param("id")
	refID := domain.ReferralID(idStr)

	// 2. Bind Request (Only the status)
	var req store.UpdateReferralEntryStatus
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}
	req.ID = refID

	// 3. Update the DB
	// The Store now handles: Transaction, Old Status Check, Role Logic, and Logging
	err := s.Store.UpdateReferralEntryStatus(c.Request().Context(), &req)
	if err != nil {
		// You can check the error type here to return 403 vs 500
		if err.Error() == "illegal status transition" {
			return c.JSON(http.StatusForbidden, err.Error())
		}
		//return c.JSON(http.StatusInternalServerError, "Failed to update status")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Internal Error",
			"debug":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, true)
}

func (s *APIV1Service) DeleteReferralEntryHandler(c *echo.Context) error {
	// 1. Get the ID from the URL
	idStr := c.Param("id")
	refID := domain.ReferralID(idStr)

	// 2. Call the "Janitor" (Store.DeleteReferralEntry)
	// We wrap the ID into the struct your store expects
	err := s.Store.DeleteReferralEntry(c.Request().Context(), &store.DeleteReferralEntry{
		ID: refID,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// 3. Return "No Content" (Status 204) to say "It's gone!"
	return c.NoContent(http.StatusNoContent)
}
