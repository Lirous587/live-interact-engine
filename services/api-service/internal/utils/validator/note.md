# 自定义验证规则文档

本文档列举了 api-service 中所有已注册的自定义验证规则。

---

## 验证规则总览

| 规则名称     | 说明           | 正则表达式                            | 使用示例               |
| ------------ | -------------- | ------------------------------------- | ---------------------- |
| `mobile_cn`  | 中国大陆手机号 | `^1[3-9]\d{9}$`                       | `binding:"mobile_cn"`  |
| `hex_color`  | 十六进制颜色码 | `^#([A-Fa-f0-9]{6}\|[A-Fa-f0-9]{3})$` | `binding:"hex_color"`  |
| `domain_url` | 域名 URL       | 见下方详述                            | `binding:"domain_url"` |
| `slug`       | URL 友好字符   | `^[a-zA-Z0-9_-]+$`                    | `binding:"slug"`       |

---

## 详细规则说明

### 1. `mobile_cn` - 中国大陆手机号验证

**规则**: `^1[3-9]\d{9}$`

**说明**:

- 以 1 开头
- 第二位为 3-9（包含不同运营商）
- 后面跟 9 个任意数字
- 总长度 11 位

**有效示例**: `13812345678`, `15912345678`, `18912345678`

---

### 2. `hex_color` - 十六进制颜色码验证

**规则**: `^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`

**说明**:

- 以 # 开头
- 后跟 6 位或 3 位十六进制数字
- 支持大小写

**有效示例**: `#FFF`, `#FFFFFF`, `#FF0000`, `#00FFAA`

---

### 3. `domain_url` - 域名 URL 验证

**验证逻辑**:

1. 合法的 URL 格式
2. 协议必须是 `http` 或 `https`
3. 主机名必须是有效的域名（不允许 IP 地址）
4. 符合域名格式

**有效示例**: `https://example.com`, `http://api.github.com`,
`https://www.lirous.com`

**无效示例**:

- `https://192.168.1.1` (IP地址 - net.ParseIP 检查)
- `http://localhost` (缺少域名点号 - 正则检查)
- `ftp://example.com` (协议不对 - Scheme 检查)

**验证流程**:

1. 检查 URL 格式 ✓
2. 检查协议是 http/https ✓
3. 排除 IP 地址 ✓
4. 正则验证需要 `domain.something` 格式 ✓

---

### 4. `slug` - URL 友好字符验证

**规则**: `^[a-zA-Z0-9_-]+$`

**说明**:

- 只允许：字母、数字、下划线、连字符
- 不允许空格、特殊字符、中文

**有效示例**: `user-profile`, `api_service`, `test123`

**无效**: `user@profile`, `用户资料`, `user.profile`

---

## 快速使用示例

```go
type RegisterRequest struct {
    Phone    string `binding:"required,mobile_cn"`
    Website  string `binding:"required,domain_url"`
    Color    string `binding:"required,hex_color"`
    Username string `binding:"required,slug"`
}
```
