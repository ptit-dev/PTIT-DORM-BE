-- 8. Manager
CREATE TABLE managers (
    id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    fullname VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    cccd VARCHAR(20),
    dob DATE,
    avatar VARCHAR(255),
    province VARCHAR(100),
    commune VARCHAR(100),
    detail_address VARCHAR(255)
);