-- Add chain-state indexes to accelerate module 10/11 lookups in replay and live scoring

CREATE INDEX IF NOT EXISTS idx_gk_chains_active_gk_started_desc
    ON geokrety_stats.gk_chains (gk_id, started_at DESC)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_gk_chains_active_last_active
    ON geokrety_stats.gk_chains (chain_last_active)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_gk_chain_members_chain_position
    ON geokrety_stats.gk_chain_members (chain_id, position);
