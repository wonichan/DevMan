# DevMan Go 后端 API 文档

> 所有方法通过 Wails v2 暴露给前端，前端通过 `wailsjs/go/main/App` 调用。
> TypeScript 类型定义在 `frontend/src/devman-types.ts` 和 `wailsjs/go/models.ts` 中。

---

## App 方法

### ScanAll

扫描所有内置环境扫描器，返回环境摘要列表。

```go
func (a *App) ScanAll() ([]models.EnvSummary, error)
```

**调用：**
```typescript
import { ScanAll } from '../bindings/go/main/App';
const summaries = await ScanAll();
```

**返回：** `EnvSummary[]`

| 字段 | 类型 | 说明 |
|------|------|------|
| Env | Env | 环境元数据 |
| Instances | EnvInstance[] | 检测到的版本实例 |
| Paths | EnvPath[] | 路径及大小 |
| TotalSize | int64 | 总占用字节 |
| Health | HealthLevel | 健康状态 |

---

### GetEnvs

获取所有已存储的环境列表。

```go
func (a *App) GetEnvs() ([]models.Env, error)
```

**返回：** `Env[]`

| 字段 | 类型 | 说明 |
|------|------|------|
| Id | int64 | 数据库 ID |
| Name | string | 显示名称 |
| Key | string | 唯一标识符 |
| Category | EnvCategory | 分类：runtime/sdk/tool |
| Icon | string | Emoji 图标 |
| Description | string | 描述 |
| Website | string | 官网链接 |
| IsManaged | bool | 是否由 DevMan 管理 |

---

### GetEnvSummary

获取指定环境的完整摘要（含实例和路径）。

```go
func (a *App) GetEnvSummary(key string) (*models.EnvSummary, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| key | string | 环境 Key，如 `"nodejs"`, `"python"` |

---

### Migrate

迁移环境到目标目录。

```go
func (a *App) Migrate(envID int64, targetDir string, useJunction bool) (*migrator.MigrationResult, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| envID | int64 | 环境 ID |
| targetDir | string | 目标目录路径 |
| useJunction | bool | 是否在原位置创建 Junction |

**返回：** `MigrationResult`

| 字段 | 类型 | 说明 |
|------|------|------|
| success | bool | 是否成功 |
| message | string | 结果消息 |
| bytesMoved | int64 | 移动字节数 |
| durationMs | int64 | 耗时毫秒 |

---

### GetDiskInfo

获取磁盘使用信息。

```go
func (a *App) GetDiskInfo() ([]models.DiskInfo, error)
```

**返回：** `DiskInfo[]`

| 字段 | 类型 | 说明 |
|------|------|------|
| Letter | string | 盘符，如 `"C:"` |
| TotalBytes | int64 | 总容量 |
| FreeBytes | int64 | 可用空间 |
| UsedBytes | int64 | 已用空间 |
| UsedPercent | int | 使用百分比 |

**注意：** Windows 使用 WMI 查询，Linux 使用 `statfs`。

---

### GetHistory

获取操作历史记录。

```go
func (a *App) GetHistory(limit int) ([]models.HistoryEntry, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| limit | int | 返回条数上限 |

**返回：** `HistoryEntry[]`

| 字段 | 类型 | 说明 |
|------|------|------|
| Id | int64 | 记录 ID |
| Action | string | 操作类型：`scan`, `migrate`, `clean` |
| TargetEnv | string | 目标环境 Key |
| DetailsJson | string | JSON 详情 |
| Success | bool | 是否成功 |
| ErrorMessage | string | 错误信息 |
| CreatedAt | time.Time | 时间 |

---

### AnalyzeCleanable

分析可清理的缓存项目。

```go
func (a *App) AnalyzeCleanable() ([]models.CleanableItem, error)
```

**返回：** `CleanableItem[]`

| 字段 | 类型 | 说明 |
|------|------|------|
| Name | string | 显示名称 |
| Path | string | 路径 |
| Description | string | 描述 |
| SizeBytes | int64 | 大小 |
| Selected | bool | 默认选中状态 |
| EnvKey | string | 来源环境 Key |

---

### CleanItems

执行清理操作。

```go
func (a *App) CleanItems(items []models.CleanableItem) (int64, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| items | CleanableItem[] | 要清理的项目（Selected=true 的项） |

**返回：** `int64` — 实际释放的字节数

---

## 数据模型

### 环境分类 (EnvCategory)
```go
type EnvCategory string
const (
    CategoryRuntime EnvCategory = "runtime"  // 运行时
    CategorySDK     EnvCategory = "sdk"      // SDK
    CategoryTool    EnvCategory = "tool"      // 工具
)
```

### 路径类型 (PathType)
```go
type PathType string
const (
    PathInstall PathType = "install"  // 安装目录
    PathCache   PathType = "cache"    // 缓存
    PathDeps    PathType = "deps"     // 依赖
    PathConfig  PathType = "config"   // 配置
    PathLog     PathType = "log"      // 日志
    PathTemp    PathType = "temp"     // 临时
    PathData    PathType = "data"     // 数据
)
```

### 健康等级 (HealthLevel)
```go
type HealthLevel string
const (
    HealthHealthy  HealthLevel = "healthy"   // 健康
    HealthInfo     HealthLevel = "info"      // 信息
    HealthWarning  HealthLevel = "warning"   // 警告
    HealthCritical HealthLevel = "critical"  // 危险
)
```

---

## Scanner 接口

```go
type Scanner interface {
    Name() string
    Detect() ([]models.EnvInstance, []models.EnvPath, error)
}
```

**Detect 返回：**
- `EnvInstance[]` — 检测到的版本实例（版本号、路径、是否默认等）
- `EnvPath[]` — 相关路径（安装目录、缓存目录等）
- `error` — 检测错误（不影响其他扫描器继续执行）

---

## 错误处理

所有 Go 方法返回 `(result, error)`，前端调用时：

```typescript
try {
  const result = await SomeMethod();
} catch (e) {
  console.error('操作失败:', e);
  // 显示错误提示给用户
}
```

**常见错误：**
| 错误 | 原因 | 处理建议 |
|------|------|----------|
| engine not initialized | Wails startup 未完成 | 等待页面加载完成 |
| registry not initialized | DB 打开失败 | 检查文件权限 |
| env not found: xxx | 环境 Key 不存在 | 先执行扫描 |
| 目标目录不可写 | 权限不足 | 以管理员运行 |
| 未能计算源目录大小 | 路径不存在 | 检查路径有效性 |
