package v1

import (
	"log"
	"net/http"
	"strconv"
	"wl/store"

	echo "github.com/labstack/echo/v5"
)

func (s *APIV1Service) GetWaitlistHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	list, err := s.Store.ListWLEntries(ctx, &store.FindWLEntry{})
	if err != nil {
		log.Printf("Database error: %v", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, list)
}

func (s *APIV1Service) CreateWLEntryHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	create := &store.CreateWLEntry{}

	if err := c.Bind(create); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	entry, err := s.Store.CreateWLEntry(ctx, create)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, entry)
}

func (s *APIV1Service) UpdateWLEntryHandler(c *echo.Context) error {
	// 1. Get the ID from the URL (e.g., /api/v1/waitlist/1)
	id, _ := strconv.Atoi(c.Param("id"))

	update := &store.UpdateWLEntry{ID: int32(id)}
	if err := c.Bind(update); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := s.Store.UpdateWLEntry(c.Request().Context(), update); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, true)
}

func (s *APIV1Service) DeleteWLEntryHandler(c *echo.Context) error {
	// 1. Get the ID from the URL (/api/v1/waitlist/15 -> 15)
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid ID format")
	}

	// 2. Call the "Janitor" (Store.DeleteWLEntry)
	// We wrap the ID into the struct your store expects
	err = s.Store.DeleteWLEntry(c.Request().Context(), &store.DeleteWLEntry{
		ID: int32(id),
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// 3. Return "No Content" (Status 204) to say "It's gone!"
	return c.NoContent(http.StatusNoContent)
}
