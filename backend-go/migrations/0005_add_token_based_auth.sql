-- 0005_add_token_based_auth.sql
-- Token 认证系统（Access Token + Refresh Token）
-- 支持 Web + App + 小程序多端登录

-- 删除旧的 session 和 nonce 表（改用 token 方案）
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS nonces CASCADE;

-- refresh_tokens 表（可撤销 + 轮换 + 设备维度）
-- 安全要求：
-- 1. token_hash 存储哈希，不存明文
-- 2. 支持 refresh rotation（每次刷新生成新 token，旧 token 作废）
-- 3. 支持设备维度管理（可单独撤销某设备）
-- 4. 支持重放检测（replaced_by 链路）
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id TEXT PRIMARY KEY,                    -- UUID v4 (作为 jti)
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL,                -- 客户端设备 ID（UUID）
    token_hash TEXT NOT NULL,               -- refresh token 的 SHA-256 哈希
    issued_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,                   -- 撤销时间（可空）
    replaced_by TEXT,                       -- 被哪个新 token 替换（轮换链路）
    last_used_at TIMESTAMP,                 -- 最后使用时间
    user_agent TEXT,                        -- 用户代理（可选）
    ip_address TEXT                         -- IP 地址（可选）
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_device_id ON refresh_tokens(device_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- oauth_identities 表（第三方登录）
-- 支持微信、Google、Apple、GitHub 等 OAuth/OIDC 提供商
CREATE TABLE IF NOT EXISTS oauth_identities (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,          -- 提供商：wechat, google, apple, github, oidc
    subject TEXT NOT NULL,                  -- 唯一标识：OIDC sub / 微信 unionid / openid
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, subject)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_oauth_identities_user_id ON oauth_identities(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_identities_provider_subject ON oauth_identities(provider, subject);

-- oauth_states 表（OAuth 流程 state 管理）
-- 用于防止 CSRF 攻击
CREATE TABLE IF NOT EXISTS oauth_states (
    state TEXT PRIMARY KEY,                 -- 随机 state（32 字节）
    provider VARCHAR(50) NOT NULL,
    redirect_uri TEXT,                      -- 回调后的前端地址
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL           -- 5 分钟过期
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_oauth_states_expires_at ON oauth_states(expires_at);

-- 触发器：自动更新 oauth_identities 的 updated_at
CREATE OR REPLACE FUNCTION update_oauth_identities_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_oauth_identities_updated_at 
BEFORE UPDATE ON oauth_identities
FOR EACH ROW EXECUTE FUNCTION update_oauth_identities_updated_at();

-- 注释说明
COMMENT ON TABLE refresh_tokens IS 'Refresh Token 表：支持多端登录、设备管理、token 轮换';
COMMENT ON COLUMN refresh_tokens.id IS 'Token ID (UUID v4，作为 JWT jti)';
COMMENT ON COLUMN refresh_tokens.device_id IS '设备 ID（客户端生成并持久化）';
COMMENT ON COLUMN refresh_tokens.token_hash IS 'Refresh Token 的 SHA-256 哈希（禁止明文存储）';
COMMENT ON COLUMN refresh_tokens.replaced_by IS '被哪个新 token 替换（用于轮换链路和重放检测）';

COMMENT ON TABLE oauth_identities IS '第三方登录身份表：支持微信、OIDC 等';
COMMENT ON COLUMN oauth_identities.provider IS '提供商：wechat, google, apple, github, oidc';
COMMENT ON COLUMN oauth_identities.subject IS '唯一标识：OIDC sub / 微信 unionid';

COMMENT ON TABLE oauth_states IS 'OAuth State 表：防止 CSRF 攻击';
COMMENT ON COLUMN oauth_states.state IS '随机 state（32 字节，5 分钟过期）';
