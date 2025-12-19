CREATE TABLE contracts (
    id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    dorm_application_id UUID NOT NULL,
    room VARCHAR,
    status VARCHAR NOT NULL,
    image_bill VARCHAR,
    monthly_fee DOUBLE PRECISION,
    total_amount DOUBLE PRECISION,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    status_payment VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    note TEXT,
    CONSTRAINT fk_contract_student FOREIGN KEY (student_id) REFERENCES students(id),
    CONSTRAINT fk_contract_application FOREIGN KEY (dorm_application_id) REFERENCES dorm_applications(id)
);