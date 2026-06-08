-- ChillCat 数据库初始化脚本
-- 阶段一：单体服务

-- 创建数据库（如果不存在）
-- CREATE DATABASE chillcat;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(64)  NOT NULL UNIQUE,
    email       VARCHAR(255) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL,
    nickname    VARCHAR(64)  NOT NULL DEFAULT '',
    avatar      VARCHAR(512) NOT NULL DEFAULT '',
    status      SMALLINT     NOT NULL DEFAULT 1, -- 1:正常 0:禁用
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

COMMENT ON TABLE users IS '用户表';
COMMENT ON COLUMN users.status IS '1:正常 0:禁用';

-- 会员信息表
CREATE TABLE IF NOT EXISTS member_info (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    member_type VARCHAR(32)  NOT NULL DEFAULT '', -- monthly/quarterly/yearly/permanent
    status      VARCHAR(32)  NOT NULL DEFAULT 'none', -- active/expired/cancelled/none
    start_date  TIMESTAMP,
    end_date    TIMESTAMP,
    auto_renew  BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_member_info_user ON member_info(user_id);

COMMENT ON TABLE member_info IS '会员信息表';
COMMENT ON COLUMN member_info.member_type IS '会员类型: monthly/quarterly/yearly/permanent';
COMMENT ON COLUMN member_info.status IS '状态: active/expired/cancelled/none';

-- 会员订单表
CREATE TABLE IF NOT EXISTS member_orders (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_no        VARCHAR(64)  NOT NULL UNIQUE,
    member_type     VARCHAR(32)  NOT NULL,
    order_status    VARCHAR(32)  NOT NULL DEFAULT 'pending', -- pending/success/failed/refunded
    amount          DECIMAL(10,2) NOT NULL DEFAULT 0,
    payment_method  VARCHAR(32)  NOT NULL DEFAULT '',
    start_date      TIMESTAMP,
    end_date        TIMESTAMP,
    paid_at         TIMESTAMP,
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_user ON member_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_no ON member_orders(order_no);

COMMENT ON TABLE member_orders IS '会员订单表';
COMMENT ON COLUMN member_orders.order_status IS '订单状态: pending/success/failed/refunded';

-- 插入测试数据（可选）
-- INSERT INTO users (username, email, password, nickname) VALUES
--     ('test', 'test@chillcat.app', '$2a$10$...', '测试用户');
