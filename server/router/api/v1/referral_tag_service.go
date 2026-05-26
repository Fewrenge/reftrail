package v1

import (
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
