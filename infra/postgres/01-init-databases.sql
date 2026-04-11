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
    id VARCHAR(255) PRIMARY KEY,
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
-- 数据库初始化完成
-- ============================================================================
