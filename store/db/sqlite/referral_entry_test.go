package sqlite

import (
	"context"
	"reftrail/internal/domain"
	"reftrail/store"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Import the sqlite driver
)

func WithUserContext(ctx context.Context, u *domain.UserContext) context.Context {
	return context.WithValue(ctx, domain.UserKey, u)
}

func TestCreateReferralEntry_Scenarios(t *testing.T) {
	s := setupTestStore(t)

	// 1. Define the structural contract for a test scenario
	type testCase struct {
		name          string
		userContext   *domain.UserContext
		input         *store.CreateReferralEntry
		expectAnError bool
		errorContains string // Optional: check if the right error message is thrown
	}

	// 2. Map out every single execution path matrix
	tests := []testCase{
		{
			name:        "Happy Path of Valid entry with complaint",
			userContext: &domain.UserContext{Username: "admin", Role: "REFTRAIL_ADMIN"},
			input: &store.CreateReferralEntry{
				PatientLastName:  "Gopher",
				PatientFirstName: "Primary",
				Source:           "REGULAR",
				Urgency:          "ELECTIVE",
				Status:           "READY_TO_BOOK",
				ConsultType:      "APP+LE",
				ReferralDate:     "2026-06-17",
				Complaints: []store.ReferralComplaint{
					{BodyPart: "KNEE", Side: "LEFT"},
				},
			},
			expectAnError: false,
		},
		{
			name:        "Database Constraint Violation - Invalid Urgency Check",
			userContext: &domain.UserContext{Username: "admin", Role: "REFTRAIL_ADMIN"},
			input: &store.CreateReferralEntry{
				PatientLastName:  "Smith",
				PatientFirstName: "John",
				Source:           "REGULAR",
				Urgency:          "CRITICAL", // Breaks CHECK
				ConsultType:      "APP+LE",
				ReferralDate:     "2026-06-17",
			},
			expectAnError: true,
		},
		{
			name:        "Auth Failure - Missing User Context Token",
			userContext: nil, // Simulates an unauthenticated call crashing audit loops
			input: &store.CreateReferralEntry{
				PatientLastName:  "Doe",
				PatientFirstName: "Jane",
				Source:           "REGULAR",
				Urgency:          "URGENT",
				ReferralDate:     "2026-06-17",
			},
			expectAnError: true,
		},
		// ADD NEW SCENARIOS HERE AS SYSTEM AMENDMENTS EXTEND:
		// - e.g., "Validation - Missing PatientLastName"
		// - e.g., "Database Constraint - Duplicate Healthcard Number"
	}

	// 3. Loop through every matrix path dynamically
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange Context
			var ctx context.Context = context.Background()
			if tc.userContext != nil {
				ctx = WithUserContext(ctx, tc.userContext)
			}

			// Act
			entry, err := s.CreateReferralEntry(ctx, tc.input)

			// Assert
			if tc.expectAnError {
				if err == nil {
					t.Errorf("Expected an error to occur, but operation succeeded smoothly")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected execution success, but encountered unexpected error: %v", err)
				}
				if entry.ID == "" {
					t.Error("System tracking failure: Missing returned layout tracking identifier string")
				}
			}
		})
	}
}
func TestBatchCreateReferralEntries_TableDriven(t *testing.T) {
	// Setup a single test store instance for the table group
	s := setupTestStore(t)
	ctx := WithUserContext(context.Background(), &domain.UserContext{Username: "admin", Role: "REFTRAIL_ADMIN"})

	// 1. Define the structural contract for a batch scenario
	type testCase struct {
		name               string
		batch              *store.BatchCreateReferralEntries
		expectAnError      bool
		errorContains      string
		verifyRollbackName string // Optional: First name to search for to ensure database rollback worked
	}

	// 2. Map out success and failure matrices
	tests := []testCase{
		{
			name: "Success Path - Import multiple valid entries smoothly",
			batch: &store.BatchCreateReferralEntries{
				ReferralEntries: []store.CreateReferralEntry{
					{PatientLastName: "Alice", PatientFirstName: "One", Status: "READY_TO_BOOK", Urgency: "ELECTIVE", ConsultType: "APP+LE", Source: "REGULAR"},
					{PatientLastName: "Bob", PatientFirstName: "Two", Status: "READY_TO_BOOK", Urgency: "ELECTIVE", ConsultType: "APP+UE", Source: "REGULAR"},
				},
			},
			expectAnError: false,
		},
		{
			name: "Failure Path - Rollback entire batch if secondary entry breaks urgency constraint",
			batch: &store.BatchCreateReferralEntries{
				ReferralEntries: []store.CreateReferralEntry{
					{
						PatientLastName:  "Rolled Back",
						PatientFirstName: "IShouldBe",
						Urgency:          "ELECTIVE",
						ReferralDate:     "2023-10-01",
					},
					{
						PatientLastName:  "Invalid",
						PatientFirstName: "IAm",
						Urgency:          "IMMEDIATELY", // Breaks database CHECK constraints
					},
				},
			},
			expectAnError:      true,
			errorContains:      "CHECK constraint",
			verifyRollbackName: "IShouldBe",
		},
	}

	// 3. Loop through every matrix path dynamically
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act: Attempt the batch operation
			err := s.BatchCreateReferralEntries(ctx, tc.batch)

			// Assert: Validate structural expectations
			if tc.expectAnError {
				if err == nil {
					t.Errorf("Expected batch insertion to fail, but it returned a nil error value instead.")
				}

				// Optional: Check if the error text looks correct
				if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain %q, but got instead: %v", tc.errorContains, err)
				}

				// Transaction Verification Checklist: If a rollback name was specified, prove nothing hit the disk
				if tc.verifyRollbackName != "" {
					paginated, queryErr := s.ListReferralEntries(ctx, &store.FindReferralEntry{
						PatientFirstName: &tc.verifyRollbackName,
					})
					if queryErr == nil && len(paginated.ReferralEntries) > 0 {
						t.Errorf("Rollback Assertion Failure! Found entry matching '%s' in the database even though the batch transaction returned an error condition.", tc.verifyRollbackName)
					}
				}

			} else {
				// Asserting for success paths
				if err != nil {
					t.Fatalf("Expected successful batch process insertion sequence, but ran into unexpected error: %v", err)
				}

				// Deep Query Validation: Query database records directly to verify array insertion lengths
				paginated, queryErr := s.ListReferralEntries(ctx, &store.FindReferralEntry{})
				if queryErr != nil {
					t.Fatalf("Could not query database schema entries to verify insertion totals: %v", queryErr)
				}

				// Note: Since we are sharing the same database container across these subtests,
				// the first test adds 2 entries. Make sure your validation count aligns with expected totals.
				if len(paginated.ReferralEntries) < len(tc.batch.ReferralEntries) {
					t.Errorf("Expected at least %d total entries in database record sets, found only %d instead.", len(tc.batch.ReferralEntries), len(paginated.ReferralEntries))
				}
			}
		})
	}
}
