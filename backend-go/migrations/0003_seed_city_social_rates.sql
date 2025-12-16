-- NeverBroke 城市社保比例初始数据
-- 基于2024年各城市公开的社保政策

-- 北京市社保比例(2024年标准)
INSERT INTO city_social_rates (id, city_code, city_name, effective_date, 
  pension_rate, medical_rate, unemployment_rate, injury_rate, maternity_rate, 
  housing_fund_rate, total_rate, is_active, remark)
VALUES (
  'beijing-2024', 'beijing', '北京', '2024-01-01',
  0.08, 0.02, 0.002, 0, 0,
  0.12, 0.202, true,
  '养老8% + 医疗2% + 失业0.2% + 公积金12%'
);

-- 上海市社保比例(2024年标准)
INSERT INTO city_social_rates (id, city_code, city_name, effective_date,
  pension_rate, medical_rate, unemployment_rate, injury_rate, maternity_rate,
  housing_fund_rate, total_rate, is_active, remark)
VALUES (
  'shanghai-2024', 'shanghai', '上海', '2024-01-01',
  0.08, 0.02, 0.005, 0, 0,
  0.07, 0.175, true,
  '养老8% + 医疗2% + 失业0.5% + 公积金7%'
);

-- 广州市社保比例(2024年标准)
INSERT INTO city_social_rates (id, city_code, city_name, effective_date,
  pension_rate, medical_rate, unemployment_rate, injury_rate, maternity_rate,
  housing_fund_rate, total_rate, is_active, remark)
VALUES (
  'guangzhou-2024', 'guangzhou', '广州', '2024-01-01',
  0.08, 0.02, 0.002, 0, 0,
  0.05, 0.152, true,
  '养老8% + 医疗2% + 失业0.2% + 公积金5%'
);

-- 深圳市社保比例(2024年标准)
INSERT INTO city_social_rates (id, city_code, city_name, effective_date,
  pension_rate, medical_rate, unemployment_rate, injury_rate, maternity_rate,
  housing_fund_rate, total_rate, is_active, remark)
VALUES (
  'shenzhen-2024', 'shenzhen', '深圳', '2024-01-01',
  0.08, 0.02, 0.005, 0, 0,
  0.05, 0.155, true,
  '养老8% + 医疗2% + 失业0.5% + 公积金5%'
);

-- 默认比例(其他城市)
INSERT INTO city_social_rates (id, city_code, city_name, effective_date,
  pension_rate, medical_rate, unemployment_rate, injury_rate, maternity_rate,
  housing_fund_rate, total_rate, is_active, remark)
VALUES (
  'default-2024', 'default', '其他城市', '2024-01-01',
  0.08, 0.02, 0.002, 0, 0,
  0.10, 0.182, true,
  '养老8% + 医疗2% + 失业0.2% + 公积金10%(参考值)'
);
