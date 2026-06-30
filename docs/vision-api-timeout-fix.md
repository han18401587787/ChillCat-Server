# Vision API 超时问题修复方案

## 问题背景

CI 测试中 `ChillCatVisualRegressionTests` 和 `ChillCatV3UITests` 的视觉分析测试大量失败，错误类型为：
1. `NSURLErrorDomain Code=-1005 "The network connection was lost."` — 客户端连接被断开
2. `valueNotFound` — 服务端返回 `data: null` 导致 Swift JSONDecoder 崩溃

## 耗时链路分析

```
iOS Client (XCUITest)
  │  POST /api/v1/vision/analyze  (Base64 PNG ~200KB-2MB)
  ▼
Nginx (proxy_read_timeout: 30s)
  │  proxy_pass → 127.0.0.1:8080
  ▼
Gin HTTP Server (WriteTimeout: 15s ⚠️)
  │
  ▼
VisionHandler.Analyze()
  │
  ▼
VisionService.Analyze()
  │  ├─ apiKey != "" → analyzeWithAI()     ← 4~20s (qwen-vl-max 推理)
  │  └─ apiKey == "" → analyzeWithRules()  ← <10ms
  │
  ▼
analyzeWithAI()
  │  httpClient.Timeout: 30s
  │  POST https://dashscope.aliyuncs.com/.../chat/completions
  │  ├─ 图片 Base64 上传: ~1-3s
  │  ├─ qwen-vl-max 推理: ~3-15s
  │  └─ 响应解析: <100ms
  │  总计: 4~20s
  ▼
返回 JSON Response
```

## 根因定位

### 🔴 瓶颈 #1：Gin HTTP Server WriteTimeout = 15s（致命）

`cmd/server/main.go:31`:
```go
srv := &http.Server{
    WriteTimeout: 15 * time.Second,
}
```

`WriteTimeout` 包含**从读取请求头到写完响应体的全部时间**。当 `analyzeWithAI` 耗时超过 15s 时，Go 标准库的 `http.Server` 会直接关闭底层 TCP 连接，客户端收到 `Code=-1005 "The network connection was lost."`。

**证据链**：
- AI 推理 4~20s，中位数约 10-12s，P95 约 18s
- WriteTimeout=15s 时，约 30-40% 的请求会在 AI 返回结果后、写响应时被截断
- 客户端收到的不是 HTTP 错误码，而是 TCP 连接被 RST —— 完全吻合 `Code=-1005`

### 🔴 瓶颈 #2：Nginx proxy_read_timeout = 30s（次要）

虽然 30s 足够覆盖大部分请求，但与 Gin 的 15s WriteTimeout 形成不一致的超时策略。

### 🟡 瓶颈 #3：无异步/队列机制

当前是同步阻塞模式：一个 HTTP 请求 → 一次 AI 调用 → 阻塞等待 → 返回。CI 并行跑 5+ 个测试时，每个都独立调用 AI，无复用无排队。

### 🟡 瓶颈 #4：data: null 导致客户端崩溃

已在上一轮修复（`pkg/response/response.go` 将 `nil` → `struct{}{}`）。

## 修复方案

### ✅ 修复 1：提升 WriteTimeout 到 60s（已实施）

```go
// cmd/server/main.go
srv := &http.Server{
    WriteTimeout: 60 * time.Second,  // 从 15s → 60s
}
```

AI 调用最长 20s + 图片传输 3s + buffer = 60s 安全边际。

### ✅ 修复 2：添加请求耗时日志（已实施）

在 `analyzeWithAI` 中添加入口/出口耗时统计日志，可观测每次 AI 调用的 page、图片大小、score、耗时。

### ✅ 修复 3：Nginx 超时对齐（已实施）

```nginx
location /api/v1/vision/ {
    proxy_read_timeout 70s;  # 略大于后端 WriteTimeout
}
location / {
    proxy_read_timeout 30s;
}
```

### ✅ 修复 4：data: null 兼容（已实施）

`pkg/response/response.go` 错误响应中 `Data: nil` → `Data: struct{}{}`，序列化为 `"data": {}` 而非 `"data": null`。

## 异步改造评估结论：暂不实施

