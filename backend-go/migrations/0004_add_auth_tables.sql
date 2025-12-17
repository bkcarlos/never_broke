-- 0004_add_auth_tables.sql
-- 用户认证系统表结构
-- 安全要求：
-- 1. email 必须唯一且小写存储（防止大小写绕过）
-- 2. password_hash 使用 argon2id 存储，永不存储明文
-- 3. sessions 表支持服务端会话管理
-- 4. nonces 表防止重放攻击（每个 nonce 只能使用一次）

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    nickname VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引以加速邮箱查询（登录时）
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- 会话表（服务端 session 管理）
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,  -- session_id (UUID v4)
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引以加速会话查询和清理过期会话
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Nonce 表（防重放攻击）
-- 每个 nonce 只能使用一次，过期后自动清理
CREATE TABLE IF NOT EXISTS nonces (
    nonce TEXT PRIMARY KEY,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引以加速过期 nonce 清理
CREATE INDEX IF NOT EXISTS idx_nonces_expires_at ON nonces(expires_at);

-- 触发器：自动更新 users 表的 updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 注释说明
COMMENT ON TABLE users IS '用户表：存储用户基本信息，密码使用 argon2id 哈希';
COMMENT ON COLUMN users.email IS '邮箱地址（唯一，小写存储）';
COMMENT ON COLUMN users.password_hash IS 'argon2id 哈希后的密码（永不存储明文）';
COMMENT ON COLUMN users.nickname IS '用户昵称';

COMMENT ON TABLE sessions IS '会话表：服务端 session 管理，配合 HttpOnly Cookie';
COMMENT ON COLUMN sessions.id IS 'Session ID (UUID v4)';
COMMENT ON COLUMN sessions.expires_at IS '过期时间（建议 24 小时）';

COMMENT ON TABLE nonces IS 'Nonce 表：防重放攻击，每个 nonce 只能使用一次';
COMMENT ON COLUMN nonces.nonce IS '一次性随机数（32 字节 base64）';
COMMENT ON COLUMN nonces.expires_at IS '过期时间（建议 60 秒）';
COMMENT ON COLUMN nonces.used IS '是否已使用';
