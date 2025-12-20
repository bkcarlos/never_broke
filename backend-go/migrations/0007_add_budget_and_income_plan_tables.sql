-- NeverBroke 多用户支持扩展
-- 新增 budgets 和 income_plans 表

-- 月度预算表
CREATE TABLE IF NOT EXISTS budgets (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  month TEXT NOT NULL,  -- 'YYYY-MM' 格式
  total_budget DOUBLE PRECISION NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  
  UNIQUE(user_id, month)
);

CREATE INDEX IF NOT EXISTS idx_budgets_user_month ON budgets(user_id, month);

-- 收入计划表
CREATE TABLE IF NOT EXISTS income_plans (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  name TEXT NOT NULL,
  amount DOUBLE PRECISION NOT NULL,
  frequency TEXT NOT NULL,  -- monthly, weekly, biweekly, yearly
  account_id TEXT NOT NULL,
  first_pay_date DATE NOT NULL,
  end_date DATE,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_income_plans_user ON income_plans(user_id);
CREATE INDEX IF NOT EXISTS idx_income_plans_user_active ON income_plans(user_id, active);
