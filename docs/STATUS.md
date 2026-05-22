# DevMan 开发进度总览

> 文档更新时间：2026-05-23
> 当前版本：v0.2.0 功能增强版

---

## 一、已完成的部分

### 1. 项目骨架 & 构建系统

| 项目 | 状态 | 说明 |
|------|------|------|
| Wails v2 + Go 1.25 + React 18 + Tailwind v4 | ✅ 完整 | 技术栈全部就位 |
| Windows 管理员 manifest | ✅ | `wails.exe.manifest` + `info.json` 版本信息 |
| Go 测试 | ✅ | scanner / registry / migrator / models 均有 `_test.go` |
| 便携模式 SQLite | ✅ | DB 自动创建在 exe 同级目录 |
| 根目录 README | ✅ | 含 badge、功能列表、快速开始 |

### 2. 前端 6 个页面 UI 骨架

| 页面 | 状态 | 说明 |
|------|------|------|
| Dashboard | ✅ UI + 绑定调用 | 调用 `ScanAll()` + `GetDiskInfo()`，刷新成功/失败接入 Toast |
| Environments | ✅ UI + 反馈 | 环境列表卡片、搜索、刷新失败 Toast |
| Migration | ✅ 向导 + 实时进度 | 4 步向导、确认对话框、`migration:progress` 实时进度事件 |
| Cleaner | ✅ 深度扫描 UI | 可清理项目列表、风险/类别展示、确认清理、清理后重新分析 |
| Versions | ✅ UI + 反馈 | 多版本列表、加载失败 Toast |
| Settings | ✅ 持久化配置 | 设置从 SQLite 加载/保存，自定义扫描路径、主题策略、迁移确认策略 |

### 3. Go 后端核心模块

| 模块 | 完成度 | 说明 |
|------|--------|------|
| `internal/models` | 100% | 已包含 Env / Instance / Path / Summary / Disk / Migration / Snapshot / History / Cleanable / AppSettings / MigrationProgress |
| `internal/registry` | ~98% | SQLite schema 增加 `settings` 表，支持设置默认值、保存、覆盖和读取 |
| `internal/scanner` | ~98% | Engine 支持自定义扫描路径，已注册 Node.js / Python / Java / Go / Flutter / Rust / Docker / pnpm / Yarn / Bun |
| `internal/migrator` | ~90% | A→B→C staging 模式 + 快照回滚 + 文件校验 + Windows Junction + PATH 更新 + Wails 进度事件 |
| `internal/utils` | ~90% | 磁盘信息查询（Windows WMI / Linux statfs） |
| `app.go` | ~90% | 所有暴露给前端的方法已定义（ScanAll / GetEnvs / Migrate / GetDiskInfo / CleanItems 等） |

---

## 二、已知问题 / 待完善

| 优先级 | 问题 | 影响范围 | 复现/说明 |
|--------|------|----------|----------|
| **已修复** | Windows 构建 `wailsbindings.exe: %1 is not a valid Win32 application` | 构建系统 | 已确认 `GOOS=windows` / `GOARCH=amd64` 正常；全局 Wails CLI 已匹配项目依赖 `v2.12.0`，`wails build` 可产出 `build/bin/devman.exe` |
| **已修复** | 版本号全是 `"detected"` 占位符 | Dashboard / Environments | 扫描器已执行 `node --version` / `python --version` / `java -version` 等真实版本命令，并设置 3 秒超时 |
| **已修复** | `ListInstances` 中有未使用变量 (`id`, `eid`) | 代码质量 | 已移除未使用变量，`gopls` 无诊断错误 |
| **已修复** | 前端错误处理薄弱 | 用户体验 | 业务页面加载/刷新/迁移/清理失败已接入 Toast，前端业务页不再保留 `console.error` |
| **已修复** | Settings 页面配置持久化未验证 | 设置功能 | 已新增 SQLite `settings` 表和 `GetSettings` / `SaveSettings`，测试覆盖默认值与覆盖保存 |
| **已核实** | Migration 预览步骤真实数据来源 | 迁移向导 | `Migration.tsx` 已调用 `GetEnvSummary` 汇总环境详情 |

---

## 三、待开发功能

