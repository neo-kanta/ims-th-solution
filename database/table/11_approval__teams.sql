-- Table: approval__teams
-- Source: 20260529000001_approval__create_tables.up.sql
CREATE TABLE approval__teams (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    team_code           VARCHAR(60)  NOT NULL,
    team_name           VARCHAR(160) NOT NULL,
    remarks             TEXT         NOT NULL DEFAULT '',
    has_co_manager      BOOLEAN      NOT NULL DEFAULT false,
    min_required_stamps INTEGER      NOT NULL DEFAULT 1,
    max_allowed_stamps  INTEGER      NOT NULL DEFAULT 1,
    is_active           BOOLEAN      NOT NULL DEFAULT true,
    created_by          UUID         REFERENCES iam_users(id),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by          UUID         REFERENCES iam_users(id),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_teams_code UNIQUE (team_code),
    CONSTRAINT chk_approval_teams_stamps CHECK (
        min_required_stamps >= 1 AND max_allowed_stamps >= min_required_stamps
    )
);
