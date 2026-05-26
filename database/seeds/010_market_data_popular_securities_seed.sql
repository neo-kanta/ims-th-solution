-- =============================================================================
-- Popular securities seed
-- =============================================================================
-- Pre-populates securities_master with a curated list of recognisable global
-- equities (AAPL, MSFT, …), Thai SET equities (KBANK.BK, PTT.BK, …), popular
-- ETFs (SPY, QQQ), and FX pairs (USDTHB, EURUSD, …), plus Yahoo Finance
-- provider mappings so:
--
--   * /api/v1/market-data/screen/watchlist returns recognisable rows the
--     moment the app starts (no manual setup required),
--   * the /market-data/securities/{id} detail page works for any seeded row,
--   * "Sync quote" reaches a real provider for the global tickers.
--
-- Idempotent: re-running this file refreshes display_symbol / name / metadata
-- and never duplicates rows. Tables involved have NO CREATE RULE statements,
-- so ON CONFLICT is safe (cf. the compliance versions table).
-- =============================================================================

BEGIN;

INSERT INTO securities_master (
    ims_symbol, display_symbol, name, asset_type, currency, country_code,
    exchange_mic, isin, status
) VALUES
    -- ---- US large-cap technology (NASDAQ, XNAS) -----------------------------
    ('US_EQ_XNAS_AAPL',  'AAPL',  'Apple Inc.',                       'EQUITY', 'USD', 'US', 'XNAS', 'US0378331005', 'ACTIVE'),
    ('US_EQ_XNAS_MSFT',  'MSFT',  'Microsoft Corporation',            'EQUITY', 'USD', 'US', 'XNAS', 'US5949181045', 'ACTIVE'),
    ('US_EQ_XNAS_GOOGL', 'GOOGL', 'Alphabet Inc. Class A',            'EQUITY', 'USD', 'US', 'XNAS', 'US02079K3059', 'ACTIVE'),
    ('US_EQ_XNAS_AMZN',  'AMZN',  'Amazon.com Inc.',                  'EQUITY', 'USD', 'US', 'XNAS', 'US0231351067', 'ACTIVE'),
    ('US_EQ_XNAS_NVDA',  'NVDA',  'NVIDIA Corporation',               'EQUITY', 'USD', 'US', 'XNAS', 'US67066G1040', 'ACTIVE'),
    ('US_EQ_XNAS_META',  'META',  'Meta Platforms Inc.',              'EQUITY', 'USD', 'US', 'XNAS', 'US30303M1027', 'ACTIVE'),
    ('US_EQ_XNAS_TSLA',  'TSLA',  'Tesla Inc.',                       'EQUITY', 'USD', 'US', 'XNAS', 'US88160R1014', 'ACTIVE'),
    ('US_EQ_XNAS_NFLX',  'NFLX',  'Netflix Inc.',                     'EQUITY', 'USD', 'US', 'XNAS', 'US64110L1061', 'ACTIVE'),

    -- ---- US large-cap (NYSE, XNYS) ------------------------------------------
    ('US_EQ_XNYS_BRK_B', 'BRK-B', 'Berkshire Hathaway Inc. Class B',  'EQUITY', 'USD', 'US', 'XNYS', 'US0846707026', 'ACTIVE'),
    ('US_EQ_XNYS_JPM',   'JPM',   'JPMorgan Chase & Co.',             'EQUITY', 'USD', 'US', 'XNYS', 'US46625H1005', 'ACTIVE'),
    ('US_EQ_XNYS_V',     'V',     'Visa Inc.',                        'EQUITY', 'USD', 'US', 'XNYS', 'US92826C8394', 'ACTIVE'),
    ('US_EQ_XNYS_JNJ',   'JNJ',   'Johnson & Johnson',                'EQUITY', 'USD', 'US', 'XNYS', 'US4781601046', 'ACTIVE'),
    ('US_EQ_XNYS_WMT',   'WMT',   'Walmart Inc.',                     'EQUITY', 'USD', 'US', 'XNYS', 'US9311421039', 'ACTIVE'),

    -- ---- Popular ETFs (NYSE Arca / NASDAQ) ----------------------------------
    ('US_ETF_ARCX_SPY',  'SPY',   'SPDR S&P 500 ETF Trust',           'ETF',    'USD', 'US', 'ARCX', 'US78462F1030', 'ACTIVE'),
    ('US_ETF_XNAS_QQQ',  'QQQ',   'Invesco QQQ Trust',                'ETF',    'USD', 'US', 'XNAS', 'US46090E1038', 'ACTIVE'),
    ('US_ETF_ARCX_VTI',  'VTI',   'Vanguard Total Stock Market ETF',  'ETF',    'USD', 'US', 'ARCX', 'US9229087690', 'ACTIVE'),

    -- ---- Thai SET / XBKK ----------------------------------------------------
    ('TH_EQ_XBKK_KBANK',   'KBANK.BK',  'Kasikornbank PCL',                 'EQUITY', 'THB', 'TH', 'XBKK', NULL, 'ACTIVE'),
    ('TH_EQ_XBKK_PTT',     'PTT.BK',    'PTT PCL',                          'EQUITY', 'THB', 'TH', 'XBKK', NULL, 'ACTIVE'),
    ('TH_EQ_XBKK_CPALL',   'CPALL.BK',  'CP ALL PCL',                       'EQUITY', 'THB', 'TH', 'XBKK', NULL, 'ACTIVE'),
    ('TH_EQ_XBKK_SCC',     'SCC.BK',    'Siam Cement PCL',                  'EQUITY', 'THB', 'TH', 'XBKK', NULL, 'ACTIVE'),
    ('TH_EQ_XBKK_BBL',     'BBL.BK',    'Bangkok Bank PCL',                 'EQUITY', 'THB', 'TH', 'XBKK', NULL, 'ACTIVE'),
    ('TH_EQ_XBKK_AOT',     'AOT.BK',    'Airports of Thailand PCL',         'EQUITY', 'THB', 'TH', 'XBKK', NULL, 'ACTIVE'),
    ('TH_EQ_XBKK_ADVANC',  'ADVANC.BK', 'Advanced Info Service PCL',        'EQUITY', 'THB', 'TH', 'XBKK', NULL, 'ACTIVE'),
    ('TH_EQ_XBKK_DELTA',   'DELTA.BK',  'Delta Electronics (Thailand) PCL', 'EQUITY', 'THB', 'TH', 'XBKK', NULL, 'ACTIVE'),

    -- ---- FX pairs -----------------------------------------------------------
    ('FX_USDTHB', 'USDTHB', 'US Dollar / Thai Baht',  'FX', 'USD', NULL, NULL, NULL, 'ACTIVE'),
    ('FX_EURUSD', 'EURUSD', 'Euro / US Dollar',       'FX', 'EUR', NULL, NULL, NULL, 'ACTIVE'),
    ('FX_USDJPY', 'USDJPY', 'US Dollar / Japanese Yen','FX','USD', NULL, NULL, NULL, 'ACTIVE'),
    ('FX_GBPUSD', 'GBPUSD', 'British Pound / US Dollar','FX','GBP', NULL, NULL, NULL, 'ACTIVE')
