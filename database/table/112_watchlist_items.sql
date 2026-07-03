-- Table: watchlist_items
-- Source: 20260624000001_watchlist__create_items.up.sql
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
