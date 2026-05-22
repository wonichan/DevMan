# DevMan 系统架构设计

## 技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| 桌面框架 | Wails v2 | 2.12.0 |
| 后端语言 | Go | ≥ 1.25.0 |
| 前端框架 | React | 18.2.0 |
| 构建工具 | Vite | 3.0.7 |
| 样式系统 | Tailwind CSS | 4.3.0 |
| 数据库 | SQLite (mattn/go-sqlite3) | 1.14.44 |
| 类型安全 | TypeScript | 4.6.4 |

---

## 架构总览

```
┌─────────────────────────────────────────────┐
│               Wails v2 Runtime               │
│  ┌─────────────┐      ┌──────────────────┐  │
│  │  WebView2   │◄────►│   Go Backend     │  │
│  │  (Chromium) │ IPC   │   (app.go)       │  │
│  │             │       │                  │  │
│  │  React 18   │       │  Scanner Engine  │  │
│  │  Tailwind 4 │       │  Registry (DB)   │  │
│  │  Vite       │       │  Migrator        │  │
│  └─────────────┘       │  Disk Utils      │  │
│                        └──────────────────┘  │
└─────────────────────────────────────────────┘
                     │
                     ▼
            ┌────────────────┐
            │  SQLite DB File │  (devman.db, 便携模式)
            │  ├─ envs        │
            │  ├─ instances   │
            │  ├─ paths       │
            │  ├─ snapshots   │
            │  ├─ history     │
            │  └─ settings    │
            └────────────────┘
```

---

## 后端架构 (Go)

### 包结构

```
devman/
├── main.go              # 入口：Wails app 初始化
├── app.go               # App 结构体：业务逻辑 orchestration
├── go.mod               # 模块定义
├── wails.json           # Wails 项目配置
│
├── internal/            # 私有包（不对外暴露）
│   ├── models/          # 数据模型定义
│   │   ├── models.go    # 核心结构体 (Env, EnvInstance, EnvPath, etc.)
│   │   └── models_test.go
│   │
│   ├── scanner/         # 环境扫描引擎
│   │   ├── scanner.go   # Engine + Scanner 接口 + ScanAll()
│   │   ├── common.go    # 通用工具 (DirSize, etc.)
│   │   ├── nodejs.go    # Node.js 扫描器
│   │   ├── ...          # Python, Java, Go, Flutter, Rust
│   │   └── *_test.go
│   │
│   ├── registry/        # SQLite 数据持久化
│   │   ├── registry.go  # DB 初始化、CRUD 操作
│   │   └── registry_test.go
│   │
│   ├── migrator/        # 环境迁移引擎
│   │   ├── migrator.go      # 核心迁移逻辑 (A→B→C staging)
│   │   ├── windows.go       # Windows 特有：Junction + PATH
│   │   ├── linux.go         # Linux 占位
│   │   └── migrator_test.go
│   │
│   └── utils/           # 工具函数
│       ├── disk.go          # 磁盘信息（Unix 通用）
│       ├── disk_windows.go  # Windows WMI 磁盘查询
│       └── disk_stub.go   # 构建占位
│
├── pkg/                 # 未来公共包（预留）
│
└── build/               # Wails 构建产物 & 资源
    ├── windows/
    │   ├── wails.exe.manifest   # Admin 权限声明
    │   ├── info.json             # 版本信息
    │   └── icon.ico
    └── appicon.png
```

### 核心组件

#### 1. App (app.go)
- Wails 生命周期管理 (`startup`, `shutdown`)
- 前端可调用的 Go 方法暴露（通过 Wails 自动绑定生成 TypeScript 接口）
- 各子系统的 orchestration（组合 scanner + registry + migrator）

**暴露给前端的方法：**
```go
ScanAll() ([]EnvSummary, error)
GetEnvs() ([]Env, error)
GetEnvSummary(key string) (*EnvSummary, error)
Migrate(envID int64, targetDir string, useJunction bool) (*MigrationResult, error)
GetDiskInfo() ([]DiskInfo, error)
GetHistory(limit int) ([]HistoryEntry, error)
AnalyzeCleanable() ([]CleanableItem, error)
CleanItems(items []CleanableItem) (int64, error)
```

