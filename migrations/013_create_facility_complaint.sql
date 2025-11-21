CREATE TABLE IF NOT EXISTS facility_complaints (
    id UUID PRIMARY KEY,
    room_id TEXT NOT NULL,
    student_id UUID NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    description TEXT,
    proof TEXT,
    status VARCHAR(10) NOT NULL DEFAULT 'pending', -- pending|accepted|rejected
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
