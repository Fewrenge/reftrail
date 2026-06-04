package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"reftrail/internal/domain"
	"reftrail/store"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Import the sqlite driver
)

func setupTestStore(t *testing.T) *store.Store {
	// 1. Open a fresh in-memory database with random name
	b := make([]byte, 4)
	rand.Read(b)
	dbName := hex.EncodeToString(b)

	// Use the unique name in the DSN
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", dbName)
	db, err := sql.Open("sqlite3", dsn)

	t.Cleanup(func() {
		db.Close()
	})

	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	db.SetMaxOpenConns(1)

	// 2. Run schema
	schema := `
	CREATE TABLE IF NOT EXISTS user (username TEXT UNIQUE PRIMARY KEY, password_hash TEXT, role TEXT, user_first_name TEXT,
    user_last_name TEXT, is_archived BOOLEAN NOT NULL DEFAULT FALSE);
	CREATE TABLE IF NOT EXISTS referral_entry (
		id TEXT PRIMARY KEY, creator_id TEXT NOT NULL, created_ts TEXT, updated_ts TEXT,
		patient_last_name TEXT, patient_first_name TEXT, patient_dob TEXT, patient_healthcard_number TEXT, patient_healthcard_version_code TEXT,
		patient_phone_number TEXT, patient_email TEXT,
		emr_patient_id TEXT, emr_referral_doc_id INTEGER,
		referring_physician TEXT, triage_note TEXT, urgency TEXT CHECK(urgency IN ('Elective', 'Urgent', 'ASAP')), status TEXT, source TEXT, referral_date TEXT,
		FOREIGN KEY (creator_id) REFERENCES user(username)
	);
	CREATE TABLE IF NOT EXISTS referral_complaint (
		id INTEGER PRIMARY KEY AUTOINCREMENT, referral_id TEXT, body_part TEXT, side TEXT, details TEXT,
		FOREIGN KEY (referral_id) REFERENCES referral_entry(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS referral_tag_definition (
    name TEXT NOT NULL UNIQUE PRIMARY KEY, 
    description TEXT
	);
	CREATE TABLE IF NOT EXISTS referral_tag (
    referral_id TEXT, 
    tag_name TEXT,
    PRIMARY KEY (referral_id, tag_name),
    FOREIGN KEY (referral_id) REFERENCES referral_entry(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_name) REFERENCES referral_tag_definition(name) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS referral_log (
    id TEXT PRIMARY KEY,
    referral_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    old_status TEXT,
    new_status TEXT,
    note TEXT,
    created_ts TEXT NOT NULL,
    FOREIGN KEY (referral_id) REFERENCES referral_entry(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(username)
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
			PatientLastName:  "Test",
			PatientFirstName: "Gopher",
			Source:           "REGULAR",
			Complaints: []store.ReferralComplaint{
				{BodyPart: "KNEE", Side: "LEFT"},
			},
			Urgency:      "Elective",
			ReferralDate: "2023-10-01",
		}

		baseCtx := context.Background()

		// Mock the user context so the Store doesn't error out
		mockUser := &domain.UserContext{Username: "admin", Role: "REFTRAIL_ADMIN"}
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

func TestBatchCreateReferralEntries_Integration(t *testing.T) {
	s := setupTestStore(t)
	ctx := WithUserContext(context.Background(), &domain.UserContext{Username: "admin", Role: "REFTRAIL_ADMIN"})

	t.Run("Should successfully import multiple entries", func(t *testing.T) {
		batch := &store.BatchCreateReferralEntries{
			ReferralEntries: []store.CreateReferralEntry{
				{PatientLastName: "Alice", PatientFirstName: "One", Status: "READY_TO_BOOK", Urgency: "Elective"},
				{PatientLastName: "Bob", PatientFirstName: "Two", Status: "READY_TO_BOOK", Urgency: "Elective"},
			},
		}

		// 1. Run the batch
		err := s.BatchCreateReferralEntries(ctx, batch)
		if err != nil {
			t.Fatalf("batch failed: %v", err)
		}

		// 2. FIXED: Drill into the paginated structural return layout contract
		paginated, err := s.ListReferralEntries(ctx, &store.FindReferralEntry{})
		if err != nil {
			t.Fatalf("could not verify entries: %v", err)
		}

		// FIXED: Check lengths of inner array and counter tracking values
		if len(paginated.ReferralEntries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(paginated.ReferralEntries))
		}
		if paginated.TotalCount != 2 {
			t.Errorf("expected global count size tracking to equal 2, got %d", paginated.TotalCount)
		}
	})

	t.Run("Should rollback entire batch if urgency is invalid", func(t *testing.T) {
		// 1. Prepare a batch where the first is valid but the second has a bad Urgency
		batch := &store.BatchCreateReferralEntries{
			ReferralEntries: []store.CreateReferralEntry{
				{
					PatientLastName:  "I Should Be",
					PatientFirstName: "Rolled Back",
					Urgency:          "Elective", // Valid
					ReferralDate:     "2023-10-01",
				},
				{
					PatientLastName:  "I Am",
					PatientFirstName: "Invalid",
					Urgency:          "IMMEDIATELY", // INVALID! (Not Elective, Urgent, or ASAP)
				},
			},
		}

		// 2. Attempt the batch import
		err := s.BatchCreateReferralEntries(ctx, batch)
		if err == nil {
			t.Error("expected error due to invalid urgency CHECK constraint, but got nil")
		}

		// 3. VERIFY ROLLBACK
		firstName := "I Should Be"
		lastName := "Rolled Back"

		// FIXED: Drill into inner struct fields to check transaction cleanup safety
		paginated, err := s.ListReferralEntries(ctx, &store.FindReferralEntry{
			PatientFirstName: &firstName,
			PatientLastName:  &lastName,
		})

		if err == nil && len(paginated.ReferralEntries) > 0 {
			t.Errorf("Rollback failed! 'I Should Be Rolled Back' was found in the database despite the batch failing.")
		}
	})
}
