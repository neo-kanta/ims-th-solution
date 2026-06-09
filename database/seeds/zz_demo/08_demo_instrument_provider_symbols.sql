-- =============================================================================
-- Demo seed — Instrument provider-symbol mappings (Thai SET equities)
-- =============================================================================
-- Backfills provider_symbol_yahoo on the seeded Thai equities so the intraday
-- valuation service can resolve internal tickers (AOT, CPALL, ...) to the
-- Yahoo Finance symbols (AOT.BK, CPALL.BK, ...).
--
-- Alpha Vantage doesn't carry SET equities on its free tier; we leave that
-- column NULL and let the provider order fall through to Yahoo at runtime.
--
-- Idempotent — re-running refreshes mappings without touching unrelated rows.
-- =============================================================================

BEGIN;

UPDATE investment__instruments
   SET provider_symbol_yahoo = v.yahoo_symbol,
       updated_at            = NOW()
  FROM (VALUES
    ('PTT',     'PTT.BK'),
    ('KBANK',   'KBANK.BK'),
    ('AOT',     'AOT.BK'),
    ('CPALL',   'CPALL.BK'),
    ('SCB',     'SCB.BK')
  ) AS v(primary_ticker, yahoo_symbol)
 WHERE investment__instruments.primary_ticker = v.primary_ticker
   AND investment__instruments.deleted_at IS NULL;

COMMIT;
