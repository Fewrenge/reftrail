package domain

type ReferralID string
type ReferralStatus string
type ReferralUrgency string
type ReferralSource string

const (
	StatusReadyToBook     ReferralStatus = "READY_TO_BOOK"
	Status1stCallComplete ReferralStatus = "1ST_CALL_COMPLETE"
	Status2ndCallComplete ReferralStatus = "2ND_CALL_COMPLETE"
	Status3rdCallComplete ReferralStatus = "3RD_CALL_COMPLETE"
	StatusBooked          ReferralStatus = "BOOKED"
	StatusUnableToContact ReferralStatus = "UNABLE_TO_CONTACT"
	StatusPatientCallback ReferralStatus = "PATIENT_TO_CALL_BACK"
	StatusDeclined        ReferralStatus = "DECLINED"
	StatusSuspended       ReferralStatus = "SUSPENDED" // Temporary halt (e.g., patient is on vacation)
	StatusClosed          ReferralStatus = "CLOSED"    // Permanent end (Audit trail: patient died, moved, etc.)
)

// TransitionRule now only needs to track what the standard user can do
type TransitionRule struct {
	AllowedTo []ReferralStatus
}

// statusRules defines the standard workflow for the Booking Team
var statusRules = map[ReferralStatus]TransitionRule{
	StatusReadyToBook: {
		AllowedTo: []ReferralStatus{Status1stCallComplete, StatusUnableToContact, StatusDeclined, StatusBooked},
	},
	Status1stCallComplete: {
		AllowedTo: []ReferralStatus{Status2ndCallComplete, StatusUnableToContact, StatusDeclined, StatusBooked},
	},
	Status2ndCallComplete: {
		AllowedTo: []ReferralStatus{Status3rdCallComplete, StatusUnableToContact, StatusDeclined, StatusBooked},
	},
	Status3rdCallComplete: {
		AllowedTo: []ReferralStatus{StatusUnableToContact, StatusDeclined, StatusBooked},
	},
	StatusPatientCallback: {
		// From a callback, they can basically re-enter the call cycle or book
		AllowedTo: []ReferralStatus{Status1stCallComplete, StatusBooked, StatusDeclined},
	},
	StatusUnableToContact: {
		AllowedTo: []ReferralStatus{StatusReadyToBook}, // Allow them to try again
	},
	// StatusBooked and StatusDeclined are intentionally empty:
	// The Booking Team cannot move them once they are in these final statuses.
	StatusBooked:   {AllowedTo: []ReferralStatus{}},
	StatusDeclined: {AllowedTo: []ReferralStatus{}},
	StatusClosed:   {AllowedTo: []ReferralStatus{}},
}

func CanTransition(old, next ReferralStatus, role UserRole) bool {
	// 1. GOD MODE: Admins bypass all matrix rules
	if role == RoleReftrailAdmin {
		return true
	}

	// 2. Lookup the rules for the current status
	rule, ok := statusRules[old]
	if !ok {
		// If the status isn't in our map, we play it safe and block the move
		return false
	}

	// 3. Check if the Booking Team is allowed to make this move
	for _, s := range rule.AllowedTo {
		if s == next {
			return true
		}
	}

	return false
}