ON CONFLICT (ims_symbol) DO UPDATE SET
    display_symbol = EXCLUDED.display_symbol,
    name           = EXCLUDED.name,
    asset_type     = EXCLUDED.asset_type,
    currency       = EXCLUDED.currency,
    country_code   = COALESCE(EXCLUDED.country_code, securities_master.country_code),
    exchange_mic   = COALESCE(EXCLUDED.exchange_mic, securities_master.exchange_mic),
    isin           = COALESCE(EXCLUDED.isin,         securities_master.isin),
    status         = EXCLUDED.status,
    updated_at     = NOW();

-- =============================================================================
-- Yahoo Finance provider mappings
-- =============================================================================
-- Yahoo is the cheapest "no API key" provider, so we register a mapping for
-- every seeded security here. Alpha Vantage mappings are intentionally omitted
-- (the global tickers are the same string anyway; the platform falls back to
-- the canonical display_symbol when no explicit mapping exists for a
-- provider).
-- =============================================================================

INSERT INTO security_provider_mappings (
    security_id, provider_code, provider_symbol, provider_exchange,
    provider_asset_type, provider_currency, priority, mapping_status, is_primary
)
SELECT sm.id,
       m.provider_code,
       m.provider_symbol,
       m.provider_exchange,
       m.provider_asset_type,
       m.provider_currency,
       m.priority,
       'ACTIVE'::varchar,
       m.is_primary
  FROM securities_master sm
  JOIN (VALUES
    -- US tech (Yahoo uses the bare ticker on NASDAQ)
    ('US_EQ_XNAS_AAPL',  'yahoo', 'AAPL',    'NASDAQ', 'EQUITY', 'USD', 10, true),
    ('US_EQ_XNAS_MSFT',  'yahoo', 'MSFT',    'NASDAQ', 'EQUITY', 'USD', 10, true),
    ('US_EQ_XNAS_GOOGL', 'yahoo', 'GOOGL',   'NASDAQ', 'EQUITY', 'USD', 10, true),
    ('US_EQ_XNAS_AMZN',  'yahoo', 'AMZN',    'NASDAQ', 'EQUITY', 'USD', 10, true),
    ('US_EQ_XNAS_NVDA',  'yahoo', 'NVDA',    'NASDAQ', 'EQUITY', 'USD', 10, true),
    ('US_EQ_XNAS_META',  'yahoo', 'META',    'NASDAQ', 'EQUITY', 'USD', 10, true),
    ('US_EQ_XNAS_TSLA',  'yahoo', 'TSLA',    'NASDAQ', 'EQUITY', 'USD', 10, true),
    ('US_EQ_XNAS_NFLX',  'yahoo', 'NFLX',    'NASDAQ', 'EQUITY', 'USD', 10, true),

    -- US NYSE (Yahoo uses BRK-B with a hyphen, not BRK.B)
    ('US_EQ_XNYS_BRK_B', 'yahoo', 'BRK-B',   'NYSE',   'EQUITY', 'USD', 10, true),
    ('US_EQ_XNYS_JPM',   'yahoo', 'JPM',     'NYSE',   'EQUITY', 'USD', 10, true),
    ('US_EQ_XNYS_V',     'yahoo', 'V',       'NYSE',   'EQUITY', 'USD', 10, true),
    ('US_EQ_XNYS_JNJ',   'yahoo', 'JNJ',     'NYSE',   'EQUITY', 'USD', 10, true),
    ('US_EQ_XNYS_WMT',   'yahoo', 'WMT',     'NYSE',   'EQUITY', 'USD', 10, true),

    -- ETFs
    ('US_ETF_ARCX_SPY',  'yahoo', 'SPY',     'NYSEARCA', 'ETF',  'USD', 10, true),
    ('US_ETF_XNAS_QQQ',  'yahoo', 'QQQ',     'NASDAQ',   'ETF',  'USD', 10, true),
    ('US_ETF_ARCX_VTI',  'yahoo', 'VTI',     'NYSEARCA', 'ETF',  'USD', 10, true),

    -- Thai SET (Yahoo uses the .BK suffix)
    ('TH_EQ_XBKK_KBANK',  'yahoo', 'KBANK.BK',  'SET', 'EQUITY', 'THB', 10, true),
    ('TH_EQ_XBKK_PTT',    'yahoo', 'PTT.BK',    'SET', 'EQUITY', 'THB', 10, true),
    ('TH_EQ_XBKK_CPALL',  'yahoo', 'CPALL.BK',  'SET', 'EQUITY', 'THB', 10, true),
    ('TH_EQ_XBKK_SCC',    'yahoo', 'SCC.BK',    'SET', 'EQUITY', 'THB', 10, true),
    ('TH_EQ_XBKK_BBL',    'yahoo', 'BBL.BK',    'SET', 'EQUITY', 'THB', 10, true),
    ('TH_EQ_XBKK_AOT',    'yahoo', 'AOT.BK',    'SET', 'EQUITY', 'THB', 10, true),
    ('TH_EQ_XBKK_ADVANC', 'yahoo', 'ADVANC.BK', 'SET', 'EQUITY', 'THB', 10, true),
    ('TH_EQ_XBKK_DELTA',  'yahoo', 'DELTA.BK',  'SET', 'EQUITY', 'THB', 10, true),

    -- FX (Yahoo uses the =X suffix)
    ('FX_USDTHB', 'yahoo', 'USDTHB=X', 'FOREX', 'FX', 'USD', 10, true),
    ('FX_EURUSD', 'yahoo', 'EURUSD=X', 'FOREX', 'FX', 'EUR', 10, true),
    ('FX_USDJPY', 'yahoo', 'USDJPY=X', 'FOREX', 'FX', 'USD', 10, true),
    ('FX_GBPUSD', 'yahoo', 'GBPUSD=X', 'FOREX', 'FX', 'GBP', 10, true)
  ) AS m(
    ims_symbol, provider_code, provider_symbol, provider_exchange,
    provider_asset_type, provider_currency, priority, is_primary
  ) ON sm.ims_symbol = m.ims_symbol
ON CONFLICT (provider_code, provider_symbol) DO UPDATE SET
    security_id         = EXCLUDED.security_id,
    provider_exchange   = EXCLUDED.provider_exchange,
    provider_asset_type = EXCLUDED.provider_asset_type,
    provider_currency   = EXCLUDED.provider_currency,
    priority            = EXCLUDED.priority,
    is_primary          = EXCLUDED.is_primary,
    mapping_status      = 'ACTIVE',
    updated_at          = NOW();

COMMIT;