#### 2. Scanner Engine (internal/scanner)
- **接口设计**：`Scanner` 接口，每个环境类型实现 `Name()` + `Detect()`
- **自动发现**：Engine 遍历所有内置 Scanner，收集实例和路径信息
- **数据持久化**：扫描结果自动写入 registry

```go
type Scanner interface {
    Name() string
    Detect() ([]EnvInstance, []EnvPath, error)
}
```

**内置扫描器：**
| 扫描器 | 检测方式 | 支持路径类型 |
|--------|----------|-------------|
| NodeScanner | registry / 常见安装目录 / PATH | install, cache, deps |
| PythonScanner | py launcher / 常见目录 / PATH | install, cache, deps |
| JavaScanner | JAVA_HOME / 注册表 / PATH | install, cache |
| GoScanner | GOROOT / GOPATH | install, cache, deps |
| FlutterScanner | flutter 命令 / 常见目录 | install, cache, deps |
| RustScanner | rustup / cargo | install, cache, deps |

#### 3. Registry (internal/registry)
- **便携模式**：DB 文件放在 exe 同级目录，无需安装
- **表结构**：
  - `envs` — 环境元数据
  - `instances` — 版本实例
  - `paths` — 路径及大小
  - `snapshots` — 迁移快照（JSON 序列化）
  - `history` — 操作历史日志
  - `settings` — 用户配置

#### 4. Migrator (internal/migrator)
- **A→B→C Staging 模式**：
  1. **A (Source)** — 原始安装目录
  2. **B (Staging)** — `.devman_tmp/` 下临时复制 + 验证
  3. **C (Target)** — 确认后重命名为最终目录
- **安全机制**：快照备份 + 错误回滚 + 文件校验
- **Windows 特有**：
  - **Junction** — 旧路径创建目录连接符，兼容已有项目
  - **PATH 更新** — 自动修改系统环境变量

---

## 前端架构 (React + TypeScript)

### 目录结构

```
frontend/
├── index.html           # 入口 HTML
├── package.json         # 依赖
├── vite.config.ts       # Vite 配置
├── tsconfig.json        # TS 配置
├── tailwind.config.js   # Tailwind 主题定制
├── postcss.config.js    # PostCSS 配置
│
├── src/
│   ├── main.tsx         # React 挂载点
│   ├── App.tsx          # 路由/页面切换 (6 个页面)
│   ├── App.css          # 全局样式
│   ├── style.css        # Tailwind 指令 + 滚动条定制
│   ├── devman-types.ts  # 共享类型定义
│   │
│   ├── components/      # 可复用组件
│   │   ├── Sidebar.tsx      # 左侧导航栏
│   │   └── Panel.tsx        # 卡片面板容器
│   │
│   ├── pages/           # 页面级组件
│   │   ├── Dashboard.tsx    # 仪表盘
│   │   ├── Environments.tsx # 环境管理
│   │   ├── Migration.tsx    # 迁移向导 (4 步)
│   │   ├── Cleaner.tsx      # 清理工具
│   │   ├── Versions.tsx     # 版本管理
│   │   └── Settings.tsx     # 设置
│   │
│   ├── bindings/        # Wails 生成的 Go 绑定
│   │   └── go/main/App.ts
│   └── assets/          # 静态资源
│       ├── images/
│       └── fonts/
│
└── wailsjs/             # Wails 运行时 + 模型生成
    ├── go/
    │   ├── main/App.d.ts    # Go 方法的 TS 声明
    │   ├── main/App.js      # Go 方法的 JS 封装
    │   └── models.ts        # Go 结构体的 TS 映射
    └── runtime/
```

### 设计系统

**色彩方案（深色主题）：**
```
bg:           #0F172A    // 主背景
bg-soft:      #111B2E    // 次级背景
bg-deep:      #08111E    // 深层背景
panel:        #101A2B    // 面板背景
panel-raised: #16243B    // 凸起面板
border:       #233552    // 边框
accent:       #22C55E    // 强调色（绿）
info:         #38BDF8    // 信息色（蓝）
warning:      #F59E0B    // 警告色（橙）
danger:       #EF4444    // 危险色（红）
text-primary: #F8FAFC    // 主文字
text-muted:   #94A3B8    // 次要文字
```

