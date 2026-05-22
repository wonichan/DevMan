# DevMan 开发进度总览

> 文档更新时间：2026-05-23
> 当前版本：v0.1.1 紧急修复版

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
| Dashboard | ✅ UI + 绑定调用 | 调用 `ScanAll()` + `GetDiskInfo()`，C 盘健康卡片、环境空间排行条形图 |
| Environments | ✅ UI 骨架 | 环境列表卡片（Icon + 版本 + 健康状态 + 路径分类） |
| Migration | ✅ UI 骨架 | 4 步向导：选择环境 → 配置目标 → 预览确认 → 执行结果 |
| Cleaner | ✅ UI 骨架 | 可清理项目列表、复选框、清理按钮 |
| Versions | ✅ UI 骨架 | 多版本列表、默认版本单选切换 UI |
| Settings | ✅ UI 骨架 | 通用设置、扫描路径、数据管理界面 |

### 3. Go 后端核心模块

| 模块 | 完成度 | 说明 |
|------|--------|------|
| `internal/models` | 100% | 所有数据结构完整（Env / Instance / Path / Summary / Disk / Migration / Snapshot / History / Cleanable） |
| `internal/registry` | ~96% | 6 张表 + CRUD + 快照导出/导入 + 历史记录 + `ON CONFLICT` 更新，已清理 `ListInstances` 未使用变量 |
| `internal/scanner` | ~95% | Engine + 6 个扫描器（Node.js / Python / Java / Go / Flutter / Rust），已执行真实版本号检测 |
| `internal/migrator` | ~85% | A→B→C staging 模式 + 快照回滚 + 文件校验 + Windows Junction + PATH 更新 |
| `internal/utils` | ~90% | 磁盘信息查询（Windows WMI / Linux statfs） |
| `app.go` | ~90% | 所有暴露给前端的方法已定义（ScanAll / GetEnvs / Migrate / GetDiskInfo / CleanItems 等） |

---

## 二、已知问题 / 待完善

| 优先级 | 问题 | 影响范围 | 复现/说明 |
|--------|------|----------|----------|
| **已修复** | Windows 构建 `wailsbindings.exe: %1 is not a valid Win32 application` | 构建系统 | 已确认 `GOOS=windows` / `GOARCH=amd64` 正常；全局 Wails CLI 已匹配项目依赖 `v2.12.0`，`wails build` 可产出 `build/bin/devman.exe` |
| **已修复** | 版本号全是 `"detected"` 占位符 | Dashboard / Environments | 扫描器已执行 `node --version` / `python --version` / `java -version` 等真实版本命令，并设置 3 秒超时 |
| **已修复** | `ListInstances` 中有未使用变量 (`id`, `eid`) | 代码质量 | 已移除未使用变量，`gopls` 无诊断错误 |
| **P2** | 前端错误处理薄弱 | 用户体验 | 部分页面仍有 `console.error`，未全部接入 Toast 用户反馈 |
| **P2** | Settings 页面配置持久化未验证 | 设置功能 | 需确认是否真正写入了 registry SQLite |
| **已核实** | Migration 预览步骤真实数据来源 | 迁移向导 | `Migration.tsx` 已调用 `GetEnvSummary` 汇总环境详情 |

---

## 三、待开发功能

### v0.1.1 — 紧急修复
- [x] **修复 Windows 构建 Bug** — `wailsbindings.exe` 32-bit 错误（P0）
- [x] **真实版本号检测** — `node --version` / `python --version` / `java -version` 等命令执行并解析
- [x] **消除编译 warning** — 清理未使用变量
- [x] **修复 Windows 原生测试路径** — registry / migrator 测试改用 `t.TempDir()`，不再依赖 `/tmp`

### v0.2.0 — 近期增强
- [ ] **主题切换** — 深色 / 浅色 / 跟随系统（当前仅深色）
- [ ] **Toast 通知系统** — 操作成功/失败的用户反馈
- [ ] **确认对话框** — 迁移/删除/重置前的二次确认
- [ ] **全局搜索** — 顶部搜索框，快速跳转到环境或功能页面
- [ ] **新增扫描器** — Docker / VS Code / JetBrains / pnpm / yarn / bun
- [ ] **Cleaner 深度扫描** — `node_modules` / `.gradle` / cargo target / pip cache 等大型目录识别
- [ ] **迁移进度实时显示** — 大文件复制时的进度条回调接口
- [ ] **自定义扫描路径** — 用户可添加非标准安装位置到 Settings

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
| 前端 UI | **~70%** | 页面全有，交互和反馈待完善 |
| Go 后端 | **~85%** | 核心逻辑都在，版本检测已落地，边界条件仍可加强 |
| 测试覆盖 | **~60%** | 有测试文件，`go test ./...` 已在 Windows 环境通过 |
| **Windows 构建** | **100%** | `wails build` 已通过，可产出 `build/bin/devman.exe` |
| 文档 | **100%** | PLAN / ARCHITECTURE / FRONTEND / API / STATUS 齐全 |

### 总体评估
**v0.1.1 紧急修复已完成，Windows 构建阻塞已解除，项目可进入正常本地测试与迭代开发循环。**

验证结果：`go test ./...`、`npm run build`、`wails build` 均已通过。

---

## 五、下一步优先级建议

### 立即做（今天）
1. **前端错误处理** — 添加全局 Toast 组件，替换剩余 `console.error`
2. **Settings 持久化验证** — 确保配置变更真正写入 SQLite

### 本周做
3. 验证 `wails dev` 热重载流程
4. **Cleaner 深度扫描** — 识别大型缓存目录
5. **新增扫描器** — Docker / pnpm / yarn

### 下周做
6. **主题切换** — 浅色模式 + 系统跟随
7. **全局搜索** — 顶部搜索框，快速跳转到环境或功能页面
8. **迁移进度实时显示** — 大文件复制时的进度条回调接口

---

## 六、相关文档

- [PLAN.md](./PLAN.md) — 完整路线图 v0.1.0 → v1.0.0
- [ARCHITECTURE.md](./ARCHITECTURE.md) — 系统架构与数据流
- [FRONTEND.md](./FRONTEND.md) — UI/UX 设计规范
- [API.md](./API.md) — Go 后端 API 参考
