# DevMan 开发进度总览

> 文档更新时间：2025-05-22
> 当前版本：v0.1.0 MVP

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
| `internal/registry` | ~95% | 6 张表 + CRUD + 快照导出/导入 + 历史记录 + `ON CONFLICT` 更新 |
| `internal/scanner` | ~90% | Engine + 6 个扫描器（Node.js / Python / Java / Go / Flutter / Rust），但版本号检测是占位符 |
| `internal/migrator` | ~85% | A→B→C staging 模式 + 快照回滚 + 文件校验 + Windows Junction + PATH 更新 |
| `internal/utils` | ~90% | 磁盘信息查询（Windows WMI / Linux statfs） |
| `app.go` | ~90% | 所有暴露给前端的方法已定义（ScanAll / GetEnvs / Migrate / GetDiskInfo / CleanItems 等） |

---

## 二、已知问题 / 待完善

| 优先级 | 问题 | 影响范围 | 复现/说明 |
|--------|------|----------|----------|
| **P0** | Windows 构建 `wailsbindings.exe: %1 is not a valid Win32 application` | **无法本地构建** | 可能是 `GOOS` 环境变量被污染为 `linux`，或临时可执行文件格式错误 |
| **P1** | 版本号全是 `"detected"` 占位符 | Dashboard / Environments 显示不准确 | `NodeScanner.readVersion()` 等未真正执行 `node --version` / `python --version` |
| **P1** | `ListInstances` 中有未使用变量 (`id`, `eid`) | 代码质量 | 编译产生 warning，不影响功能 |
| **P2** | 前端错误处理薄弱 | 用户体验 | 大量 `console.error`，缺少用户可见的 Toast / 弹窗提示 |
| **P2** | Settings 页面配置持久化未验证 | 设置功能 | 需确认是否真正写入了 registry SQLite |
| **P2** | Migration 预览步骤可能未调用真实数据 | 迁移向导 | 需确认 `GetEnvSummary` 是否在前端正确调用并显示 |

---

## 三、待开发功能

### v0.1.1 — 紧急修复
- [ ] **修复 Windows 构建 Bug** — `wailsbindings.exe` 32-bit 错误（P0）
- [ ] **真实版本号检测** — `node --version` / `python --version` / `java -version` 等命令执行并解析
- [ ] **消除编译 warning** — 清理未使用变量

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
| Go 后端 | **~80%** | 核心逻辑都在，版本检测和边界条件待加强 |
| 测试覆盖 | **~60%** | 有测试文件，但覆盖率未知 |
| **Windows 构建** | **0%** | P0 Bug 阻塞，无法产出可执行文件 |
| 文档 | **100%** | PLAN / ARCHITECTURE / FRONTEND / API / STATUS 齐全 |

### 总体评估
**v0.1.0 MVP 功能骨架已完成约 85%，但有一个致命的 Windows 构建阻塞 Bug 未解决。**

修复该 Bug 后，项目即可进入正常的本地测试 → 迭代开发循环。

---

## 五、下一步优先级建议

### 立即做（今天）
1. **修复 Windows 构建** — 检查 `GOOS` 环境变量，清理临时文件，重新编译
2. 验证修复后 `wails build` 和 `wails dev` 都能正常运行

### 本周做
3. **真实版本号检测** — 所有扫描器执行 `--version` 并解析输出
4. **前端错误处理** — 添加全局 Toast 组件，替换 `console.error`
5. **Settings 持久化验证** — 确保配置变更真正写入 SQLite

### 下周做
6. **主题切换** — 浅色模式 + 系统跟随
7. **Cleaner 深度扫描** — 识别大型缓存目录
8. **新增扫描器** — Docker / pnpm / yarn

---

## 六、相关文档

- [PLAN.md](./PLAN.md) — 完整路线图 v0.1.0 → v1.0.0
- [ARCHITECTURE.md](./ARCHITECTURE.md) — 系统架构与数据流
- [FRONTEND.md](./FRONTEND.md) — UI/UX 设计规范
- [API.md](./API.md) — Go 后端 API 参考