**圆角系统：**
- `card`: 22px — 大卡片
- `panel`: 24px — 面板容器
- `badge`: 12px — 标签徽章
- `button`: 12px — 按钮
- `input`: 16px — 输入框

**字体：**
- 正文：`IBM Plex Sans` / `Inter` → `Noto Sans`
- 等宽：`JetBrains Mono` / `Fira Code` → `Consolas`

### 页面路由

```tsx
type Page = 'dashboard' | 'environments' | 'migration' | 'cleaner' | 'versions' | 'settings';
```

使用 `useState` 管理当前页面，Sidebar 控制切换。无真实路由库，适合单页桌面应用。

---

## 数据流

### 扫描流程

```
[用户点击"刷新"]
    │
    ▼
[Dashboard.tsx]  ScanAll()  ──►  [App.go]
    │                              │
    │                              ▼
    │                         [scanner.Engine]
    │                              │
    │                              ├── NodeScanner.Detect()
    │                              ├── PythonScanner.Detect()
    │                              ├── ...
    │                              │
    │                              ▼
    │                         [registry.SaveEnv/Instance/Path]
    │                              │
    │                              ▼
    │                         [App.go 组装 EnvSummary[]]
    │                              │
    ◄──────────────────────────────┘
    │
    ▼
[React State Update] 渲染 Dashboard
```

### 迁移流程

```
[Migration.tsx]
    │
    ▼ Step 1: 选择环境
[App.GetEnvs()] ──► 显示环境列表
    │
    ▼ Step 2: 配置目标路径
用户输入 targetDir + junction 选项
    │
    ▼ Step 3: 预览
[App.GetEnvSummary(key)] ──► 显示占用空间
    │
    ▼ Step 4: 执行迁移
[App.Migrate(envID, targetDir, useJunction)]
    │
    ├── 1. preCheck()     检查目标目录可写
    ├── 2. createSnapshot() 创建 DB 快照
    ├── 3. copyDir()      A → B (staging)
    ├── 4. verifyCopy()   文件校验
    ├── 5. updateEnvVars() PATH 更新
    ├── 6. os.Rename()    B → C (final)
    ├── 7. createJunction() 旧路径创建连接符
    └── 8. SaveHistory()  记录操作日志
    │
    ▼
返回 MigrationResult 到前端
```

---

## 安全设计

1. **Admin 权限**：Windows manifest 声明 `requireAdministrator`，确保 PATH 修改和 Junction 创建成功
2. **Staging 模式**：迁移不直接覆盖，先复制到临时目录验证后再提交
3. **快照回滚**：每次迁移前自动导出 registry 快照，失败可恢复
4. **文件校验**：复制后对比文件数量和总大小
5. **便携模式**：不写入系统目录，无注册表污染，可放 U 盘带走

---

## 构建配置

### Wails 配置 (wails.json)
```json
{
  "name": "devman",
  "outputfilename": "devman",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "author": { "name": "yachiyo", "email": "1012403910@qq.com" }
}
```

### 构建命令
```bash
# 开发模式（热重载）
wails dev

# Windows 生产构建
wails build -platform windows/amd64

# 输出
build/bin/devman.exe  (~12MB, 含 WebView2 运行时检测)
```

### Windows 特有构建产物
- `wails.exe.manifest` — UAC 管理员权限请求
- `info.json` — 文件版本信息 (FileVersion, ProductVersion)

---

## 扩展性设计

### 新增扫描器

1. 在 `internal/scanner/` 创建 `xxx.go`
2. 实现 `Scanner` 接口
3. 在 `Engine.scanners` 切片中注册
4. 在 `modelsForScanner()` 中添加映射

### 新增页面

1. 在 `frontend/src/pages/` 创建组件
2. 在 `App.tsx` 的 `Page` 联合类型中添加路由名
3. 在 `Sidebar.tsx` 中添加导航项
4. 在 `App.tsx` 的 JSX 中添加条件渲染

---

## 性能考虑

| 场景 | 策略 |
|------|------|
| 大目录扫描 | 异步执行，前端显示 loading 状态 |
| 大文件迁移 | 分段复制，预留进度回调接口 |
| DB 查询 | 预加载 + 内存缓存（EnvSummary） |
| 前端渲染 | 虚拟列表预留（环境数量 >100 时） |
| 磁盘空间计算 | 异步 + debounce，避免频繁 IO |
