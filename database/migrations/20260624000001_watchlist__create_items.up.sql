CREATE TABLE watchlist_items (
    id              uuid        PRIMARY KEY,
    scope_type      text        NOT NULL
        CHECK (scope_type IN ('PERSONAL', 'PORTFOLIO')),
    owner_user_id   uuid,
    portfolio_id    uuid,
    security_id     uuid        NOT NULL,
    display_order   integer,
    pinned          boolean     NOT NULL DEFAULT false,
    note            text,
    status          text        NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'DISABLED')),
    created_by      uuid        NOT NULL,
    updated_by      uuid,
    deleted_by      uuid,
    created_at      timestamptz NOT NULL DEFAULT NOW(),
    updated_at      timestamptz NOT NULL DEFAULT NOW(),
    deleted_at      timestamptz,

    CONSTRAINT chk_watchlist_items_personal_owner
        CHECK (scope_type != 'PERSONAL' OR owner_user_id IS NOT NULL),
    CONSTRAINT chk_watchlist_items_personal_portfolio_null
        CHECK (scope_type != 'PERSONAL' OR portfolio_id IS NULL),
    CONSTRAINT chk_watchlist_items_portfolio_id
        CHECK (scope_type != 'PORTFOLIO' OR portfolio_id IS NOT NULL),
    CONSTRAINT chk_watchlist_items_portfolio_owner_null
        CHECK (scope_type != 'PORTFOLIO' OR owner_user_id IS NULL)
);

-- Active personal uniqueness: one active item per owner + security
CREATE UNIQUE INDEX uq_watchlist_items_personal_active
    ON watchlist_items (owner_user_id, security_id)
    WHERE scope_type = 'PERSONAL' AND deleted_at IS NULL;

-- Active portfolio uniqueness: one active item per portfolio + security
CREATE UNIQUE INDEX uq_watchlist_items_portfolio_active
    ON watchlist_items (portfolio_id, security_id)
    WHERE scope_type = 'PORTFOLIO' AND deleted_at IS NULL;

CREATE INDEX idx_watchlist_items_owner_status
    ON watchlist_items (owner_user_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_watchlist_items_portfolio_status
    ON watchlist_items (portfolio_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_watchlist_items_security
    ON watchlist_items (security_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_watchlist_items_scope_updated
    ON watchlist_items (scope_type, updated_at DESC)
    WHERE deleted_at IS NULL;
