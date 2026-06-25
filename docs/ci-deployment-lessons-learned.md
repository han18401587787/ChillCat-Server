# CI/CD 部署问题复盘与经验沉淀

> **项目**: ChillCat-Server（绪安）
> **日期**: 2026-06-25
> **涉及分支**: v3.0-dev
> **CI 平台**: GitHub Actions

---

## 问题总览

| # | 问题 | 类型 | 严重程度 | 根因 |
|---|------|------|----------|------|
| 1 | errcheck lint 未通过 | 代码质量 | 中 | 未检查函数返回值 |
| 2 | CI workflow 文件丢失 | 流程 | 高 | force push 覆盖了分支上的 .github/workflows/ |
| 3 | golangci-lint 版本不兼容 | 工具链 | 高 | golangci-lint v1.64.8 不支持 Go 1.25 |
| 4 | git pull 分支分叉 | 流程 | 高 | SSH action 中环境变量为空 + 分支历史分叉 |
| 5 | 部署后健康检查超时 | 部署 | 中 | sleep 5 秒不足以等待服务完全启动 |
| 6 | PostgreSQL 密码认证失败 | 环境配置 | 致命 | pgdata volume 保留旧密码，与 compose 默认值不一致 |

---

## 问题 1：errcheck lint 未通过

### 现象
```
cmd/seed/main.go:21:16: Error return value of `db.AutoMigrate` is not checked (errcheck)
internal/handler/course_handler.go:58:22: Error return value of `h.commentRepo.Create` is not checked (errcheck)
```

### 根因
Go 的 `errcheck` linter 要求所有返回 error 的函数调用必须检查返回值，项目中两处调用忽略了 error。

### 修复
- `cmd/seed/main.go`: 为 `db.AutoMigrate()` 添加 `if err := ...; err != nil { logger.Fatalf(...) }`
- `internal/handler/course_handler.go`: 为 `commentRepo.Create()` 添加错误检查，失败返回 500

### 经验
1. **所有返回 error 的函数调用都应检查返回值**，即使你认为它"不可能"失败
2. CI 中应始终启用 `errcheck` linter，在 PR 阶段就能拦截此类问题

---

## 问题 2：CI workflow 文件被 force push 覆盖

### 现象
```bash
git push --force origin main:v3.0-dev
```
推送后 GitHub Actions 不再触发任何 CI run。

### 根因
- 本地仓库没有 `.github/workflows/ci.yml`（未被 git 追踪）
- `force push` 用本地仓库完全覆盖了远程分支，导致远程的 workflow 文件被删除
- GitHub Actions 的 workflow 文件**必须存在于对应分支上**才会触发

### 修复
从历史 commit `7fb394b` 恢复了 `ci.yml` 文件并重新推送。

### 经验
1. **force push 前务必确认本地包含所有远程文件**，尤其是 `.github/workflows/`
2. **在本地仓库中保持 workflow 文件的同步**：`git pull` 或 `git fetch` 后再修改
3. 可以考虑设置 GitHub 分支保护规则，禁止 force push 到关键分支
4. **替代方案**：使用 `git push origin main:v3.0-dev`（不带 `--force`），让 Git 在冲突时报错

---

## 问题 3：golangci-lint 版本不兼容 Go 1.25

### 现象
```
Error: can't load config: the Go language version (go1.24) used to build golangci-lint
is lower than the targeted Go version (1.25)
```

### 根因
- `go.mod` 声明 `go 1.25`
- CI 中 `golangci-lint-action@v6` 的 `version: latest` 拉到了 v1.64.8
- golangci-lint v1.64.8 用 Go 1.24 构建，无法解析 `go 1.25` 指令
- golangci-lint 有一条规则：构建它的 Go 版本必须 ≥ 目标代码的 Go 版本

### 修复
```yaml
# 修改前
uses: golangci/golangci-lint-action@v6
with:
  version: latest

# 修改后
uses: golangci/golangci-lint-action@v7
with:
  version: v2.12.2
```
golangci-lint v2.12.2 的 `go.mod` 声明 `go 1.25.0`，与项目匹配。