### v0.1.1 — 紧急修复
- [x] **修复 Windows 构建 Bug** — `wailsbindings.exe` 32-bit 错误（P0）
- [x] **真实版本号检测** — `node --version` / `python --version` / `java -version` 等命令执行并解析
- [x] **消除编译 warning** — 清理未使用变量
- [x] **修复 Windows 原生测试路径** — registry / migrator 测试改用 `t.TempDir()`，不再依赖 `/tmp`

### v0.2.0 — 近期增强
- [x] **主题策略基础** — `dark` / `system` 设置持久化，当前深色主题保持稳定
- [x] **Toast 通知系统** — 操作成功/失败的用户反馈
- [x] **确认对话框** — 迁移、清理、自定义路径删除等关键操作前可取消
- [x] **全局搜索** — 顶部入口 + `Ctrl+K`，可搜索页面、操作和环境
- [x] **新增扫描器** — Docker / pnpm / yarn / bun
- [x] **Cleaner 深度扫描** — npm/pnpm/yarn/pip/gradle/cargo/go/bun 缓存与 `node_modules` / `target` 识别
- [x] **迁移进度实时显示** — 后端按步骤 emit `migration:progress`，前端实时展示进度条与日志
- [x] **自定义扫描路径** — 用户可在 Settings 添加非标准安装位置，`ScanAll()` 自动读取

### v0.3.0 — 中期扩展
- [ ] **一键安装环境** — 集成 winget / scoop / choco 安装常用开发工具
- [ ] **网络代理统一配置** — npm / pip / cargo / go 镜像源一键切换（淘宝/清华/中科大）
- [ ] **数据可视化图表** — 空间占用历史趋势、环境使用统计
- [ ] **快捷键支持** — `Ctrl+R` 刷新、`Ctrl+K` 搜索、`Ctrl+数字` 切换页面
- [ ] **自动清理策略** — 定时任务扫描并提醒清理

### v1.0.0 — 长期目标
- [ ] **macOS 完整支持** — Homebrew 集成、Apple Silicon 适配
- [ ] **Linux 完整支持** — apt / yum / pacman / snap 集成
- [ ] **AI 智能推荐** — 根据项目类型（package.json / Cargo.toml / go.mod）推荐环境配置
- [ ] **插件系统** — Scanner 插件 API、主题插件、第三方集成
- [ ] **配置云同步** — 跨设备同步用户配置和环境数据

---

## 四、当前完成度估算

| 维度 | 完成度 | 备注 |
|------|--------|------|
| 项目骨架 | **100%** | 完整可用 |
| 前端 UI | **~90%** | v0.2.0 反馈、设置、搜索、清理、迁移进度交互已落地 |
| Go 后端 | **~92%** | Settings、扩展扫描器、自定义路径、Cleaner 深度扫描和迁移事件已落地 |
| 测试覆盖 | **~65%** | registry settings 持久化测试已补充，`go test ./...` 已通过 |
| **Windows 构建** | **100%** | `wails build` 已通过，可产出 `build/bin/devman.exe` |
| 文档 | **100%** | PLAN / ARCHITECTURE / FRONTEND / API / STATUS 齐全 |

### 总体评估
**v0.2.0 阶段一、二、三计划功能已完成，项目可进入发布前手动 QA 与边界场景验证。**

验证结果：`go test ./...`、`npm run build`、`wails build` 均已通过。

---

## 五、下一步优先级建议

### 立即做（发布前）
1. 在真实 Windows 开发机上执行手动 QA：扫描、设置保存重启、Cleaner 取消/执行、Migration 进度展示。
2. 验证 Docker / pnpm / yarn / bun 在多种安装方式下的路径识别准确性。

### 后续迭代
3. VS Code / JetBrains 扫描器可作为 v0.2.x 增量补充。
4. 完整浅色主题建议单独做 token 化重构，不混入 v0.2.0 发布。
5. Cleaner 可继续扩展项目根目录发现策略，避免只扫描当前工作目录。

---

## 六、相关文档

- [PLAN.md](./PLAN.md) — 完整路线图 v0.1.0 → v1.0.0
- [ARCHITECTURE.md](./ARCHITECTURE.md) — 系统架构与数据流
- [FRONTEND.md](./FRONTEND.md) — UI/UX 设计规范
- [API.md](./API.md) — Go 后端 API 参考
