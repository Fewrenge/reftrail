package sqlite

import (
	"context"
	"reftrail/internal/domain"
	"reftrail/store"
	"testing"
)

func TestReferralTag_Integration(t *testing.T) {
	// 1. Setup isolated memory DB
	s := setupTestStore(t)
	ctx := WithUserContext(context.Background(), &domain.UserContext{
		Username: "admin", Role: "REFTRAIL_ADMIN",
	})

	// 2. Setup: Pre-create a referral to tag
	ref, err := s.CreateReferralEntry(ctx, &store.CreateReferralEntry{
		PatientLastName:  "Tag Testing",
		PatientFirstName: "Patient",
		Urgency:          "ELECTIVE",
		ConsultType:      "APP+LE",
		Status:           "READY_TO_BOOK",
		Source:           "REGULAR",
		Complaints: []store.ReferralComplaint{
			{
				BodyPart: "KNEE",
				Side:     "LEFT",
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create referral for testing: %v", err)
	}

	t.Run("Full Lifecycle: Create, Assign, Remove, and Cascade Delete", func(t *testing.T) {
		// --- A. Create a Tag Definition ---
		tag, err := s.CreateReferralTag(ctx, &store.CreateReferralTag{
			Name:        "PEACH",
			Description: "X-ray not done",
		})
		if err != nil {
			t.Fatalf("Failed to create tag: %v", err)
		}

		// --- B. Assign Tag to Referral ---
		err = s.AssignTagToReferral(ctx, ref.ID, tag.Name)
		if err != nil {
			t.Errorf("Failed to assign tag: %v", err)
		}

		// --- C. Remove Tag from Referral ---
		err = s.RemoveTagFromReferral(ctx, ref.ID, tag.Name)
		if err != nil {
			t.Errorf("Failed to remove tag: %v", err)
		}

		// --- D. Test Cascade Delete ---
		// Re-assign tag first
		_ = s.AssignTagToReferral(ctx, ref.ID, tag.Name)

		// Delete the definition (Admin only)
		err = s.DeleteReferralTag(ctx, &store.DeleteReferralTag{Name: tag.Name})
		if err != nil {
			t.Errorf("Failed to delete tag definition: %v", err)
		}

	})
}