### 经验
1. **golangci-lint 的版本必须与项目的 Go 版本匹配**：lint 工具的 Go 版本 ≥ 项目的 Go 版本
2. **避免使用 `version: latest`**，应锁定具体版本号
3. 升级 Go 版本时，同步检查 golangci-lint 是否有对应版本
4. 参考：[golangci-lint FAQ - Go version support](https://golangci-lint.run/docs/welcome/faq/)

---

## 问题 4：git pull 分支分叉 + 环境变量为空

### 现象
```
hint: You have divergent branches and need to specify how to reconcile them.
fatal: Need to specify how to reconcile divergent branches.
```

同时 `${GITHUB_REF}` 和 `$GITHUB_SHA` 在 SSH action 中输出为空。

### 根因
1. **环境变量问题**：`appleboy/ssh-action@v1` 在远程服务器上执行脚本，GitHub Actions 的运行时环境变量（如 `GITHUB_REF`、`GITHUB_SHA`）不会自动传递到 SSH session 中
2. **分支分叉**：服务器上本地分支和远程分支历史不一致（之前可能有手动操作或不同的 push 路径）

### 修复
```yaml
# 修改前
echo "分支: ${GITHUB_REF#refs/heads}"
git pull origin ${GITHUB_REF#refs/heads}

# 修改后
BRANCH="${{ github.ref_name }}"   # GitHub Actions 模板变量，编译时注入
git reset --hard origin/$BRANCH   # 强制对齐远程，避免分叉问题
```

### 经验
1. **在 SSH action 中不要依赖 GitHub Actions 运行时环境变量**，应使用 `${{ }}` 模板语法在编译时注入
2. **CI 部署场景使用 `git reset --hard origin/$BRANCH` 替代 `git pull`**，服务器应是只读副本
3. 模板变量和运行时变量的区别：
   - `${{ github.ref_name }}` → 在 workflow 解析阶段替换为字面值
   - `$GITHUB_REF` → 在 runner 上运行时读取环境变量

---

## 问题 5：部署后健康检查超时

### 现象
```
docker compose up -d --build
sleep 5
curl http://localhost:8080/health  # 失败
```
容器状态显示 `Restarting`，但 5 秒后 curl 报 connection refused。

### 根因
- `docker compose up --build` 需要重新编译 Go 二进制（~35 秒） + 数据库迁移（~10 秒）
- `sleep 5` 远不足以等待服务就绪
- 旧的 `sleep 5` 在首次部署或冷启动时大概率失败

### 修复
将固定等待改为**重试循环**：
```bash
for i in $(seq 1 12); do
  sleep 5
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
  if [ "$STATUS" = "200" ]; then
    echo "✅ 部署成功 (尝试 $i)"
    exit 0
  fi
  echo "⏳ 等待服务就绪... ($i/12)"
done
```
最多等待 60 秒（12 次 × 5 秒），一旦返回 200 立即退出。

### 经验
1. **永远不要用固定 sleep 等待服务就绪**，使用重试循环或 `docker compose` 的 healthcheck
2. 为部署脚本添加**完整的失败诊断日志**：容器状态、应用日志、数据库日志、容器详情
3. CI 中的健康检查 curl 应使用 `-w "%{http_code}"` 输出具体状态码，便于排查

---

## 问题 6：PostgreSQL 密码认证失败（致命）

### 现象
```
FATAL: password authentication failed for user "chillcat" (SQLSTATE 28P01)
```
容器持续重启，日志反复输出密码错误。

### 根因
这是整个部署过程中最隐蔽的问题，涉及多个层面的不一致：

1. **PostgreSQL volume 持久化**：`POSTGRES_PASSWORD` 环境变量只在 volume **首次初始化**时写入数据库，之后修改环境变量不会更新已有密码
2. **config.yaml 密码不一致**：代码中 `configs/config.yaml` 的默认密码是 `chillcat_pass`，而 compose 中 postgres 初始化用的是 `chillcat_prod_2026`
3. **环境变量覆盖失效**：虽然代码有 `getEnv("DB_PASSWORD", ...)` 逻辑，但 Docker 镜像构建时 config.yaml 已被打包，如果 compose 没有正确传递环境变量就会用 YAML 中的默认值

### 诊断过程
```
1. 检查 compose 环境变量 → postgres: POSTGRES_PASSWORD=chillcat_prod_2026 ✅
2. 检查 app 容器日志 → password authentication failed ❌
3. 在 postgres 容器内本地连接 → psql -U chillcat 无需密码就连上了（走 socket trust）
4. 说明：本地 socket 是 trust 认证，但 Docker network TCP 连接走 md5 密码认证
5. 执行 ALTER USER 重置密码 → 问题解决
```

### 修复
```sql
-- 在 postgres 容器内重置密码
ALTER USER chillcat WITH PASSWORD 'chillcat_prod_2026';
```
同时将 `configs/config.yaml` 的默认密码也改为 `chillcat_prod_2026`，确保兜底值正确。

### 经验
1. **生产环境密码必须通过环境变量或 Secret 管理**，不要在配置文件中硬编码
2. **PostgreSQL Docker 镜像的密码只在首次初始化时生效**，如果 volume 已存在需要手动 `ALTER USER` 修改
3. **区分 pg_hba.conf 的两种认证路径**：
   - `local` + socket → 通常 trust
   - `host` + TCP/IP → md5/scram-sha-256
4. **调试密码问题的有效方法**：
   - `docker exec <container> env | grep PASSWORD` — 检查实际环境变量
   - `docker exec <container> psql -U <user> -c "SELECT 1"` — 测试本地连接
   - `docker logs <container>` — 查看 postgres 日志中的认证失败详情

---

## 流程改进建议

### 1. 部署脚本标准化
```bash
# 推荐的 CI 部署脚本模板
set -e
BRANCH="${{ github.ref_name }}"

# Git 同步
git fetch origin && git checkout $BRANCH && git reset --hard origin/$BRANCH

# 构建 + 启动
docker compose -f docker-compose.prod.yml up -d --build --remove-orphans --force-recreate

# 健康检查（重试循环 + 诊断日志）
for i in $(seq 1 12); do
  sleep 5
  if curl -sf http://localhost:8080/health; then exit 0; fi
done
# 失败诊断
docker compose ps -a && docker compose logs --tail 50
exit 1
```

### 2. 版本锁定清单
| 工具 | 当前版本 | 备注 |
|------|---------|------|
| Go | 1.25.11 | `go.mod` + CI `setup-go` |
| golangci-lint | v2.12.2 | 必须 ≥ Go 版本 |
| golangci-lint-action | v7 | GitHub Actions |
| PostgreSQL | 16-alpine | 生产与 CI 一致 |
| golang Docker image | 1.25-alpine | 与 go.mod 匹配 |

### 3. 检查清单
部署前确认：
- [ ] `.github/workflows/ci.yml` 存在于本地仓库
- [ ] `golangci-lint` 版本 ≥ `go.mod` 中的 Go 版本
- [ ] `config.yaml` 默认值与 compose 环境变量一致
- [ ] PostgreSQL 密码通过 `ALTER USER` 确认一致
- [ ] 不使用 `git push --force`，改用 `reset --hard` + push

---

## 提交记录

| Commit | 描述 |
|--------|------|
| `fbc416a` | fix: check error return values for errcheck lint |
| `16c870f` | fix: restore CI workflow file lost during force push |
| `a931252` | fix: upgrade golangci-lint to v2.12.2 for Go 1.25 compatibility |
| `d0349a0` | fix: use github context vars + reset --hard to avoid divergent branch error |
| `50e928c` | fix: add retry loop for health check, app startup needs more than 5s |
| `c948e7d` | fix: add container logs output on health check failure for debugging |
| `5e2c77e` | feat: comprehensive deployment logging with full diagnostics on failure |
| `005296a` | fix: remove obsolete version field, add --remove-orphans to compose up |
| `5b2955a` | fix: update default DB password in config.yaml to match compose |
