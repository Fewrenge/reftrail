package domain

import "errors"

var (
	ErrInvalidCredentials    = errors.New("invalid username or password")
	ErrUnauthorized          = errors.New("unauthorized: missing user context")
	ErrForbidden             = errors.New("forbidden: administrator privilege required")
	ErrUserArchived          = errors.New("the user is archived")
	ErrReferralEntryNotFound = errors.New("referral entry not found")
	ErrTagNotFound           = errors.New("referral tag not found")
	ErrIllegalTransition     = errors.New("illegal transition")
	ErrDataValidationFailed  = errors.New("data provided failed validation rules")
	ErrUserNotFound          = errors.New("user not found")
	ErrSelfResetBlocked      = errors.New("admins cannot reset their own password via override")
	ErrPasswordMismatch      = errors.New("incorrect current password")
	ErrCannotArchiveSelf     = errors.New("you cannot archive your own account")
	ErrLastAdminLockout      = errors.New("cannot demote, archive, or delete the only remaining active admin")
)
