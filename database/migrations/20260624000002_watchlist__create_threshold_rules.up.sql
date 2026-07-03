CREATE TABLE watchlist_threshold_rules (
    id                   uuid        PRIMARY KEY,
    watchlist_item_id    uuid        NOT NULL REFERENCES watchlist_items(id),
    metric_type          text        NOT NULL DEFAULT 'MARKET_PRICE'
        CHECK (metric_type = 'MARKET_PRICE'),
    direction            text        NOT NULL
        CHECK (direction IN ('ABOVE', 'BELOW')),
    threshold_value      numeric(24,8) NOT NULL
        CHECK (threshold_value > 0),
    currency             text,
    cooldown_minutes     integer     NOT NULL DEFAULT 60
        CHECK (cooldown_minutes >= 0),
    status               text        NOT NULL DEFAULT 'ENABLED'
        CHECK (status IN ('ENABLED', 'DISABLED')),
    last_state           text        NOT NULL DEFAULT 'UNKNOWN'
        CHECK (last_state IN ('UNKNOWN', 'NON_BREACHED', 'BREACHED')),
    last_observed_price  numeric(24,8),
    last_observed_at     timestamptz,
    last_evaluated_at    timestamptz,
    last_state_changed_at timestamptz,
    last_alerted_at      timestamptz,
    last_quote_stale     boolean     NOT NULL DEFAULT false,
    last_stale_reason    text,
    created_by           uuid        NOT NULL,
    updated_by           uuid,
    deleted_by           uuid,
    created_at           timestamptz NOT NULL DEFAULT NOW(),
    updated_at           timestamptz NOT NULL DEFAULT NOW(),
    deleted_at           timestamptz
);

-- Prevent duplicate active rules for same item + metric + direction + value
CREATE UNIQUE INDEX uq_watchlist_rules_active
    ON watchlist_threshold_rules (watchlist_item_id, metric_type, direction, threshold_value)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_watchlist_rules_evaluator
    ON watchlist_threshold_rules (status, last_evaluated_at)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_watchlist_rules_item
    ON watchlist_threshold_rules (watchlist_item_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_watchlist_rules_state_status
    ON watchlist_threshold_rules (last_state, status)
    WHERE deleted_at IS NULL;
