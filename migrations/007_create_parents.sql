-- 7. Parent (Nhiều phụ huynh cho 1 sinh viên)
CREATE TABLE parents (
    id UUID PRIMARY KEY,
    student_id UUID REFERENCES students(id) ON DELETE CASCADE,
    type VARCHAR(10), -- Bố/Mẹ
    fullname VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    dob DATE,
    address VARCHAR(255)
);