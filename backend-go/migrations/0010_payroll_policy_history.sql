-- Reviewed policy periods are seeded idempotently by payroll.SeedPolicyHistory.
ALTER TABLE recurring_income_plans ADD COLUMN IF NOT EXISTS salary_versions TEXT;
CREATE TABLE IF NOT EXISTS policy_histories (
 id VARCHAR(80) PRIMARY KEY,
 region_code TEXT NOT NULL,
 effective_from VARCHAR(10) NOT NULL,
 effective_through VARCHAR(10) NOT NULL,
 payload TEXT NOT NULL,
 verified_at VARCHAR(10) NOT NULL,
 CHECK (effective_from <= effective_through)
);
CREATE INDEX IF NOT EXISTS idx_policy_period ON policy_histories(region_code, effective_from);

CREATE TABLE IF NOT EXISTS source_documents (
 id VARCHAR(64) PRIMARY KEY,
 url TEXT,
 name TEXT,
 sha256 VARCHAR(64),
 fetched_at TIMESTAMPTZ,
 content BYTEA
);
