-- Revert: drop the per-fund pre-trade preview toggle.

ALTER TABLE investment__funds DROP COLUMN IF EXISTS require_pretrade_preview;
