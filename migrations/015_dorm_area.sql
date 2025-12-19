-- Bảng cho RegistrationPeriod
CREATE TABLE public.registration_periods (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    starttime TIMESTAMP NOT NULL,
    endtime TIMESTAMP NOT NULL,
    description TEXT,
    status VARCHAR NOT NULL
);

-- Bảng cho DormArea
CREATE TABLE public.dorm_areas (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    branch VARCHAR NOT NULL,
    address VARCHAR NOT NULL,
    fee DOUBLE PRECISION NOT NULL,
    description TEXT,
    image VARCHAR,
    status VARCHAR NOT NULL
);