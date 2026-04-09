-- ============================================================================
-- 数据库初始化脚本
-- ============================================================================

-- 创建用户服务数据库
CREATE DATABASE user_service;

-- 为 user_service 数据库创建表
\connect user_service

-- ============================================================================
-- users 表: 用户信息存储
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    is_active BOOLEAN DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- ============================================================================
-- user_room_roles 表: 用户房间角色及权限
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_room_roles (
    user_id VARCHAR(255) NOT NULL,
    room_id VARCHAR(255) NOT NULL,
    permissions JSONB DEFAULT '[]'::jsonb,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    PRIMARY KEY (user_id, room_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_room_roles_room_id ON user_room_roles(room_id);
CREATE INDEX IF NOT EXISTS idx_user_room_roles_created_at ON user_room_roles(created_at);

-- ============================================================================
-- 数据库初始化完成
-- ============================================================================
