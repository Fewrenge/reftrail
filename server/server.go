package server

import (
	"reftrail/server/auth"
	v1 "reftrail/server/router/api/v1" // This is your v1 package
	"reftrail/store"

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
	protected.Use(auth.JWTMiddleware)

	// --- THE CLEAN LIST ---
	// Get the whole list
	protected.GET("/waitlist", v1Service.GetWaitlistHandler)

	// Get ONE specific waitlist entry (The :id sniper)
	protected.GET("/waitlist/:id", v1Service.GetReferralEntryHandler)

	// Create a new waitlist entry
	protected.POST("/waitlist", v1Service.CreateReferralEntryHandler)

	// Update a waitlist entry (The state switcher)
	protected.PATCH("/waitlist/:id", v1Service.UpdateReferralEntryHandler)

	// Get the history logs
	protected.GET("/waitlist/:id/logs", v1Service.ListReferralLogsHandler)

	// Get current user
	protected.GET("/users/me", v1Service.GetCurrentUserHandler)

	// Log out
	protected.POST("/logout", v1Service.LogoutHandler)

	// Change password
	protected.PATCH("/users/password", v1Service.ChangePasswordHandler)

	admin := protected.Group("")
	admin.Use(auth.AdminOnlyMiddleware)                                 // Add the extra gatekeeper
	admin.POST("/users", v1Service.CreateUserHandler)                   // Create a user
	admin.GET("/users", v1Service.ListUsersHandler)                     // List Users
	admin.DELETE("/users/:id", v1Service.DeleteUserHandler)             // Delete a user
	admin.DELETE("/waitlist/:id", v1Service.DeleteReferralEntryHandler) // Delete a waitlist entry
}
