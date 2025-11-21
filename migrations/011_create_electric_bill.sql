CREATE TABLE IF NOT EXISTS electric_bills (
    id UUID PRIMARY KEY,
    room_id TEXT NOT NULL,
    month VARCHAR(7) NOT NULL, -- YYYY-MM
    prev_electric INT NOT NULL,
    curr_electric INT NOT NULL,
    amount INT NOT NULL,
    is_confirmed BOOLEAN DEFAULT FALSE,
    payment_status VARCHAR(10) NOT NULL DEFAULT 'unpaid', -- unpaid|paid
    payment_proof TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
