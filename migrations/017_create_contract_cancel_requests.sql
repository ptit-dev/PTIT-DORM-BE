CREATE TABLE IF NOT EXISTS contract_cancel_requests (
    id UUID PRIMARY KEY,
    contract_id UUID NOT NULL REFERENCES contracts(id),
    student_id UUID NOT NULL REFERENCES students(id),
    reason TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending|approved|rejected
    manager_note TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP
);