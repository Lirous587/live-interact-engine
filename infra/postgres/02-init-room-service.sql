-- ==================== Room Service 数据库初始化 ====================

-- 创建用户服务数据库
CREATE DATABASE room_service;

-- 为 room_service 数据库创建表
\connect room_service

-- 创建 rooms 表
-- 注：owner_id 不设为外键，因为 rooms 和 users 在不同的微服务数据库中
-- 用户现有性验证由应用层在 room-service 中通过 user-service RPC 调用完成
CREATE TABLE IF NOT EXISTS rooms (
    room_id VARCHAR(100) PRIMARY KEY,
    owner_id VARCHAR(100) NOT NULL,
    title VARCHAR(30) NOT NULL,
    description TEXT,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    is_active BOOLEAN DEFAULT true
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_rooms_owner_id ON rooms(owner_id);
CREATE INDEX IF NOT EXISTS idx_rooms_is_active ON rooms(is_active);

-- 创建 user_room_roles 表
-- 注：user_id 不设为外键，因为 users 表在 user_service 库中
-- user_id 有效性验证由应用层完成
CREATE TABLE IF NOT EXISTS user_room_roles (
    user_id VARCHAR(100) NOT NULL,
    room_id VARCHAR(100) NOT NULL,
    role_name VARCHAR(50) NOT NULL,
    permissions JSONB DEFAULT '[]',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    PRIMARY KEY (user_id, room_id),
    FOREIGN KEY (room_id) REFERENCES rooms(room_id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_user_room_roles_room_id ON user_room_roles(room_id);
CREATE INDEX IF NOT EXISTS idx_user_room_roles_user_id ON user_room_roles(user_id);
