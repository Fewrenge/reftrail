package domain

import "errors"

var (
	ErrInvalidCredentials    = errors.New("invalid username or password")
	ErrUnauthorized          = errors.New("unauthorized: missing user context")
	ErrForbidden             = errors.New("forbidden: administrator privilege required")
	ErrUserArchived          = errors.New("the user is archived")
	ErrReferralEntryNotFound = errors.New("referral entry not found")
	ErrTagNotFound           = errors.New("referral tag not found")
	ErrUserNotFound          = errors.New("user account not found")
	ErrIllegalTransition     = errors.New("illegal transition")
	ErrDataValidationFailed  = errors.New("data provided failed validation rules")
)
