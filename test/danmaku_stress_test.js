/**
 * 弹幕 WebSocket 极限压测脚本（Stress Test）
 *
 * 目标：快速找到 WebSocket 并发连接数的崩溃边界
 *
 * 测试策略：4 档直接跳升，每档稳定 20s，总计约 90s（不含 setup）
 *   100 → 300 → 600 → 1000 → 0
 *
 * setup 优化：http.batch() 并发注册 + 登录，耗时从 ~40s 降至 ~5s
 *
 * 关键指标：
 *   - ws_connecting p(95)：连接建立延迟，飙升说明达到 goroutine/fd 上限
 *   - danmaku_received_total/s：接收吞吐，停止增长说明达到广播瓶颈
 *   - ws_connect_error_rate：错误率超过 5% 视为达到瓶颈
 *
 * 运行方式：
 *   k6 run -e BASE_URL=http://localhost:8080 test/load/danmaku_stress_test.js
 */

import ws from 'k6/ws';
import http from 'k6/http';
import { check } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// ==================== 环境变量 ====================
const BASE_URL    = __ENV.BASE_URL || 'http://localhost:8080';
const NUM_USERS   = 20; // batch 并发注册，20 个 token 足够 1000 VU 复用
const WS_BASE_URL = BASE_URL.replace(/^http/, 'ws');

// ==================== 自定义指标 ====================
const danmakuReceived  = new Counter('danmaku_received_total');
const danmakuSent      = new Counter('danmaku_sent_total');
const wsConnectErrors  = new Counter('ws_connect_errors_total');
const wsConnectSuccess = new Counter('ws_connect_success_total');
const wsConnectTime    = new Trend('ws_connect_time_ms', true);
const wsErrorRate      = new Rate('ws_connect_error_rate');

// ==================== k6 爬坡配置 ====================
export const options = {
  scenarios: {
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '15s', target: 100  }, // 热身 + 基准
        { duration: '20s', target: 300  }, // 中压
        { duration: '20s', target: 600  }, // 高压
        { duration: '20s', target: 1000 }, // 极限
        { duration: '15s', target: 0    }, // 降载
      ],
      gracefulRampDown: '10s',
    },
  },
  thresholds: {
    ws_connect_error_rate:                ['rate < 0.05'],
    'ws_connect_time_ms{scenario:stress}': ['p(95)<2000'],
  },
};

// ==================== setup：只执行一次 ====================
export function setup() {
  const ts = Date.now();
  const password = 'Password123!';
  const accounts = Array.from({ length: NUM_USERS }, (_, i) => ({
    email:    `lt_dm_st_${ts}_${i}@test.com`,
    username: `lt_dm_st_${ts}_${i}`,
    password,
  }));

  // 并发注册
  console.log(`[setup] 并发注册 ${NUM_USERS} 个用户...`);
  const headers = { 'Content-Type': 'application/json' };
  http.batch(accounts.map(u => ['POST', `${BASE_URL}/api/v1/user/register`,
    JSON.stringify({ email: u.email, username: u.username, password: u.password }), { headers }]));

  // 并发登录
  const loginResps = http.batch(accounts.map(u => ['POST', `${BASE_URL}/api/v1/user/login`,
    JSON.stringify({ email: u.email, password: u.password }), { headers }]));

  const sessions = loginResps
    .filter(r => r.status === 200)
    .map(r => {
      const body = JSON.parse(r.body);
      return { token: body.data.access_token, userID: body.data.user_id, username: body.data.username };
    });

  if (sessions.length === 0) throw new Error('[setup] 登录失败，终止压测');
  console.log(`[setup] 获取到 ${sessions.length} 个 session`);

  const roomRes = http.post(
    `${BASE_URL}/api/v1/room`,
    JSON.stringify({ title: `StressTest-DM-${ts}`, description: 'k6 极限压测' }),
    { headers: { ...headers, Authorization: `Bearer ${sessions[0].token}` } },
  );
  if (roomRes.status !== 200) throw new Error(`[setup] 创建房间失败: ${roomRes.status}`);

  const roomID = JSON.parse(roomRes.body).data.room_id;
  console.log(`[setup] roomID=${roomID}`);
  return { sessions, roomID };
}

// ==================== VU 主函数 ====================
export default function (data) {
  const { sessions, roomID } = data;
  const session = sessions[(__VU - 1) % sessions.length];

  const connectStart = Date.now();
  const res = ws.connect(
    `${WS_BASE_URL}/api/v1/danmaku/ws?room_id=${roomID}`,
    { headers: { Authorization: `Bearer ${session.token}` } },
    function (socket) {
      wsConnectTime.add(Date.now() - connectStart);

      let sendTimer = null;

      socket.on('open', function () {
        sendTimer = socket.setInterval(function () {
          socket.send(JSON.stringify({
            type:         'send',
            username:     session.username,
            content:      `stress VU=${__VU} t=${Date.now()}`,
            danmaku_type: 0,
          }));
          danmakuSent.add(1);
        }, 1000);
      });

      socket.on('message', function (raw) {
        try {
          const msg = JSON.parse(raw);
          if (msg.type === 'danmaku') danmakuReceived.add(1);
        } catch (_) {}
      });

      socket.on('error', function () {
        wsConnectErrors.add(1);
        wsErrorRate.add(1);  // 记录错误率（分子）
        socket.close();
      });

      // socket 关闭后 setInterval 回调自动停止，无需手动 clearInterval

      // 每次 VU 迭代持续 ~20s，与最长稳定阶段匹配
      socket.setTimeout(() => socket.close(), 20000);
    },
  );

  const ok = res && res.status === 101;
  check(res, { 'WS 升级成功 (101)': () => ok });
  if (ok) {
    wsConnectSuccess.add(1);
    wsErrorRate.add(0);  // 成功计入分母
  } else {
    wsConnectErrors.add(1);
    wsErrorRate.add(1);
  }
}
