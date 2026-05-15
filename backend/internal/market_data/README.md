# Module: market_data

This module owns market-data ingestion and lookup support.

## Scope

- Alpha Vantage official provider adapter
- Yahoo Finance unofficial fallback adapter
- quote and daily price lookup API
- provider fallback to stale cached PostgreSQL snapshots
- optional Redis quote cache
- provider request logging
