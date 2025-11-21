CREATE TABLE IF NOT EXISTS electric_bill_complaints (
    id UUID PRIMARY KEY,
    student_id UUID NOT NULL REFERENCES users(id),
    electric_bill_id UUID NOT NULL REFERENCES electric_bills(id),
    note TEXT,
    proof TEXT,
    status VARCHAR(10) NOT NULL DEFAULT 'pending', -- pending|accepted|rejected
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
