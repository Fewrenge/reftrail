package v1

import (
	"net/http"
	"reftrail/store"

	echo "github.com/labstack/echo/v5"
)

func (s *APIV1Service) CreateTagHandler(c *echo.Context) error {
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
