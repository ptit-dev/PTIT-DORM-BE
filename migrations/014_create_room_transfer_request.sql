CREATE TABLE IF NOT EXISTS room_transfer_requests (
    id UUID PRIMARY KEY,
    requester_user_id UUID NOT NULL,
    target_user_id UUID NOT NULL,
    target_room_id TEXT NOT NULL,
    transfer_time TIMESTAMP NOT NULL,
    reason TEXT,
    peer_confirm_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending|accepted|rejected
    manager_confirm_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending|accepted|rejected
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);