package server

import (
	"wl/server/auth"
	v1 "wl/server/router/api/v1" // This is your v1 package
	"wl/store"

	echo "github.com/labstack/echo/v5"
)

type Server struct {
	Store  *store.Store
	Engine *echo.Echo
}

func NewServer(s *store.Store) *Server {
	e := echo.New()

	srv := &Server{
		Store:  s,
		Engine: e,
	}

	// Now this will work because we define it below
	srv.registerWaitlistRoutes()

	return srv
}

// FIX: Added the missing Start method
func (s *Server) Start(address string) error {
	return s.Engine.Start(address)
}

func (s *Server) registerWaitlistRoutes() {
	v1Service := &v1.APIV1Service{Store: s.Store}

	// PUBLIC
	s.Engine.POST("/api/v1/login", v1Service.LoginHandler)

	// PROTECTED (Requires JWT)
	protected := s.Engine.Group("/api/v1")
	protected.Use(auth.JWTMiddleware) // Keep this off until we see John Doe!

	// --- THE CLEAN LIST ---
	// Rule 1: Get the whole list
	protected.GET("/waitlist", v1Service.GetWaitlistHandler)

	// Rule 2: Get ONE specific patient (The :id sniper)
	protected.GET("/waitlist/:id", v1Service.GetWLEntryHandler)

	// Rule 3: Create a new patient
	protected.POST("/waitlist", v1Service.CreateWLEntryHandler)

	// Rule 4: Update a patient (The state switcher)
	protected.PATCH("/waitlist/:id", v1Service.UpdateWLEntryHandler)

	// Rule 5: Get the history logs
	protected.GET("/waitlist/:id/logs", v1Service.ListWLLogsHandler)

	protected.POST("/users", v1Service.CreateUserHandler)
}
