package sqlite

import (
	"context"
	"reftrail/store"
	"testing"

	"reftrail/internal/domain"
)

func TestListReferralEntries_Filtering_TableDriven(t *testing.T) {
	// 1. Seed database once at the top
	s := setupTestStore(t)
	ctx := WithUserContext(context.Background(), &domain.UserContext{Username: "admin", Role: "REFTRAIL_ADMIN"})

	seedData := &store.BatchCreateReferralEntries{
		ReferralEntries: []store.CreateReferralEntry{
			{
				PatientFirstName: "Target", PatientLastName: "One",
				Urgency: "URGENT", Status: "READY_TO_BOOK", ReferralDate: "2026-04-15",
				Source:          "REGULAR",
				Complaints:      []store.ReferralComplaint{{BodyPart: "KNEE", Side: "LEFT"}, {BodyPart: "ANKLE", Side: "RIGHT"}},
				ConsultType:     "APP+LE",
				CreatorUsername: "admin",
			},
			{
				PatientFirstName: "Target", PatientLastName: "Two",
				Urgency: "URGENT", Status: "READY_TO_BOOK", ReferralDate: "2026-04-20",
				Source:          "REGULAR",
				Complaints:      []store.ReferralComplaint{{BodyPart: "ANKLE", Side: "BILATERAL"}},
				ConsultType:     "APP+LE",
				CreatorUsername: "admin",
			},
			{
				PatientFirstName: "Excluded", PatientLastName: "WrongUrgency",
				Urgency: "ELECTIVE", Status: "READY_TO_BOOK", ReferralDate: "2026-04-22",
				Source:          "REGULAR",
				Complaints:      []store.ReferralComplaint{{BodyPart: "KNEE", Side: "RIGHT"}},
				ConsultType:     "APP+LE",
				CreatorUsername: "admin",
			},
		},
	}

	if err := s.BatchCreateReferralEntries(ctx, seedData); err != nil {
		t.Fatalf("failed setup seeding data: %v", err)
	}

	// 2. Setup reusable variables for reference types
	dateFrom := "2026-04-01"
	dateTo := "2026-04-18"
	limitOne := 1
	offsetZero := 0

	// 3. Define the structural contract for a filtering scenario
	type testCase struct {
		name               string
		filter             *store.FindReferralEntry
		expectedPageLen    int
		expectedTotalCount int
		extraValidation    func(t *testing.T, res *store.PaginatedReferralEntries) // Optional custom hooks
	}

	// 4. Map out your filter matrices
	tests := []testCase{
		{
			name: "Multi-Select Filter: (Knee OR Ankle) + URGENT",
			filter: &store.FindReferralEntry{
				Urgencies: []domain.ReferralUrgency{"URGENT"},
				BodyParts: []string{"KNEE", "ANKLE"},
			},
			expectedPageLen:    2,
			expectedTotalCount: 2,
			extraValidation: func(t *testing.T, res *store.PaginatedReferralEntries) {
				for _, entry := range res.ReferralEntries {
					if entry.Urgency != "URGENT" {
						t.Errorf("Expected Urgency to be Urgent, but got %s", entry.Urgency)
					}
					if entry.PatientLastName == "One" && len(entry.Complaints) != 2 {
						t.Errorf("Stitching failure! Target One should have 2 complaints records, got %d", len(entry.Complaints))
					}
				}
			},
		},
		{
			name: "Date Range Boundary Constraints",
			filter: &store.FindReferralEntry{
				ReferralDateFrom: &dateFrom,
				ReferralDateTo:   &dateTo,
			},
			expectedPageLen:    1,
			expectedTotalCount: 1,
		},
		{
			name: "Pagination Offset and Limit Controls",
			filter: &store.FindReferralEntry{
				Limit:  &limitOne,
				Offset: &offsetZero,
			},
			expectedPageLen:    1, // Only 1 item returned on this slice page
			expectedTotalCount: 3, // Global table rows count stays 3
		},
	}

	// 5. Run the suite loop dynamically
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			paginated, err := s.ListReferralEntries(ctx, tc.filter)
			if err != nil {
				t.Fatalf("Store query execution failed: %v", err)
			}

			// Validate returned page slice size
			if len(paginated.ReferralEntries) != tc.expectedPageLen {
				t.Errorf("Slice length mismatch! Expected %d, found %d", tc.expectedPageLen, len(paginated.ReferralEntries))
			}

			// Validate unpaged global metadata counters
			if paginated.TotalCount != tc.expectedTotalCount {
				t.Errorf("TotalCount metric mismatch! Expected %d, received %d", tc.expectedTotalCount, paginated.TotalCount)
			}

			// Trigger custom validator hooks if they were supplied for the test case
			if tc.extraValidation != nil {
				tc.extraValidation(t, paginated)
			}
		})
	}
}
