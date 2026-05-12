package v1

import (
	"net/http"
	"reftrail/store"

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
