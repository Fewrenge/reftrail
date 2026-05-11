package sqlite

import (
	"context"
	"database/sql"
	"reftrail/internal/domain"
	"reftrail/store"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Import the sqlite driver
)

func setupTestStore(t *testing.T) *store.Store {
	// 1. Open a fresh in-memory database
	// ":memory:" tells SQLite not to save a file to disk
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// 2. Run schema
	schema := `
	CREATE TABLE user (id TEXT PRIMARY KEY, username TEXT, password_hash TEXT, role TEXT);
	CREATE TABLE referral_entry (
		id TEXT PRIMARY KEY, creator_id INTEGER, created_ts TEXT, updated_ts TEXT,
		patient_name TEXT, patient_dob TEXT, txt_customer_id TEXT, int_customer_doc_id TEXT,
		referring_physician TEXT, triage_note TEXT, urgency TEXT, status TEXT, source TEXT
	);
	CREATE TABLE referral_complaint (
		id INTEGER PRIMARY KEY AUTOINCREMENT, referral_id TEXT, body_part TEXT, side TEXT, details TEXT
	);`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// 3. Initialize real Driver and Store using this memory DB
	driver := NewWithDB(db)
	return store.NewStore(driver)
}

func WithUserContext(ctx context.Context, u *domain.UserContext) context.Context {
	return context.WithValue(ctx, domain.UserKey, u)
}

func TestCreateReferralEntry_Integration(t *testing.T) {
	s := setupTestStore(t)

	t.Run("Should save entry and complaints in a transaction", func(t *testing.T) {
		req := &store.CreateReferralEntry{
			PatientName: "Test Gopher",
			Source:      "REGULAR",
			Complaints: []store.ReferralComplaint{
				{BodyPart: "KNEE", Side: "LEFT"},
			},
		}

		baseCtx := context.Background()

		// Mock the user context so the Store doesn't error out
		mockUser := &domain.UserContext{ID: 1, Role: "REFTRAIL_ADMIN"}
		ctx := WithUserContext(baseCtx, mockUser)

		// 3. Run the store method using the context we just built
		entry, err := s.CreateReferralEntry(ctx, req)

		// 4. Use t.Fatal or t.Error instead of "return err"
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if entry.ID == "" {
			t.Error("expected a generated ID")
		}
	})

}
