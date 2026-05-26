-- 1. User Table (Needed for your Login/Accountability)
CREATE TABLE IF NOT EXISTS user (
    username TEXT NOT NULL UNIQUE PRIMARY KEY,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'BOOKING_TEAM' check (role IN ('BOOKING_TEAM', 'REFTRAIL_ADMIN')),
    user_first_name TEXT,
    user_last_name TEXT,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE
);

-- 2. Referral Table (Requirement #1 through #10)
CREATE TABLE IF NOT EXISTS referral_entry (
    id TEXT PRIMARY KEY,
    created_ts TEXT NOT NULL,
    updated_ts TEXT NOT NULL, 
    creator_id TEXT NOT NULL,
    patient_last_name TEXT NOT NULL,
    patient_first_name TEXT NOT NULL,
    patient_dob TEXT NOT NULL,
    patient_healthcard_number TEXT NOT NULL,
    patient_healthcard_version_code TEXT NOT NULL,
    txt_customer_id TEXT,
    int_customer_doc_id INTEGER,
    referring_physician TEXT,
    consult_type TEXT CHECK(consult_type IN ('APP+LE','APP+UE','APP+SX','SX','OTHER')),
    consult_type_details TEXT, -- e.g. when patient has a preference
    triage_note TEXT,
    urgency TEXT CHECK(urgency IN ('Elective', 'Urgent', 'ASAP')),
    status TEXT NOT NULL DEFAULT 'READY_TO_BOOK' CHECK (status IN ('READY_TO_BOOK', '1ST_CALL_COMPLETE', '2ND_CALL_COMPLETE',
    '3RD_CALL_COMPLETE', 'BOOKED', 'UNABLE_TO_CONTACT', 'PATIENT_TO_CALL_BACK', 'DECLINED', 'SUSPENDED','CLOSED')),
    source TEXT CHECK(source IN ('REGULAR', 'FRACTURE_CLINIC', 'OTHER')),
    referral_date TEXT NOT NULL,
    FOREIGN KEY (creator_id) REFERENCES user(username) ON UPDATE CASCADE -- ON DELETE SET NULL?
);

-- 3. Audit Log (Requirement #9 - Tracking who changed the status)
CREATE TABLE IF NOT EXISTS referral_log (
    id TEXT PRIMARY KEY,
    referral_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    old_status TEXT,
    new_status TEXT,
    note TEXT,
    created_ts TEXT NOT NULL,
    FOREIGN KEY (referral_id) REFERENCES referral_entry(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(username) ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_referral_log_entry_id ON referral_log(referral_id);
CREATE INDEX IF NOT EXISTS idx_referral_healthcard ON referral_entry(patient_healthcard_number);

CREATE TABLE IF NOT EXISTS referral_appointment (
    id TEXT PRIMARY KEY,
    referral_id TEXT NOT NULL,
    complaint_target TEXT NOT NULL,
    appt_date_and_time TEXT,
    practitioner TEXT,
    juvonno_appt_id TEXT,
    created_ts TEXT NOT NULL,
    creator_id TEXT,
    FOREIGN KEY (referral_id) REFERENCES referral_entry(id) ON DELETE CASCADE
);

-- This table stores the actual body parts for each referral
CREATE TABLE IF NOT EXISTS referral_complaint (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    referral_id TEXT NOT NULL,
    body_part TEXT NOT NULL CHECK(body_part IN ('SHOULDER', 'KNEE', 'HIP', 'ELBOW', 'WRIST', 'ANKLE', 'FOOT', 'OTHER')),
    side TEXT NOT NULL CHECK(side IN ('LEFT', 'RIGHT', 'BILATERAL', 'OTHER')), -- OTHER is for cases where side is not applicable (e.g., "LOW BACK")
    details TEXT, -- For when body_part is 'OTHER' (e.g., "Femur")
    FOREIGN KEY (referral_id) REFERENCES referral_entry(id) ON DELETE CASCADE
);

-- Definition of Tags
-- Only Admin can edit Tags
-- Tags are for internal use to help categorize referrals (e.g. "X-Ray completed at hospital", "Online Booking Eligible", etc.)
CREATE TABLE IF NOT EXISTS referral_tag_definition (
    name TEXT PRIMARY KEY,
    description TEXT
);

-- Junction table (Many-to-Many)
CREATE TABLE IF NOT EXISTS referral_tag (
    referral_id TEXT NOT NULL,
    tag_name TEXT NOT NULL,
    PRIMARY KEY (referral_id, tag_name),
    FOREIGN KEY (referral_id) REFERENCES referral_entry(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_name) REFERENCES referral_tag_definition(name) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_referral_tag_ref ON referral_tag(referral_id);