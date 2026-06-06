package versionmanager

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

type ToolVersionCatalog struct {
	ToolKey   string             `json:"ToolKey"`
	Versions  []AvailableVersion `json:"Versions"`
	FetchedAt time.Time          `json:"FetchedAt"`
	SourceURL string             `json:"SourceUrl"`
}

type AvailableVersion struct {
	Version     string    `json:"Version"`
	Stable      bool      `json:"Stable"`
	ReleaseDate time.Time `json:"ReleaseDate"`
	Arch        string    `json:"Arch"`
	DownloadURL string    `json:"DownloadUrl"`
	Checksum    string    `json:"Checksum"`
}

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

type VersionInstallPlan struct {
	ToolKey            string            `json:"ToolKey"`
	Version            string            `json:"Version"`
	TargetDir          string            `json:"TargetDir"`
	DownloadURL        string            `json:"DownloadUrl"`
	ArchiveName        string            `json:"ArchiveName"`
	ExtractedDir       string            `json:"ExtractedDir"`
	WillOverwrite      bool              `json:"WillOverwrite"`
	ResolverReason     string            `json:"ResolverReason"`
	EnvironmentChanges map[string]string `json:"EnvironmentChanges"`
}

type VersionOperationResult struct {
	Success             bool              `json:"Success"`
	Message             string            `json:"Message"`
	ToolKey             string            `json:"ToolKey"`
	Version             string            `json:"Version"`
	AffectedPaths       []string          `json:"AffectedPaths"`
	AffectedEnvironment map[string]string `json:"AffectedEnvironment"`
	RollbackAvailable   bool              `json:"RollbackAvailable"`
	VerificationCommand string            `json:"VerificationCommand"`
	VerificationOutput  string            `json:"VerificationOutput"`
}

type VersionManagerConflict struct {
	ToolKey  string `json:"ToolKey"`
	Manager  string `json:"Manager"`
	Evidence string `json:"Evidence"`
	Detected bool   `json:"Detected"`
}

type ToolVersionState struct {
	ToolKey         string                  `json:"ToolKey"`
	Name            string                  `json:"Name"`
	LocalVersions   []ManagedVersion        `json:"LocalVersions"`
	CurrentDefault  *ManagedVersion         `json:"CurrentDefault,omitempty"`
	ActiveCommand   string                  `json:"ActiveCommand"`
	PathConflict    string                  `json:"PathConflict"`
	ManagerConflict *VersionManagerConflict `json:"ManagerConflict,omitempty"`
	LastInstallPlan *VersionInstallPlan     `json:"LastInstallPlan,omitempty"`
}
