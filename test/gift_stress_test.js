/**
 * 礼物打赏极限压测脚本（Stress Test）
 *
 * 目标：找到 gift/send 接口的最大 RPS 边界
 *
 * ─────────────────────────────────────────────────────────────────
 * 名词解释
 * ─────────────────────────────────────────────────────────────────
 * RPS (Requests Per Second)
 *   每秒请求数，衡量吞吐量的核心指标。
 *   本脚本用 ramping-arrival-rate 控制 RPS，而不是控制并发数。
 *
 * VU (Virtual User，虚拟用户)
 *   k6 内部的并发执行单元，每个 VU 独立发请求、等响应、再发请求。
 *   在 arrival-rate 模式下，VU 是为了维持目标 RPS 而自动申请的"工人"：
 *     所需 VU ≈ 目标 RPS × 平均响应时间(秒)
 *     例：8000 RPS × 0.228s ≈ 1824 VU
 *   若 VU 不够，k6 会打印 "Insufficient VUs" 并产生 dropped_iterations，
 *   此时是压测工具本身的瓶颈，而非服务瓶颈。
 *
 * p(95) 延迟 (第 95 百分位延迟)
 *   将所有请求的响应时间从小到大排序后，第 95% 位置的值。
 *   含义：100 个请求里最慢的 5 个不算，剩下 95 个中最慢的那个。
 *   为什么不看平均值？因为平均值会被大量快请求拉低，掩盖真实的慢请求。
 *   p(95) 能更真实地反映"大多数用户的体验上限"。
 *   类似指标：p(90) 更宽松，p(99) 更严格（关注极端慢请求）。
 *
 * dropped_iterations
 *   k6 按计划应该发出但实际没发出的请求数。
 *   原因通常是：VU 数量上限不足（见上），或服务响应极慢导致 VU 全部卡住。
 *   dropped_iterations 多 ≠ 服务报错，只代表压力没打到位。
 *
 * gift_error_rate
 *   业务失败率（HTTP 非 200）。
 *   飙升原因：余额耗尽 / 幂等冲突 / 服务过载返回 5xx。
 *   阈值设为 < 5%，超过则本次压测判定为不达标。
 *
 * thresholds（阈值）
 *   压测的通过/失败判定条件。全部满足 → stress ✓，任一不满足 → stress ✗。
 * ─────────────────────────────────────────────────────────────────
 *
 * 测试策略：ramping-arrival-rate —— 6 档跳升，总计约 100s（不含 setup）
 *   1000 → 2000 → 4000 → 6000 → 8000 → 0
 *
 * 运行方式：
 *   k6 run -e BASE_URL=http://localhost:8080 test/gift_stress_test.js
 */

import http from 'k6/http';
import { check } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// ==================== 环境变量 ====================
const BASE_URL  = __ENV.BASE_URL || 'http://localhost:8080';
const NUM_USERS       = 200;
const BATCH_SIZE      = 50;
// 100s × 8000 RPS / 199 用户 ≈ 4020 次/人，充值 20000 留足 5x 余量
const RECHARGE_AMOUNT = 20000;

// ==================== 自定义指标 ====================
const giftSentTotal = new Counter('gift_sent_total');
const giftErrorRate = new Rate('gift_error_rate');
const giftLatency   = new Trend('gift_req_duration_ms', true);

// ==================== k6 到达速率爬坡配置 ====================
export const options = {
  setupTimeout: '300s',
  scenarios: {
    stress: {
      executor: 'ramping-arrival-rate',
      // avg 600ms × 8000 RPS ≈ 4800 VU，6000 留足 1.25x 余量
      preAllocatedVUs: 2000,
      maxVUs: 6000,
      startRate: 500,
      timeUnit: '1s',
      stages: [
        { duration: '10s', target: 1000 }, // 快速热身到已知稳定区
        { duration: '20s', target: 2000 }, // 超出已知上限，观察拐点
        { duration: '20s', target: 4000 }, // 高压
        { duration: '20s', target: 6000 }, // 极限冲刺
        { duration: '20s', target: 8000 }, // 绝对上限探测
        { duration: '10s', target: 0    }, // 降载
      ],
    },
  },
  thresholds: {
    gift_error_rate:                       ['rate < 0.05'],
    'http_req_duration{scenario:stress}':  ['p(95)<2000'],
  },
};

// ==================== 辅助：分批执行 http.batch ====================
function batchChunked(requests, size) {
  const results = [];
  for (let i = 0; i < requests.length; i += size) {
    const chunk = requests.slice(i, i + size);
    results.push(...http.batch(chunk));
  }
  return results;
}

