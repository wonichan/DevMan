# DevMan

> 你的 Windows 开发环境管家 — 一键扫描、迁移、清理、管理开发工具链。

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Wails](https://img.shields.io/badge/Wails-v2-ff4d4d?logo=wails)](https://wails.io)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://react.dev)
[![Tailwind](https://img.shields.io/badge/Tailwind-4-06B6D4?logo=tailwindcss)](https://tailwindcss.com)

---

## 功能特性

| 功能 | 说明 |
|------|------|
| 🔍 **自动扫描** | 检测 Node.js / Python / Java / Go / Flutter / Rust 等开发环境 |
| 📦 **环境迁移** | 将开发环境从 C 盘迁移到 D 盘，自动更新 PATH，支持 Junction 回链 |
| 🧹 **一键清理** | 扫描并清理各环境的缓存目录，释放磁盘空间 |
| 📋 **版本管理** | 多版本共存，一键切换默认版本 |
| 💾 **便携模式** | 数据库存放在 exe 同级目录，无需安装，U 盘即走 |
| 📊 **空间监控** | 实时监控磁盘健康状态，可视化空间占用 |

---

## 截图

> TODO: 添加实际运行截图

---

## 快速开始

### 环境要求

- **Go** >= 1.25.0
- **Node.js** >= 18.0
- **Wails CLI** >= v2.12.0
- **Windows** 10/11 (当前主要支持平台)
- **Mingw-w64** (用于 CGO 编译 sqlite3)

### 安装依赖

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 克隆仓库
git clone https://github.com/wonichan/DevMan.git
cd DevMan

# 前端依赖
cd frontend
npm install
cd ..
```

### 开发模式

```bash
wails dev
```

### 构建

```bash
# Windows 生产构建 (含管理员权限 manifest)
wails build -platform windows/amd64

# 输出: build/bin/DevMan.exe
```

---

## 技术栈

| 层级 | 技术 |
|------|------|
| 桌面框架 | Wails v2 (WebView2 + Go) |
| 后端 | Go 1.25+ |
| 前端 | React 18 + TypeScript |
| 样式 | Tailwind CSS 4 |
| 构建 | Vite 3 |
| 数据库 | SQLite (mattn/go-sqlite3) |

---

## 项目结构

```
DevMan/
├── docs/                    # 项目文档
│   ├── PLAN.md             # 开发计划与路线图
│   ├── ARCHITECTURE.md     # 系统架构设计
│   ├── FRONTEND.md         # 前端 UI/UX 设计
│   └── API.md              # Go 后端 API 文档
│
├── frontend/                # React 前端
│   ├── src/pages/          # 6 个页面 (Dashboard, Environments...)
│   ├── src/components/     # 可复用组件 (Sidebar, Panel)
│   └── wailsjs/            # Wails 自动生成的 Go 绑定
│
├── internal/                # Go 私有包
│   ├── models/             # 数据模型
│   ├── scanner/            # 环境扫描引擎
│   ├── registry/           # SQLite 持久化
│   ├── migrator/           # 环境迁移引擎
│   └── utils/              # 工具函数
│
├── build/                   # Wails 构建配置
│   └── windows/            # Windows manifest & 图标
│
├── app.go                   # Wails App 入口
├── main.go                  # 程序入口
├── go.mod                   # Go 模块
└── wails.json               # Wails 项目配置
```

---

## 详细文档

- [开发计划 & 路线图](./PLAN.md)
- [系统架构设计](./ARCHITECTURE.md)
- [前端 UI/UX 设计](./FRONTEND.md)
- [后端 API 文档](./API.md)

---

## 开发指南

### 新增环境扫描器

1. 在 `internal/scanner/` 创建 `xxx.go`
2. 实现 `Scanner` 接口：
   ```go
   type MyScanner struct{}
   func (s *MyScanner) Name() string { return "MyLang" }
   func (s *MyScanner) Detect() ([]models.EnvInstance, []models.EnvPath, error) {
       // 检测逻辑
   }
   ```
3. 在 `scanner.go` 的 `NewEngine` 中注册

### 新增前端页面

1. 在 `frontend/src/pages/` 创建组件
2. 在 `App.tsx` 的 `Page` 类型和渲染逻辑中添加路由
3. 在 `Sidebar.tsx` 中添加导航项

---

## 路线图

| 版本 | 计划 |
|------|------|
| v0.1.0 | ✅ MVP：扫描、迁移、清理、版本管理 |
| v0.2.0 | 扫描增强（Docker、IDE）、Cleaner 升级、主题切换 |
| v0.3.0 | 一键安装、代理配置、数据可视化 |
| v1.0.0 | macOS/Linux 支持、AI 推荐、插件系统 |

详见 [PLAN.md](./PLAN.md)

---

## License

MIT © wonichan
