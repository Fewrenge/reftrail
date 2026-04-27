-- 1. User Table (Needed for your Login/Accountability)
CREATE TABLE IF NOT EXISTS user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'BOOKING_TEAM'
);

-- 2. Waitlist Table (Requirement #1 through #10)
CREATE TABLE IF NOT EXISTS referral_entry (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_ts BIGINT NOT NULL,
    updated_ts BIGINT NOT NULL, 
    creator_id INTEGER NOT NULL,
    patient_name TEXT NOT NULL,
    patient_dob TEXT NOT NULL,
    txt_customer_id TEXT,
    int_customer_doc_id TEXT,
    referring_physician TEXT,
    complaint TEXT,
    triage_note TEXT,
    urgency TEXT CHECK(urgency IN ('Elective', 'Urgent', 'ASAP')),
    state TEXT NOT NULL DEFAULT 'READY_TO_BOOK',
    
    -- Appointment Info (Requirement #11)
    appt_date TEXT,
    appt_time TEXT,
    practitioner TEXT,
    juvonno_appt_id TEXT,
    
    FOREIGN KEY (creator_id) REFERENCES user(id)
);

-- 3. Audit Log (Requirement #9 - Tracking who changed the state)
CREATE TABLE IF NOT EXISTS referral_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    old_state TEXT,
    new_state TEXT,
    note TEXT,
    created_ts BIGINT NOT NULL,
    FOREIGN KEY (entry_id) REFERENCES referral_entry(id),
    FOREIGN KEY (user_id) REFERENCES user(id)
);
