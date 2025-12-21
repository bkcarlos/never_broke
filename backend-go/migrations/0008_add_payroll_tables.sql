-- 0008_add_payroll_tables.sql
-- 薪资计算器相关表

-- 1. 地区表
CREATE TABLE IF NOT EXISTS payroll_regions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. 地区规则表（按年度版本）
CREATE TABLE IF NOT EXISTS payroll_region_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_code VARCHAR(50) NOT NULL,
    year INT NOT NULL,
    ss_base_min NUMERIC(15,2) NOT NULL DEFAULT 0,
    ss_base_max NUMERIC(15,2) NOT NULL DEFAULT 0,
    hf_base_min NUMERIC(15,2) NOT NULL DEFAULT 0,
    hf_base_max NUMERIC(15,2) NOT NULL DEFAULT 0,
    ss_person_rate_sum NUMERIC(10,6) NOT NULL DEFAULT 0,
    ss_company_rate_sum NUMERIC(10,6) NOT NULL DEFAULT 0,
    hf_rate_min NUMERIC(10,6) NOT NULL DEFAULT 0.05,
    hf_rate_max NUMERIC(10,6) NOT NULL DEFAULT 0.12,
    default_hf_rate_person NUMERIC(10,6) NOT NULL DEFAULT 0.12,
    default_hf_rate_company NUMERIC(10,6) NOT NULL DEFAULT 0.12,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(region_code, year)
);

CREATE INDEX IF NOT EXISTS idx_payroll_region_rules_region_code ON payroll_region_rules(region_code);
CREATE INDEX IF NOT EXISTS idx_payroll_region_rules_year ON payroll_region_rules(year);

-- 3. 综合所得税率表（中国）
CREATE TABLE IF NOT EXISTS tax_brackets_cn (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year INT NOT NULL,
    lower_bound NUMERIC(15,2) NOT NULL,
    upper_bound NUMERIC(15,2), -- NULL 表示无上限
    rate NUMERIC(10,6) NOT NULL,
    quick_deduction NUMERIC(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tax_brackets_cn_year ON tax_brackets_cn(year);

-- 4. 插入默认地区数据
INSERT INTO payroll_regions (code, name) VALUES 
    ('CN-DEFAULT', '全国通用'),
    ('CN-BJ', '北京'),
    ('CN-SH', '上海'),
    ('CN-GD-SZ', '深圳'),
    ('CN-GD-GZ', '广州'),
    ('CN-ZJ-HZ', '杭州')
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, updated_at = NOW();

-- 5. 插入2025年地区规则（CN-DEFAULT 通用）
INSERT INTO payroll_region_rules (
    region_code, year, 
    ss_base_min, ss_base_max, 
    hf_base_min, hf_base_max,
    ss_person_rate_sum, ss_company_rate_sum,
    hf_rate_min, hf_rate_max,
    default_hf_rate_person, default_hf_rate_company
) VALUES 
    -- 全国通用（保守估计）
    ('CN-DEFAULT', 2025, 4000, 30000, 2500, 30000, 0.105, 0.295, 0.05, 0.12, 0.12, 0.12),
    -- 北京
    ('CN-BJ', 2025, 6326, 35283, 2500, 35283, 0.105, 0.280, 0.05, 0.12, 0.12, 0.12),
    -- 上海
    ('CN-SH', 2025, 7310, 36549, 2590, 36549, 0.105, 0.280, 0.05, 0.12, 0.07, 0.07),
    -- 深圳
    ('CN-GD-SZ', 2025, 2360, 26421, 2360, 41190, 0.105, 0.265, 0.05, 0.12, 0.05, 0.05),
    -- 广州
    ('CN-GD-GZ', 2025, 5284, 26421, 2300, 39377, 0.105, 0.270, 0.05, 0.12, 0.05, 0.12),
    -- 杭州
    ('CN-ZJ-HZ', 2025, 4812, 24060, 2490, 39377, 0.105, 0.290, 0.05, 0.12, 0.12, 0.12)
ON CONFLICT (region_code, year) DO UPDATE SET 
    ss_base_min = EXCLUDED.ss_base_min,
    ss_base_max = EXCLUDED.ss_base_max,
    hf_base_min = EXCLUDED.hf_base_min,
    hf_base_max = EXCLUDED.hf_base_max,
    ss_person_rate_sum = EXCLUDED.ss_person_rate_sum,
    ss_company_rate_sum = EXCLUDED.ss_company_rate_sum,
    hf_rate_min = EXCLUDED.hf_rate_min,
    hf_rate_max = EXCLUDED.hf_rate_max,
    default_hf_rate_person = EXCLUDED.default_hf_rate_person,
    default_hf_rate_company = EXCLUDED.default_hf_rate_company,
    updated_at = NOW();

-- 6. 插入2025年综合所得税率表（7档）
INSERT INTO tax_brackets_cn (year, lower_bound, upper_bound, rate, quick_deduction) VALUES
    (2025, 0, 36000, 0.03, 0),
    (2025, 36000, 144000, 0.10, 2520),
    (2025, 144000, 300000, 0.20, 16920),
    (2025, 300000, 420000, 0.25, 31920),
    (2025, 420000, 660000, 0.30, 52920),
    (2025, 660000, 960000, 0.35, 85920),
    (2025, 960000, NULL, 0.45, 181920)
ON CONFLICT DO NOTHING;

-- 备注：
-- ss_person_rate_sum = 养老8% + 医疗2% + 失业0.5% = 10.5%
-- ss_company_rate_sum 各地不同，一般在 26%~30%
-- 公积金比例 5%~12%，各城市默认比例不同
