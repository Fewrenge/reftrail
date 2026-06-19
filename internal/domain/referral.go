package domain

type ReferralID string
type ReferralStatus string
type ReferralUrgency string
type ReferralSource string
type ReferralConsultType string

const (
	UrgencyElective ReferralUrgency = "ELECTIVE"
	UrgencyUrgent   ReferralUrgency = "URGENT"
	UrgencyAsap     ReferralUrgency = "ASAP"
)

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

const (
	ConsultTypeAppLowerExtremity ReferralConsultType = "APP+LE"
	ConsultTypeAppUpperExtremity ReferralConsultType = "APP+UE"
	ConsultTypeAppAnySurgeon     ReferralConsultType = "APP+SX"
	ConsultTypeSurgeon           ReferralConsultType = "SX"
	ConsultTypeOther             ReferralConsultType = "OTHER"
)

// TransitionRule now only needs to track what the standard user can do
type TransitionRule struct {
	AllowedTo []ReferralStatus
}

// statusRules defines the standard workflow for the Booking Team
var statusRules = map[ReferralStatus]TransitionRule{
	StatusReadyToBook: {
		AllowedTo: []ReferralStatus{Status1stCallComplete, StatusUnableToContact, StatusDeclined, StatusBooked, StatusSuspended, StatusClosed},
	},
	Status1stCallComplete: {
		AllowedTo: []ReferralStatus{Status2ndCallComplete, StatusUnableToContact, StatusDeclined, StatusBooked, StatusSuspended, StatusClosed},
	},
	Status2ndCallComplete: {
		AllowedTo: []ReferralStatus{Status3rdCallComplete, StatusUnableToContact, StatusDeclined, StatusBooked, StatusSuspended, StatusClosed},
	},
	Status3rdCallComplete: {
		AllowedTo: []ReferralStatus{StatusUnableToContact, StatusDeclined, StatusBooked, StatusSuspended, StatusClosed},
	},
	StatusPatientCallback: {
		// From a callback, they can basically re-enter the call cycle or book
		AllowedTo: []ReferralStatus{StatusReadyToBook, Status1stCallComplete, StatusBooked, StatusDeclined, StatusSuspended, StatusClosed},
	},
	StatusUnableToContact: {
		AllowedTo: []ReferralStatus{StatusReadyToBook, StatusSuspended, StatusClosed}, // Allow them to try again
	},
	// StatusBooked and StatusDeclined are intentionally empty:
	// The Booking Team cannot move them once they are in these final statuses.
	StatusBooked:   {AllowedTo: []ReferralStatus{}},
	StatusDeclined: {AllowedTo: []ReferralStatus{}},
	StatusClosed:   {AllowedTo: []ReferralStatus{}},
}

var ImportDocumentHeaderSchema = map[string]bool{
	"last name":           true,
	"first name":          true,
	"complaint":           true,
	"complaint side":      true,
	"urgency":             true,
	"referral date":       true,
	"health card":         false,
	"date of birth":       false,
	"phone number":        false,
	"email":               false,
	"referring physician": false,
	"source":              false,
	"complaint details":   false,
	"consult type":        false,
	"triage note":         false,
	"tag":                 false,
	"emr patient id":      false,
	"emr referral doc id": false,
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

func ValidateCSVHeaders(headerMap map[string]int) []string {
	var missingFields []string
	for fieldName, isRequired := range ImportDocumentHeaderSchema {
		if isRequired {
			if _, exists := headerMap[fieldName]; !exists {
				missingFields = append(missingFields, fieldName)
			}
		}
	}
	return missingFields
}
