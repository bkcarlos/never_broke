-- NeverBroke 数据库初始化脚本
-- 单一数据库 + user_id 隔离

-- 用户表（简单结构，主要用于标识）
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 账户表
CREATE TABLE IF NOT EXISTS accounts (
  id TEXT PRIMARY KEY,
  user_id UUID NOT NULL,
  type TEXT NOT NULL,
  institution TEXT,
  card_last4 TEXT,
  currency TEXT NOT NULL,
  balance DOUBLE PRECISION NOT NULL DEFAULT 0,
  credit_limit DOUBLE PRECISION,
  temp_limit DOUBLE PRECISION,
  archived BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_accounts_user ON accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_accounts_user_archived ON accounts(user_id, archived);

-- 交易表
CREATE TABLE IF NOT EXISTS transactions (
  id TEXT PRIMARY KEY,
  user_id UUID NOT NULL,
  type TEXT NOT NULL,
  amount DOUBLE PRECISION NOT NULL,
  category_key TEXT,
  remark TEXT,
  occurred_at TIMESTAMPTZ NOT NULL,
  account_id TEXT,
  from_account_id TEXT,
  to_account_id TEXT,
  installment_plan_id TEXT,
  installment_period INT,
  installment_total INT
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_time 
  ON transactions(user_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_user_account 
  ON transactions(user_id, account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_installment_plan 
  ON transactions(installment_plan_id) WHERE installment_plan_id IS NOT NULL;

-- 分期计划表
CREATE TABLE IF NOT EXISTS installment_plans (
  id TEXT PRIMARY KEY,
  user_id UUID NOT NULL,
  account_id TEXT NOT NULL,
  principal DOUBLE PRECISION NOT NULL,
  total_interest DOUBLE PRECISION NOT NULL,
  total_amount DOUBLE PRECISION NOT NULL,
  term_months INT NOT NULL,
  monthly_rate DOUBLE PRECISION NOT NULL,
  repay_method TEXT NOT NULL,
  currency TEXT NOT NULL,
  first_due_date DATE NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_installment_user_account 
  ON installment_plans(user_id, account_id);
CREATE INDEX IF NOT EXISTS idx_installment_user_status 
  ON installment_plans(user_id, status);

-- 周期收入计划表
CREATE TABLE IF NOT EXISTS recurring_income_plans (
  id TEXT PRIMARY KEY,
  user_id UUID NOT NULL,
  type TEXT NOT NULL,
  label TEXT NOT NULL,
  account_id TEXT NOT NULL,
  amount DOUBLE PRECISION NOT NULL,
  currency TEXT NOT NULL,
  frequency TEXT NOT NULL,
  day_of_month INT,
  weekday INT,
  next_due_date DATE NOT NULL,
  last_fired_date DATE,
  city_code TEXT,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_recurring_income_user 
  ON recurring_income_plans(user_id);
CREATE INDEX IF NOT EXISTS idx_recurring_income_user_active 
  ON recurring_income_plans(user_id, active);
CREATE INDEX IF NOT EXISTS idx_recurring_income_next_due_date 
  ON recurring_income_plans(next_due_date) WHERE active = TRUE;




