package sqlite

import (
	"context"
	"reftrail/store"
	"testing"

	"reftrail/internal/domain"
)

func TestListReferralEntries_Filtering(t *testing.T) {
	// 1. REUSING YOUR EXACT EXISTENT ENGINE HELPER
	s := setupTestStore(t)

	mockUser := &domain.UserContext{Username: "admin", Role: "REFTRAIL_ADMIN"}
	ctx := WithUserContext(context.Background(), mockUser)

	// 2. Prepare mock referrals payload via your public Batch interface
	seedData := &store.BatchCreateReferralEntries{
		ReferralEntries: []store.CreateReferralEntry{
			{
				PatientFirstName: "Target",
				PatientLastName:  "One",
				Urgency:          "Urgent",
				Status:           "READY_TO_BOOK",
				ReferralDate:     "2026-04-15",
				Complaints: []store.ReferralComplaint{
					{BodyPart: "KNEE", Side: "LEFT"},
					{BodyPart: "SHOULDER", Side: "RIGHT"},
				},
				CreatorUsername: "admin",
			},
			{
				PatientFirstName: "Target",
				PatientLastName:  "Two",
				Urgency:          "Urgent",
				Status:           "READY_TO_BOOK",
				ReferralDate:     "2026-04-20",
				Complaints: []store.ReferralComplaint{
					{BodyPart: "ANKLE", Side: "BILATERAL"},
				},
				CreatorUsername: "admin",
			},
			{
				PatientFirstName: "Excluded",
				PatientLastName:  "WrongUrgency",
				Urgency:          "Elective",
				Status:           "READY_TO_BOOK",
				ReferralDate:     "2026-04-22",
				Complaints: []store.ReferralComplaint{
					{BodyPart: "KNEE", Side: "RIGHT"},
				},
				CreatorUsername: "admin",
			},
		},
	}

	// 3. Populate your isolated DB instance
	if err := s.BatchCreateReferralEntries(ctx, seedData); err != nil {
		t.Fatalf("failed setup seeding data: %v", err)
	}

	// ========================================================================
	// TEST CASE 1: Multi-Select Filter (Knee OR Ankle) + Urgent
	// ========================================================================
	t.Run("Complex Multi-Filter Assertion", func(t *testing.T) {
		filter := &store.FindReferralEntry{
			Urgencies: []domain.ReferralUrgency{"Urgent"},
			BodyParts: []string{"KNEE", "ANKLE"},
		}

		// FIXED: paginated holds the *store.PaginatedReferralEntries container struct
		paginated, err := s.ListReferralEntries(ctx, filter)
		if err != nil {
			t.Fatalf("Store query execution failed: %v", err)
		}

		// FIXED: Check the actual list length inside the container struct
		if len(paginated.ReferralEntries) != 2 {
			t.Errorf("Expected exactly 2 matched entries, but found %d", len(paginated.ReferralEntries))
		}

		// FIXED: Check that totalCount across pagination boundaries says exactly 2 records
		if paginated.TotalCount != 2 {
			t.Errorf("Expected TotalCount to evaluate to 2, received %d", paginated.TotalCount)
		}

		for _, entry := range paginated.ReferralEntries {
			if entry.Urgency != "Urgent" {
				t.Errorf("Expected Urgency to be Urgent, but got %s", entry.Urgency)
			}

			if entry.PatientLastName == "One" && len(entry.Complaints) != 2 {
				t.Errorf("Stitching failure! Target One should have 2 complaints records, got %d", len(entry.Complaints))
			}
		}
	})

	// ========================================================================
	// TEST CASE 2: Date Boundary Intersect Check
	// ========================================================================
	t.Run("Date Range Boundary Constraints", func(t *testing.T) {
		dateFrom := "2026-04-01"
		dateTo := "2026-04-18"

		filter := &store.FindReferralEntry{
			ReferralDateFrom: &dateFrom,
			ReferralDateTo:   &dateTo,
		}

		paginated, err := s.ListReferralEntries(ctx, filter)
		if err != nil {
			t.Fatalf("Store date query execution failed: %v", err)
		}

		if len(paginated.ReferralEntries) != 1 {
			t.Errorf("Expected exactly 1 entry within range, but found %d", len(paginated.ReferralEntries))
		}

		if paginated.TotalCount != 1 {
			t.Errorf("Expected total date intersection window size count of 1, got %d", paginated.TotalCount)
		}
	})

	// ========================================================================
	// TEST CASE 3: Chunk Pagination Check (Limit & Offset)
	// ========================================================================
	t.Run("Pagination Offset and Limit Controls", func(t *testing.T) {
		limitValue := 1
		offsetValue := 0

		filter := &store.FindReferralEntry{
			Limit:  &limitValue,
			Offset: &offsetValue,
		}

		paginated, err := s.ListReferralEntries(ctx, filter)
		if err != nil {
			t.Fatalf("Store pagination query execution failed: %v", err)
		}

		// FIXED: The limit constraints ensure this SPECIFIC page returns exactly 1 item
		if len(paginated.ReferralEntries) != 1 {
			t.Errorf("Expected page size limit constraint of 1, received %d items", len(paginated.ReferralEntries))
		}

		// FIXED: But the TotalCount must still equal 3, because there are 3 total rows in the table!
		if paginated.TotalCount != 3 {
			t.Errorf("Expected global unpaged record counter matching 3, received %d", paginated.TotalCount)
		}
	})
}
