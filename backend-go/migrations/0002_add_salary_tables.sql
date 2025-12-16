-- NeverBroke 薪资功能数据库扩展
-- 新增城市社保比例表和扩展周期收入计划表

-- 城市社保比例表
CREATE TABLE IF NOT EXISTS city_social_rates (
  id TEXT PRIMARY KEY,
  city_code TEXT NOT NULL,
  city_name TEXT NOT NULL,
  effective_date DATE NOT NULL,
  
  -- 养老保险(个人缴纳比例)
  pension_rate DOUBLE PRECISION NOT NULL,
  -- 医疗保险(个人缴纳比例)
  medical_rate DOUBLE PRECISION NOT NULL,
  -- 失业保险(个人缴纳比例)
  unemployment_rate DOUBLE PRECISION NOT NULL,
  -- 工伤保险(个人通常不缴纳,但保留字段)
  injury_rate DOUBLE PRECISION DEFAULT 0,
  -- 生育保险(个人通常不缴纳,但保留字段)
  maternity_rate DOUBLE PRECISION DEFAULT 0,
  -- 住房公积金(个人缴纳比例)
  housing_fund_rate DOUBLE PRECISION NOT NULL,
  
  -- 总比例(自动计算: pension + medical + unemployment + housing_fund)
  total_rate DOUBLE PRECISION NOT NULL,
  
  -- 是否当前生效
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  
  -- 备注(政策说明)
  remark TEXT,
  
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_city_social_rates_city_active 
  ON city_social_rates(city_code, is_active);
CREATE INDEX IF NOT EXISTS idx_city_social_rates_effective_date 
  ON city_social_rates(city_code, effective_date DESC);

-- 新增薪资专用字段到 recurring_income_plans 表
ALTER TABLE recurring_income_plans 
  ADD COLUMN IF NOT EXISTS salary_pre_tax_monthly DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS salary_social_rate DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS salary_extra_deductions DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS salary_pay_day INT,
  ADD COLUMN IF NOT EXISTS salary_start_date DATE;
