package sqlite

import (
	"context"
	"reftrail/store"
	"strings"
	"testing"
)

// Helper utility to quickly allocate string pointers for table test structures
func ptr(s string) *string {
	return &s
}

// 1. Test standard Store CRUD pipelines & validation rules
func TestReferralPhysician_CRUD_And_Validations(t *testing.T) {
	ctx := context.Background()

	t.Run("Create and Get Happy Path", func(t *testing.T) {
		s := setupTestStore(t)

		docPayload := &store.CreateReferralPhysician{
			CPSONumber:     ptr("12345"),
			FirstName:      "John",
			LastName:       "Smith",
			EMRPhysicianID: ptr("emr_id_99"),
		}

		// Verify creation returns a populated ID
		created, err := s.CreateReferralPhysician(ctx, docPayload)
		if err != nil {
			t.Fatalf("Expected successful physician registration, got: %v", err)
		}
		if created.ID == "" {
			t.Fatal("Expected system to generate a valid tracking UUID string, but received empty text")
		}

		// Verify we can pull the exact record back out by its ID
		fetched, err := s.GetReferralPhysicianByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("Expected to find registered physician, got error: %v", err)
		}
		if fetched.FirstName != "John" || fetched.LastName != "Smith" {
			t.Errorf("Data corruption: expected John Smith, fetched: %s %s", fetched.FirstName, fetched.LastName)
		}
	})

	t.Run("Validation enforces mandatory name fields", func(t *testing.T) {
		s := setupTestStore(t)

		badPayload := &store.CreateReferralPhysician{
			FirstName: "", // Empty names must be rejected by store level business validations
			LastName:  "Smith",
		}

		_, err := s.CreateReferralPhysician(ctx, badPayload)
		if err == nil {
			t.Fatal("Expected business validation error for empty first name, but operation succeeded silently")
		}
		if !strings.Contains(err.Error(), "mandatory fields") {
			t.Errorf("Expected signature name error statement, got instead: %v", err)
		}
	})

	t.Run("Update Happy Path", func(t *testing.T) {
		s := setupTestStore(t)

		created, _ := s.CreateReferralPhysician(ctx, &store.CreateReferralPhysician{
			FirstName: "Jane",
			LastName:  "Doe",
		})

		// Modify details
		var updated store.UpdateReferralPhysician
		updated.FirstName = ptr("Janet")
		updated.CPSONumber = ptr("99999")

		err := s.UpdateReferralPhysician(ctx, &updated)
		if err != nil {
			t.Fatalf("Expected clean update execution, got: %v", err)
		}

		fetched, _ := s.GetReferralPhysicianByID(ctx, created.ID)
		if fetched.FirstName != "Janet" || *fetched.CPSONumber != "99999" {
			t.Errorf("Update failed to apply to DB storage targets. Found: %s (%v)", fetched.FirstName, fetched.CPSONumber)
		}
	})
}

// 2. Test Foreign Key enforcement: SQLite must block deleting doctors who have patient referrals
func TestReferralPhysician_Delete_Constraints(t *testing.T) {
	ctx := context.Background()

	t.Run("Delete unreferenced physician succeeds cleanly", func(t *testing.T) {
		s := setupTestStore(t)

		created, _ := s.CreateReferralPhysician(ctx, &store.CreateReferralPhysician{
			FirstName: "Isolated",
			LastName:  "Doctor",
		})

		err := s.DeleteReferralPhysician(ctx, &store.DeleteReferralPhysician{ID: created.ID})
		if err != nil {
			t.Fatalf("Expected clean extraction drop for unlinked physician, encountered: %v", err)
		}
	})

	// ADD TEST CASE:
	// 1. Creates a physician
	// 2. Creates a referral_entry linked to that physician's ID
	// 3. Asserts that calling s.DeleteReferralPhysician returns your custom error:
	//    "cannot remove physician: doctor is actively linked to ongoing patient referral records"
}

// 3. Test Jaro-Winkler String Similarity scoring thresholds
func TestReferralPhysician_FuzzyMatching_Logic(t *testing.T) {
	tests := []struct {
		name          string
		stringA       string
		stringB       string
		expectMatch   bool
		minConfidence float64
	}{
		{
			name:          "Exact names ignore capitalization",
			stringA:       "Dr. John Smith",
			stringB:       "dr. john smith",
			expectMatch:   true,
			minConfidence: 1.0,
		},
		{
			name:          "High similarity with a single character typo",
			stringA:       "John Smith",
			stringB:       "John Smyth",
			expectMatch:   true,
			minConfidence: 0.85, // Jaro-Winkler scores this ~0.89
		},
		{
			name:          "Complete mismatch name entries",
			stringA:       "Gregory House",
			stringB:       "Jane Smith",
			expectMatch:   false,
			minConfidence: 0.50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := store.CalculateSimilarity(tt.stringA, tt.stringB)

			if tt.expectMatch && score < tt.minConfidence {
				t.Errorf("Expected high similarity score for close strings '%s' and '%s'. Got: %f (wanted >= %f)",
					tt.stringA, tt.stringB, score, tt.minConfidence)
			}
			if !tt.expectMatch && score >= 0.75 {
				t.Errorf("Expected low score for different strings '%s' and '%s'. Got an unexpectedly high score: %f",
					tt.stringA, tt.stringB, score)
			}
		})
	}
}
