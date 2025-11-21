-- Lịch trực (mỗi lịch trực chỉ có 1 cán bộ quản túc)
CREATE TABLE IF NOT EXISTS duty_schedules (
    id UUID PRIMARY KEY,
    date DATE NOT NULL,
    area_id TEXT NOT NULL,
    staff_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    description TEXT
);
