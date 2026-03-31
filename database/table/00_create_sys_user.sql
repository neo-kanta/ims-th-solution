CREATE SCHEMA IF NOT EXISTS mwdp AUTHORIZATION postgres;

USE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Users table
CREATE TABLE IF NOT EXISTS mwdp.iam_users (
    id UUID PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    -- Audit
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP NULL,
    force_logout_at TIMESTAMP NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(50) DEFAULT 'system',
    updated_by VARCHAR(50) DEFAULT 'system'
);

-- 2. Sessions table
CREATE TABLE IF NOT EXISTS mwdp.iam_sessions (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL REFERENCES mwdp.iam_users (id) ON DELETE CASCADE,
    session_token UUID NOT NULL DEFAULT uuid_generate_v4 (),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(50) DEFAULT 'system',
);

-- 3. Table
CREATE TABLE org_employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    employee_code VARCHAR(50) NOT NULL,
    first_name_en VARCHAR(100) NOT NULL,
    last_name_en VARCHAR(100) NOT NULL,
    first_name_th VARCHAR(100),
    last_name_th VARCHAR(100),
    display_name VARCHAR(200) NOT NULL,
    department VARCHAR(100),
    title VARCHAR(100),
    reporting_to UUID REFERENCES org_employees (id),
    hire_date DATE,
    termination_date DATE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_org_employees_code UNIQUE (employee_code)
);