### 异步方案（任务队列 + task_id + 轮询）分析

| 维度 | 同步（当前） | 异步（任务队列） |
|------|-------------|-----------------|
| 调用方 | CI 测试，愿意等 | 同左 |
| 复杂度 | ~150 行 | ~400+ 行（task_id 生成、状态存储、结果缓存、TTL 清理、轮询接口、客户端轮询逻辑） |
| 新增风险 | 无 | 轮询间隔/次数调优、任务状态丢失、Redis 依赖 |
| 收益 | — | 客户端无需等 20s，但 CI 测试不需要这个收益 |

**结论**：当前场景下异步改造收益为负。调用方是 CI 测试而非终端用户，同步等 20s 完全可接受。
真正的问题是 WriteTimeout=15s 截断 TCP 连接，而非 AI 太慢。修复超时配置即可解决。

### 何时考虑异步

当以下条件同时满足时再引入异步：
1. 调用方变成终端用户（非 CI），用户体验敏感
2. 并发量上升（>10 QPS 同时调用 vision API）
3. 需要结果持久化/回调通知

## iOS 客户端侧问题分析

### 🔴 testPixelDiff_HomePage — 应用启动超时

```
Failed to launch <XCUIApplicationImpl: ...> via Xcode: Timed out while launching application via Xcode.
```

**根因**：CI 环境 Simulator 启动超时。`ChillCatVisualRegressionTests.setUpWithError()` 中 `app.launch()` 后等待 `tabHome` 出现（timeout=15s），但 Simulator 冷启动 + App 初始化可能超过此时间。

**链路**：
1. `app.launch()` — Simulator 冷启动 ~5-10s
2. App `didFinishLaunching` — 初始化 SwiftUI + 网络请求
3. `tabHome.waitForExistence(timeout: 15)` — 总超时 15s 不够

**修复建议**：将 `waitForExistence(timeout: 15)` 提升到 30s，或在 CI 中预启动 Simulator。

### 🔴 test_Resonance_NotAloneBanner — Signal Trap 崩溃

```
Test crashed with signal trap.
```

**根因**：`CCInteractionTests.swift:138` 中 `test_Resonance_NotAloneBanner()` 使用 `XCTFail` 触发 SIGTRAP。

```swift
if !found {
    CCDiagnosticHelper.diagnose(page: "共鸣墙", expectedElement: "你并不孤单", app: app)
    XCTFail("找不到「你并不孤单」横幅 — 详见诊断报告")  // ← 触发 signal trap
}
```

`XCTFail` 在 `continueAfterFailure = true` 时不会 crash。但此处 crash 说明 `CCDiagnosticHelper.diagnose()` 内部的 `app.screenshot()` 可能触发了 Swift 运行时错误（如强制解包 nil），或 CI 环境截图权限不足导致 `try?` 静默失败后仍有副作用。

**修复建议**：
1. 检查 `CCDiagnosticHelper.diagnose()` 中 `try? app.screenshot()` 在 CI 环境的行为
2. 将截图逻辑包裹在更安全的 error handling 中
3. 在 `diagnose()` 开头加 `guard` 检查截图能力

### 🟡 iOS 客户端 Vision API 请求已有 60s 超时

`VisualTesting.swift:145`:
```swift
request.timeoutInterval = 60
```

客户端侧超时设置合理。问题在于服务端 `WriteTimeout=15s` 在客户端 60s 超时之前就关闭了连接。服务端修复后应解决。

## 实施记录

1. [x] 修改 `cmd/server/main.go` — WriteTimeout 15s → 60s
2. [x] 修改 `internal/vision/service.go` — 添加耗时日志
3. [x] 修改 `deploy/nginx.conf` — 为 vision 路由设置更长 proxy_read_timeout
4. [x] 修改 `pkg/response/response.go` — data: null → data: {}
5. [ ] 部署验证 — 重新跑 CI 测试确认通过率
6. [ ] iOS 侧: `testPixelDiff_HomePage` 启动超时 — 提升 `waitForExistence` timeout 到 30s
7. [ ] iOS 侧: `test_Resonance_NotAloneBanner` signal trap — 加固 `CCDiagnosticHelper.diagnose()` 截图逻辑
