-- 1. User Table (Needed for your Login/Accountability)
CREATE TABLE IF NOT EXISTS user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'BOOKING_TEAM' check (role IN ('BOOKING_TEAM', 'REFTRAIL_ADMIN'))
);

-- 2. Referral Table (Requirement #1 through #10)
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
    triage_note TEXT,
    urgency TEXT CHECK(urgency IN ('Elective', 'Urgent', 'ASAP')),
    status TEXT NOT NULL DEFAULT 'READY_TO_BOOK' CHECK (status IN ('READY_TO_BOOK', '1ST_CALL_COMPLETE', '2ND_CALL_COMPLETE',
    '3RD_CALL_COMPLETE', 'BOOKED', 'UNABLE_TO_CONTACT', 'PATIENT_TO_CALL_BACK', 'DECLINED', 'SUSPENDED','CLOSED')),
    
    -- Source Info
    source TEXT CHECK(source IN ('REGULAR', 'FRACTURE_CLINIC', 'OTHER')),
    
    FOREIGN KEY (creator_id) REFERENCES user(id)
);

-- 3. Audit Log (Requirement #9 - Tracking who changed the status)
CREATE TABLE IF NOT EXISTS referral_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    old_status TEXT,
    new_status TEXT,
    note TEXT,
    created_ts BIGINT NOT NULL,
    FOREIGN KEY (entry_id) REFERENCES referral_entry(id),
    FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE INDEX IF NOT EXISTS idx_referral_log_entry_id ON referral_log(entry_id);

CREATE TABLE IF NOT EXISTS appointment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    referral_id INTEGER NOT NULL,
    complaint_target TEXT NOT NULL,
    appt_date TEXT,
    appt_time TEXT,
    practitioner TEXT,
    juvonno_appt_id TEXT,
    created_ts BIGINT NOT NULL,
    creator_id INTEGER,
    FOREIGN KEY (referral_id) REFERENCES referral_entry(id)
);

-- This table stores the actual body parts for each referral
CREATE TABLE IF NOT EXISTS referral_complaint (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    referral_id INTEGER NOT NULL,
    body_part TEXT NOT NULL CHECK(body_part IN ('SHOULDER', 'KNEE', 'HIP', 'ELBOW', 'WRIST', 'ANKLE', 'FOOT', 'OTHER')),
    side TEXT NOT NULL CHECK(side IN ('LEFT', 'RIGHT', 'BILATERAL')),
    details TEXT, -- For when body_part is 'OTHER' (e.g., "Femur")
    FOREIGN KEY (referral_id) REFERENCES referral_entry(id)
);