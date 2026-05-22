# DevMan 开发计划

> DevMan — Windows 开发环境管理器
> Wails v2 + Go 1.25 + React + Tailwind v4

---

## 当前版本：v0.1.0 (MVP)

**已完成的核心功能：**

| 模块 | 状态 | 说明 |
|------|------|------|
| 环境扫描 | ✅ | Node.js / Python / Java / Go / Flutter / Rust 自动检测 |
| 环境注册表 | ✅ | SQLite 本地持久化，便携模式（exe 同级目录） |
| Dashboard | ✅ | C 盘健康、环境数量、空间占用排行 |
| Environments | ✅ | 环境列表、详情查看、健康状态 |
| Migration Wizard | ✅ | 4 步迁移：选择→配置→预览→执行，支持 Junction + PATH 更新 |
| Cleaner | ✅ | 缓存扫描、选择性清理、历史记录 |
| Versions | ✅ | 多版本管理、默认版本切换 |
| Settings | ✅ | 基础配置 |
| 构建系统 | ✅ | Windows cross-compile 正常，admin manifest + 版本信息 |
| Go 测试 | ✅ | scanner / registry / migrator / models 测试覆盖 |

---

## v0.2.0 路线图 (近期)

### 🔍 增强扫描能力
- [ ] **pnpm / yarn / bun 识别** — 当前只识别 Node.js 本体，需识别包管理器
- [ ] **Docker 环境检测** — Docker Desktop, WSL2 发行版
- [ ] **VS Code / JetBrains 工具链检测**
- [ ] **自定义扫描路径** — 用户可添加非标准安装位置

### 🧹 Cleaner 增强
- [ ] **深度扫描** — node_modules / pip cache / cargo target 等大型目录
- [ ] **智能建议** — 基于占用空间的自动清理建议
- [ ] **定时清理任务** — 支持配置自动清理策略

### 📦 Migration 增强
- [ ] **批量迁移** — 同时迁移多个环境
- [ ] **迁移进度实时显示** — 大文件迁移时的进度条
- [ ] **迁移回滚 UI** — 可视化快照恢复

### 🎨 UI 优化
- [ ] **深色/浅色主题切换** — 当前只有深色主题
- [ ] **Toast 通知系统** — 操作反馈
- [ ] **全局搜索** — 快速跳转到环境/功能

---

## v0.3.0 路线图 (中期)

### 🔧 环境安装管理
- [ ] **一键安装** — 集成 winget / scoop / choco 安装常用环境
- [ ] **版本下载** — 从官方源下载特定版本（Node.js, Python 等）
- [ ] **虚拟环境支持** — Python venv / Node.js nvm-windows 集成

### 🌐 网络代理配置
- [ ] **全局代理管理** — npm / pip / cargo / go 代理统一配置
- [ ] **镜像源切换** — 国内镜像源一键切换（淘宝、清华、中科大）

### 📊 数据可视化
- [ ] **历史趋势图表** — 空间占用变化趋势
- [ ] **环境使用统计** — 活跃环境、使用频率

---

## v1.0.0 路线图 (长期)

### 💻 多平台支持
- [ ] **macOS 完整支持** — Homebrew 集成
- [ ] **Linux 完整支持** — apt / yum / pacman 集成
- [ ] **跨平台数据同步** — 配置文件云同步

### 🤖 智能化
- [ ] **AI 推荐** — 根据项目类型推荐环境配置
- [ ] **冲突检测** — 版本冲突自动识别和解决建议
- [ ] **健康诊断** — 自动检测环境问题并给出修复方案

### 🔌 插件系统
- [ ] **Scanner 插件 API** — 第三方可扩展扫描器
- [ ] **主题插件** — 自定义 UI 主题
- [ ] **集成插件** — CI/CD 工具链集成

---

## 技术债务

| 优先级 | 问题 | 计划版本 |
|--------|------|----------|
| P0 | Windows 构建 `wailsbindings.exe` 32-bit 错误 | v0.1.1 |
| P1 | go.mod 版本要求提升到 1.25 | ✅ 已完成 |
| P1 | 前端 React 18 → 19 升级评估 | v0.2.0 |
| P2 | Tailwind v4 完整功能验证 | v0.2.0 |
| P2 | 错误处理和用户提示完善 | v0.2.0 |

---

## 里程碑时间线

```
v0.1.0  [2025-05]  MVP 完成，基础功能可用
v0.1.1  [2025-06]  修复 Windows 构建问题，稳定性提升
v0.2.0  [2025-07]  扫描增强 + Cleaner 升级 + UI 优化
v0.3.0  [2025-09]  安装管理 + 代理配置 + 数据可视化
v1.0.0  [2025-12]  多平台 + 智能化 + 插件系统
```

---

## 参与贡献

1. Fork 仓库
2. 创建 feature 分支 (`git checkout -b feature/xxx`)
3. 提交更改 (`git commit -m 'feat: xxx'`)
4. 推送到分支 (`git push origin feature/xxx`)
5. 创建 Pull Request

**Commit 规范：** 遵循 Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `build:`)
