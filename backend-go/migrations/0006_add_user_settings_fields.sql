-- 0006_add_user_settings_fields.sql
-- 为 users 表添加用户设置字段（替代内存存储）

-- 添加 locale 字段（语言/区域设置）
ALTER TABLE users ADD COLUMN IF NOT EXISTS locale VARCHAR(10) NOT NULL DEFAULT 'zh-CN';

-- 添加 show_amount 字段（是否显示金额）
ALTER TABLE users ADD COLUMN IF NOT EXISTS show_amount BOOLEAN NOT NULL DEFAULT TRUE;

-- 注释说明
COMMENT ON COLUMN users.locale IS '用户语言/区域设置（如 zh-CN, en-US）';
COMMENT ON COLUMN users.show_amount IS '是否显示金额（隐私设置）';
