-- 初始化礼物数据
INSERT INTO gifts (id, name, description, icon_url, cache_key, price, vip_only, status, created_at, updated_at) VALUES
-- 普通礼物
('550e8400-e29b-41d4-a716-446655440001'::uuid, '点赞', '最简单的表达方式', 'https://example.com/thumbs-up.png', 'gift:thumbs_up', 1, false, 'online', now(), now()),
('550e8400-e29b-41d4-a716-446655440002'::uuid, '玫瑰花', '传递爱意的礼物', 'https://example.com/rose.png', 'gift:rose', 10, false, 'online', now(), now()),
('550e8400-e29b-41d4-a716-446655440003'::uuid, '钻戒', '贵重的礼物', 'https://example.com/diamond_ring.png', 'gift:diamond_ring', 100, false, 'online', now(), now()),
('550e8400-e29b-41d4-a716-446655440004'::uuid, '劳斯莱斯', '极致奢华', 'https://example.com/rolls_royce.png', 'gift:rolls_royce', 1000, false, 'online', now(), now()),
('550e8400-e29b-41d4-a716-446655440005'::uuid, '城堡', '送你一座城堡', 'https://example.com/castle.png', 'gift:castle', 5000, false, 'online', now(), now()),

-- VIP专属礼物
('550e8400-e29b-41d4-a716-446655440010'::uuid, 'VIP勋章', 'VIP用户专属', 'https://example.com/vip_badge.png', 'gift:vip_badge', 50, true, 'online', now(), now()),
('550e8400-e29b-41d4-a716-446655440011'::uuid, '皇冠', 'VIP身份象征', 'https://example.com/crown.png', 'gift:crown', 200, true, 'online', now(), now()),

-- 限时礼物
('550e8400-e29b-41d4-a716-446655440020'::uuid, '新春红包', '新年限定礼物', 'https://example.com/red_envelope.png', 'gift:red_envelope', 88, false, 'limited_time', now(), now()),
('550e8400-e29b-41d4-a716-446655440021'::uuid, '圣诞礼物', '圣诞节限定', 'https://example.com/christmas_gift.png', 'gift:christmas_gift', 66, false, 'limited_time', now(), now()),

-- 下线礼物（作为示例）
('550e8400-e29b-41d4-a716-446655440030'::uuid, '过期礼物', '已下路的礼物', 'https://example.com/expired.png', 'gift:expired', 1, false, 'offline', now(), now());

-- 允许再次运行脚本时不出错（或自动更新），可选择添加以下注释代码：
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  icon_url = EXCLUDED.icon_url,
  cache_key = EXCLUDED.cache_key,
  price = EXCLUDED.price,
  vip_only = EXCLUDED.vip_only,
  status = EXCLUDED.status,
  updated_at = now();
