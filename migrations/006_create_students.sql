-- 6. Student (1-1 với User)
CREATE TABLE students (
    id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    fullname VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    cccd VARCHAR(20),
    dob DATE,
    avatar VARCHAR(255),
    province VARCHAR(100),
    commune VARCHAR(100),
    detail_address VARCHAR(255),
    type VARCHAR(50),
    course VARCHAR(50),
    major VARCHAR(100),
    class VARCHAR(50)
);