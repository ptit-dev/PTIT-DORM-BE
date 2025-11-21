
-- Bảng hợp đồng (Contracts)
CREATE TABLE IF NOT EXISTS contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    dorm_application_id UUID NOT NULL REFERENCES dorm_applications(id) ON DELETE CASCADE,
    room VARCHAR(50),
    status VARCHAR(20) NOT NULL, -- trạng thái duyệt: tạm thời, đã duyệt, đã hủy
    image_bill VARCHAR(255),
    monthly_fee NUMERIC(12,2),
    total_amount NUMERIC(14,2),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status_payment VARCHAR(10) NOT NULL DEFAULT 'unpaid', -- trạng thái thanh toán: đã, chưa
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    note TEXT
);

-- Bảng khu KTX (Dorm Areas)
CREATE TABLE IF NOT EXISTS dorm_areas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    branch VARCHAR(100) NOT NULL,
    address VARCHAR(255) NOT NULL,
    fee NUMERIC(12,2) NOT NULL,
    description TEXT,
    image VARCHAR(255),
    status VARCHAR(20) NOT NULL
);

-- Bảng đợt đăng ký (Registration Periods)
CREATE TABLE IF NOT EXISTS registration_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    starttime TIMESTAMP NOT NULL,
    endtime TIMESTAMP NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL
);
-- 1. User
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL,
    role_type VARCHAR(20),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2. Role
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 3. Permission
CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 4. UserRole (Many-to-Many)
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- 5. RolePermission (Many-to-Many)
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

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


CREATE TABLE dorm_applications (
    id UUID PRIMARY KEY,
    student_id VARCHAR(32) NOT NULL,
    full_name VARCHAR(128) NOT NULL,
    dob DATE,
    gender VARCHAR(16),
    cccd VARCHAR(32),
    cccd_issue_date DATE,
    cccd_issue_place VARCHAR(128),
    phone VARCHAR(32),
    email VARCHAR(128),
    avatar_front VARCHAR(255),
    avatar_back VARCHAR(255),
    class VARCHAR(32),
    course VARCHAR(32),
    faculty VARCHAR(64),
    ethnicity VARCHAR(32),
    religion VARCHAR(32),
    hometown VARCHAR(128),
    guardian_name VARCHAR(128),
    guardian_phone VARCHAR(32),
    priority_proof VARCHAR(255),
    preferred_site VARCHAR(64),
    preferred_dorm VARCHAR(32),
    priority_group VARCHAR(64),
    admission_type VARCHAR(32),
    status VARCHAR(16) DEFAULT 'pending',
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_dorm_applications_student_id ON dorm_applications(student_id);





