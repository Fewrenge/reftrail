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
	srv.registerReferralRoutes()

	return srv
}

// FIX: Added the missing Start method
func (s *Server) Start(address string) error {
	return s.Engine.Start(address)
}

func (s *Server) registerReferralRoutes() {
	v1Service := &v1.APIV1Service{Store: s.Store}

	// PUBLIC
	s.Engine.POST("/api/v1/login", v1Service.LoginHandler)

	// PROTECTED (Requires JWT)
	protected := s.Engine.Group("/api/v1")
	protected.Use(auth.JWTMiddleware)

	// --- THE CLEAN LIST ---
	// Get the whole list
	protected.GET("/referrals", v1Service.GetReferralEntriesHandler)

	// Get ONE specific referrals entry (The :id sniper)
	protected.GET("/referrals/:id", v1Service.GetReferralEntryHandler)

	// Create a new referrals entry
	protected.POST("/referrals", v1Service.CreateReferralEntryHandler)

	// Update a referral entry's status
	protected.PATCH("/referrals/:id/status", v1Service.UpdateReferralEntryStatusHandler)

	// Get the history logs
	protected.GET("/referrals/:id/logs", v1Service.ListReferralLogsHandler)

	// Get current user
	protected.GET("/users/me", v1Service.GetCurrentUserHandler)

	// Log out
	protected.POST("/logout", v1Service.LogoutHandler)

	// Change password
	protected.PATCH("/users/password", v1Service.ChangePasswordHandler)

	admin := protected.Group("")
	admin.Use(auth.AdminOnlyMiddleware)               // Add the extra gatekeeper
	admin.POST("/users", v1Service.CreateUserHandler) // Create a user
	admin.POST("/referrals/batch", v1Service.BatchCreateReferralEntriesHandler)
	admin.GET("/users", v1Service.ListUsersHandler)                      // List Users
	admin.DELETE("/users/:id", v1Service.DeleteUserHandler)              // Delete a user
	admin.DELETE("/referrals/:id", v1Service.DeleteReferralEntryHandler) // Delete a referral entry
	// admin.PATCH("/referrals/:id/status", v1Service.UpdateReferralEntryHandler) // Gotta change the URL?

	admin.POST("/tags", v1Service.CreateReferralTagHandler)
}
