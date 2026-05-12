package v1

import (
	"net/http"
	"reftrail/store"
	"strconv"

	echo "github.com/labstack/echo/v5"
)

func (s *APIV1Service) CreateReferralTagHandler(c *echo.Context) error {
	var req store.CreateReferralTag
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid data")
	}

	tag, err := s.Store.CreateReferralTag(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, tag)
}

func (s *APIV1Service) ListReferralTagsHandler(c *echo.Context) error {
	tags, err := s.Store.ListReferralTags(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// Returning an empty slice [] instead of nil if no tags exist
	// This makes it easier for your TypeScript frontend to .map()
	return c.JSON(http.StatusOK, tags)
}

func (s *APIV1Service) DeleteReferralTagHandler(c *echo.Context) error {
	// 1. Get ID from the URL: /api/v1/admin/tags/2
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid tag ID format")
	}

	// 2. Call the store
	err = s.Store.DeleteReferralTag(c.Request().Context(), &store.DeleteReferralTag{ID: id})
	if err != nil {
		// If you returned "not found" in the driver, this will send that error
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// 3. Return 204 No Content (Standard for successful deletion)
	return c.NoContent(http.StatusNoContent)
}
