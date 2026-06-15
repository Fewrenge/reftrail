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

	protected.Use(auth.JWTMiddleware(s.Store))

	// Get all referral entries (with pagination, filtering, etc.)
	protected.GET("/referrals", v1Service.ListReferralEntriesHandler)

	// Get ONE specific referrals entry (The :id sniper)
	protected.GET("/referrals/:id", v1Service.GetReferralEntryHandler)

	// Update a referral entry's status
	protected.PATCH("/referrals/:id/status", v1Service.UpdateReferralEntryStatusHandler)

	// Add a log to a referral entry (for recording notes that don't correspond to a status change)
	protected.POST("/referrals/:id/logs", v1Service.CreateReferralLogHandler)

	// Get the history logs
	protected.GET("/referrals/:id/logs", v1Service.ListReferralLogsHandler)

	// Get current user
	protected.GET("/users/me", v1Service.GetCurrentUserHandler)

	// Log out
	protected.POST("/logout", v1Service.LogoutHandler)

	// Change own password (old password required)
	protected.PATCH("/users/me/password", v1Service.ChangeOwnPasswordHandler)

	// List tags
	protected.GET("/tags", v1Service.ListReferralTagsHandler)

	admin := protected.Group("")

	admin.Use(auth.AdminOnlyMiddleware) // Add the extra gatekeeper

	// ------ USER MANAGEMENT ------

	// Create a user
	admin.POST("/users", v1Service.CreateUserHandler)

	// Get all users
	admin.GET("/users", v1Service.ListUsersHandler)

	// Delete a user
	admin.DELETE("/users/:username", v1Service.DeleteUserHandler)

	// Archive a user (soft delete)
	// TODO: kick the user out if they're currently logged in? (Might require some sort of token blacklist or short-lived tokens with refresh tokens)
	admin.PUT("/users/:username/archive", v1Service.ArchiveUserHandler)

	// PUT /api/v1/users/:username/role
	admin.PUT("/users/:username/role", v1Service.UpdateUserRoleHandler)

	// Reset a user's password
	admin.PATCH("/users/:username/password", v1Service.ResetUserPasswordHandler)

	// Update a user's info (Username, first name, last name)
	admin.PATCH("/users/:username", v1Service.UpdateUserHandler)

	// ------ REFERRAL ENTRY MANAGEMENT ------

	// Create a new referral entry
	admin.POST("/referrals", v1Service.CreateReferralEntryHandler)

	// Batch create referral entries
	admin.POST("/referrals/batch", v1Service.BatchCreateReferralEntriesHandler)

	// Delete a referral entry
	admin.DELETE("/referrals/:id", v1Service.DeleteReferralEntryHandler)

	// Update a referral entry (admin use only, for miscellaneous updates like correcting a typo, changing urgency, etc.)
	admin.PATCH("/referrals/:id", v1Service.UpdateReferralEntryHandler)

	// ------ TAG MANAGEMENT ------

	admin.POST("/tags", v1Service.CreateReferralTagHandler) // Add a tag to the database

	admin.PATCH("/tags/:id", v1Service.UpdateReferralTagDefinitionHandler)

	admin.DELETE("/tags/:id", v1Service.DeleteReferralTagHandler) // Delete a tag from the database

	admin.POST("/referrals/:id/tags/:tagName", v1Service.AssignTagHandler) // Assign a tag to a referral

	admin.DELETE("/referrals/:id/tags/:tagName", v1Service.RemoveTagHandler) // Remove a tag from a referral

	// ------ ANALYTICS ------

	// Get urgency distribution for pie chart
	admin.GET("/analytics/urgency-distribution", v1Service.GetUrgencyDistributionAnalyticsHandler)

	admin.GET("/analytics/referral-trend", v1Service.GetReferralVolumeAnalyticsHandler)

	admin.GET("/analytics/direct-booking-waiting-time", v1Service.GetDirectBookingWaitingTimeAnalyticsHandler)
}
