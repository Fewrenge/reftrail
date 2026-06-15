package v1

import (
	"errors"
	"net/http"
	"net/url"
	"reftrail/internal/domain"
	"reftrail/store"
	"strings"

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

// UpdateReferralTagDefinitionHandler handles the tag configuration update request
// PATCH /api/v1/tags/:id
func (s *APIV1Service) UpdateReferralTagDefinitionHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	var update store.UpdateReferralTagDefinition

	if err := c.Bind(&update); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body format"})
	}

	// Capture the current name from path parameter
	update.OldName = c.Param("id")
	if update.OldName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing identifier parameter"})
	}

	// Fallback mechanism: If newName wasn't changed/sent in JSON, protect the current value
	if update.NewName == "" {
		update.NewName = update.OldName
	}

	updatedTag, err := s.Store.UpdateReferralTagDefinition(ctx, &update)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access denied"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, updatedTag)
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
	// 1. Get ID from the URL: /api/v1/tags/:tagName
	idStr := c.Param("id")
	name, err := url.PathUnescape(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid tag ID format")
	}

	// 2. Call the store
	err = s.Store.DeleteReferralTag(c.Request().Context(), &store.DeleteReferralTag{Name: name})
	if err != nil {
		// If you returned "not found" in the driver, this will send that error
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// 3. Return 204 No Content (Standard for successful deletion)
	return c.NoContent(http.StatusNoContent)
}

func (s *APIV1Service) AssignTagHandler(c *echo.Context) error {
	refIDStr := c.Param("id")
	refID := domain.ReferralID(refIDStr)
	tagName := c.Param("tagName")
	tagName, err := url.PathUnescape(tagName)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Malformed tag name parameter"})
	}

	cleanTagName := strings.ToUpper(strings.TrimSpace(tagName))
	if err := s.Store.AssignTagToReferral(c.Request().Context(), refID, cleanTagName); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusCreated)
}

func (s *APIV1Service) RemoveTagHandler(c *echo.Context) error {
	refIDStr := c.Param("id")
	refID := domain.ReferralID(refIDStr)
	tagName := c.Param("tagName")

	tagName, err := url.PathUnescape(tagName)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Malformed tag name parameter"})
	}

	if tagName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing tag name parameter"})
	}

	cleanTagName := strings.ToUpper(strings.TrimSpace(tagName))
	if err := s.Store.RemoveTagFromReferral(c.Request().Context(), refID, cleanTagName); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
