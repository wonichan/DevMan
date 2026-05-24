package models

import "time"

type EnvCategory string

const (
	CategoryRuntime EnvCategory = "runtime"
	CategorySDK     EnvCategory = "sdk"
	CategoryTool    EnvCategory = "tool"
)

type PathType string

const (
	PathInstall PathType = "install"
	PathCache   PathType = "cache"
	PathDeps    PathType = "deps"
	PathConfig  PathType = "config"
	PathLog     PathType = "log"
	PathTemp    PathType = "temp"
	PathData    PathType = "data"
)

type HealthLevel string

const (
	HealthHealthy  HealthLevel = "healthy"
	HealthInfo     HealthLevel = "info"
	HealthWarning  HealthLevel = "warning"
	HealthCritical HealthLevel = "critical"
)

type Env struct {
	ID          int64       `json:"Id"`
	Name        string      `json:"Name"`
	Key         string      `json:"Key"`
	Category    EnvCategory `json:"Category"`
	Icon        string      `json:"Icon"`
	Description string      `json:"Description"`
	Website     string      `json:"Website"`
	IsManaged   bool        `json:"IsManaged"`
	CreatedAt   time.Time   `json:"CreatedAt"`
	UpdatedAt   time.Time   `json:"UpdatedAt"`
}

type EnvInstance struct {
	ID          int64     `json:"Id"`
	EnvID       int64     `json:"EnvId"`
	Version     string    `json:"Version"`
	InstallPath string    `json:"InstallPath"`
	IsDefault   bool      `json:"IsDefault"`
	IsActive    bool      `json:"IsActive"`
	Source      string    `json:"Source"`
	DetectedAt  time.Time `json:"DetectedAt"`
}

type EnvPath struct {
	ID         int64     `json:"Id"`
	EnvID      int64     `json:"EnvId"`
	InstanceID *int64    `json:"InstanceId,omitempty"`
	Type       PathType  `json:"Type"`
	Path       string    `json:"Path"`
	SizeBytes  int64     `json:"SizeBytes"`
	IsMovable  bool      `json:"IsMovable"`
	LastSized  time.Time `json:"LastSized"`
}

type EnvSummary struct {
	Env       Env           `json:"Env"`
	Instances []EnvInstance `json:"Instances"`
	Paths     []EnvPath     `json:"Paths"`
	TotalSize int64         `json:"TotalSize"`
	Health    HealthLevel   `json:"Health"`
}

type DiskInfo struct {
	Letter      string `json:"Letter"`
	TotalBytes  int64  `json:"TotalBytes"`
	FreeBytes   int64  `json:"FreeBytes"`
	UsedBytes   int64  `json:"UsedBytes"`
	UsedPercent int    `json:"UsedPercent"`
}

type MigrationTask struct {
	ID             int64      `json:"Id"`
	Name           string     `json:"Name"`
	Status         string     `json:"Status"`
	SourcePaths    []string   `json:"SourcePaths"`
	TargetDir      string     `json:"TargetDir"`
	UseJunction    bool       `json:"UseJunction"`
	CreateSnapshot bool       `json:"CreateSnapshot"`
	SnapshotID     *int64     `json:"SnapshotId,omitempty"`
	StartedAt      *time.Time `json:"StartedAt,omitempty"`
	CompletedAt    *time.Time `json:"CompletedAt,omitempty"`
	ErrorMsg       string     `json:"ErrorMsg"`
}

type Snapshot struct {
	ID        int64     `json:"Id"`
	Name      string    `json:"Name"`
	DataJSON  string    `json:"DataJson"`
	CreatedAt time.Time `json:"CreatedAt"`
}

type HistoryEntry struct {
	ID           int64     `json:"Id"`
	Action       string    `json:"Action"`
	TargetEnv    string    `json:"TargetEnv"`
	DetailsJSON  string    `json:"DetailsJson"`
	Success      bool      `json:"Success"`
	ErrorMessage string    `json:"ErrorMessage"`
	CreatedAt    time.Time `json:"CreatedAt"`
}

type CleanableItem struct {
	Name        string `json:"Name"`
	Path        string `json:"Path"`
	Description string `json:"Description"`
	SizeBytes   int64  `json:"SizeBytes"`
	Selected    bool   `json:"Selected"`
	EnvKey      string `json:"EnvKey"`
	Category    string `json:"Category"`
	RiskLevel   string `json:"RiskLevel"`
}

type AppSettings struct {
	AutoScanOnStartup      bool     `json:"AutoScanOnStartup"`
	ConfirmBeforeMigration bool     `json:"ConfirmBeforeMigration"`
	Theme                  string   `json:"Theme"`
	CustomScanPaths        []string `json:"CustomScanPaths"`
}

type MetricSnapshot struct {
	ID         int64     `json:"Id"`
	MetricKey  string    `json:"MetricKey"`
	TargetKey  string    `json:"TargetKey"`
	ValueBytes int64     `json:"ValueBytes"`
	CapturedAt time.Time `json:"CapturedAt"`
}

type MigrationProgress struct {
	Step       string `json:"Step"`
	StepIndex  int    `json:"StepIndex"`
	TotalSteps int    `json:"TotalSteps"`
	Percent    int    `json:"Percent"`
	Message    string `json:"Message"`
	EnvKey     string `json:"EnvKey"`
}
