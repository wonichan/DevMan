package versionstore

import "time"

type VersionSource string

const (
	SourceDevMan         VersionSource = "devman"
	SourceExternal       VersionSource = "external"
	SourceVersionManager VersionSource = "version_manager"
)

type DeletePolicy string

const (
	DeletePolicyDirect        DeletePolicy = "direct"
	DeletePolicyRemoveOnly    DeletePolicy = "remove_tracking"
	DeletePolicyForceRequired DeletePolicy = "force_required"
	DeletePolicyBlocked       DeletePolicy = "blocked"
)

type ManagedVersion struct {
	ID           int64         `json:"Id"`
	ToolKey      string        `json:"ToolKey"`
	Version      string        `json:"Version"`
	InstallPath  string        `json:"InstallPath"`
	BinPath      string        `json:"BinPath"`
	Source       VersionSource `json:"Source"`
	IsDefault    bool          `json:"IsDefault"`
	IsActive     bool          `json:"IsActive"`
	CanDelete    bool          `json:"CanDelete"`
	DeletePolicy DeletePolicy  `json:"DeletePolicy"`
	DetectedAt   time.Time     `json:"DetectedAt"`
}

type InstallStrategy struct {
	ToolKey   string    `json:"ToolKey"`
	RootDir   string    `json:"RootDir"`
	Reason    string    `json:"Reason"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}