// ==================== setup：只执行一次 ====================
export function setup() {
  const ts       = Date.now();
  const password = 'Password123!';
  const headers  = { 'Content-Type': 'application/json' };
  const accounts = Array.from({ length: NUM_USERS }, (_, i) => ({
    email:    `lt_gift_st_${ts}_${i}@test.com`,
    username: `lt_gift_st_${ts}_${i}`,
    password,
  }));

  // 分批并发注册（每批 BATCH_SIZE，避免瞬间打满注册接口）
  console.log(`[setup] 分批注册 ${NUM_USERS} 个用户（每批 ${BATCH_SIZE}）...`);
  batchChunked(
    accounts.map(u => ['POST', `${BASE_URL}/api/v1/user/register`,
      JSON.stringify({ email: u.email, username: u.username, password: u.password }), { headers }]),
    BATCH_SIZE,
  );

  // 分批并发登录
  const loginResps = batchChunked(
    accounts.map(u => ['POST', `${BASE_URL}/api/v1/user/login`,
      JSON.stringify({ email: u.email, password: u.password }), { headers }]),
    BATCH_SIZE,
  );

  const sessions = loginResps
    .filter(r => r.status === 200)
    .map(r => {
      const body = JSON.parse(r.body);
      return { token: body.data.access_token, userID: body.data.user_id };
    });

  if (sessions.length === 0) throw new Error('[setup] 登录失败，无可用 session');
  console.log(`[setup] ${sessions.length} 个 session 就绪`);

  const anchorSession = sessions[0];

  // 创建房间
  const roomRes = http.post(
    `${BASE_URL}/api/v1/room`,
    JSON.stringify({ title: `StressTest-Gift-${ts}` }),
    { headers: { ...headers, Authorization: `Bearer ${anchorSession.token}` } },
  );
  if (roomRes.status !== 200) throw new Error(`[setup] 创建房间失败: ${roomRes.status} ${roomRes.body}`);
  const roomID = JSON.parse(roomRes.body).data.room_id;

  // 获取礼物（需要 auth，gift/list 若是公开接口则无需）
  const giftListRes = http.get(`${BASE_URL}/api/v1/gift/list`,
    { headers: { Authorization: `Bearer ${anchorSession.token}` } });
  if (giftListRes.status !== 200) throw new Error(`[setup] 礼物列表获取失败: ${giftListRes.status}`);
  const body  = JSON.parse(giftListRes.body);
  const gifts = body.data && body.data.gifts;
  if (!gifts || gifts.length === 0) {
    throw new Error(`[setup] 礼物列表为空，请先执行 seed 脚本初始化礼物数据\n响应: ${giftListRes.body}`);
  }
  // 优先选最便宜的非 VIP 礼物，减少余额消耗
  const gift = gifts
    .filter(g => !g.vip_only && g.status === 'available')
    .sort((a, b) => a.price - b.price)[0] || gifts[0];
  const giftID = gift.id;
  console.log(`[setup] 使用礼物: ${gift.name}（price=${gift.price}，id=${giftID}）`);

  // 分批并发充值（跳过 anchor，其发送礼物会被 ErrSelfGifting 拒绝）
  const rechargeResps = batchChunked(
    sessions.slice(1).map(s => ['POST', `${BASE_URL}/api/v1/wallet/recharge`,
      JSON.stringify({ amount: RECHARGE_AMOUNT }),
      { headers: { ...headers, Authorization: `Bearer ${s.token}` } }]),
    BATCH_SIZE,
  );
  const recharged = rechargeResps.filter(r => r.status === 200).length;
  console.log(`[setup] 充值完成: ${recharged}/${sessions.length - 1}，roomID=${roomID}，giftID=${giftID}`);

  return { sessions: sessions.slice(1), anchorID: anchorSession.userID, roomID, giftID };
}

// ==================== VU 主函数 ====================
export default function (data) {
  const { sessions, anchorID, roomID, giftID } = data;
  const session = sessions[(__VU - 1) % sessions.length];

  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/gift/send`,
    JSON.stringify({
      anchor_id: anchorID,
      room_id:   roomID,
      gift_id:   giftID,
      amount:    1,
    }),
    {
      headers: {
        'Content-Type': 'application/json',
        Authorization:  `Bearer ${session.token}`,
      },
      tags: { name: 'gift_send' },
    },
  );

  giftLatency.add(Date.now() - start);

  const ok = check(res, { '礼物发送成功 (200)': (r) => r.status === 200 });
  if (ok) {
    giftSentTotal.add(1);
    giftErrorRate.add(0);
  } else {
    giftErrorRate.add(1);
    const body = res.body || '';
    if (!body.includes('insufficient')) {
      console.warn(`[VU ${__VU}] 失败 status=${res.status} body=${body.substring(0, 120)}`);
    }
  }
}
