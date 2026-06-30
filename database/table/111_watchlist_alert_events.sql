-- Table: watchlist_alert_events
-- Source: 20260624000003_watchlist__create_alert_events.up.sql
CREATE TABLE watchlist_alert_events (
    id                   uuid        PRIMARY KEY,
    watchlist_item_id    uuid        NOT NULL,
    threshold_rule_id    uuid        NOT NULL,
    scope_type           text        NOT NULL
        CHECK (scope_type IN ('PERSONAL', 'PORTFOLIO')),
    owner_user_id        uuid,
    portfolio_id         uuid,
    security_id          uuid        NOT NULL,
    created_by_user_id   uuid        NOT NULL,
    direction            text        NOT NULL
        CHECK (direction IN ('ABOVE', 'BELOW')),
    previous_state       text        NOT NULL,
    current_state        text        NOT NULL
        CHECK (current_state = 'BREACHED'),
    observed_price       numeric(24,8) NOT NULL,
    threshold_value      numeric(24,8) NOT NULL,
    currency             text,
    quote_provider       text,
    observed_at          timestamptz NOT NULL,
    evaluated_at         timestamptz NOT NULL,
    stale                boolean     NOT NULL DEFAULT false,
    stale_reason         text,
    idempotency_key      text        NOT NULL,
    notification_status  text        NOT NULL DEFAULT 'PENDING'
        CHECK (notification_status IN ('PENDING', 'CREATED', 'SUPPRESSED', 'FAILED', 'SKIPPED')),
    notification_id      uuid,
    notification_error   text,
    acknowledged_by      uuid,
    acknowledged_at      timestamptz,
    acknowledgement_note text,
    created_at           timestamptz NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_watchlist_alert_events_idempotency
        UNIQUE (idempotency_key),
    CONSTRAINT chk_watchlist_alert_ack_consistent
        CHECK ((acknowledged_by IS NULL) = (acknowledged_at IS NULL))
);
