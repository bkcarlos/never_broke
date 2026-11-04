-- Existing subscriptions and new fixed expenses share versioned workspace settings.
ALTER TABLE IF EXISTS workspace_settings ADD COLUMN IF NOT EXISTS account_id text;
ALTER TABLE IF EXISTS workspace_settings ADD COLUMN IF NOT EXISTS end_date text;
