export interface Env {
  Id: number;
  Name: string;
  Key: string;
  Category: string;
  Icon: string;
  Description: string;
  Website: string;
  IsManaged: boolean;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface EnvInstance {
  Id: number;
  EnvId: number;
  Version: string;
  InstallPath: string;
  IsDefault: boolean;
  IsActive: boolean;
  Source: string;
  DetectedAt: string;
}

export interface EnvPath {
  Id: number;
  EnvId: number;
  InstanceId?: number;
  Type: string;
  Path: string;
  SizeBytes: number;
  IsMovable: boolean;
  LastSized: string;
}

export interface EnvSummary {
  Env: Env;
  Instances: EnvInstance[];
  Paths: EnvPath[];
  TotalSize: number;
  Health: string;
}

export interface DiskInfo {
  Letter: string;
  TotalBytes: number;
  FreeBytes: number;
  UsedBytes: number;
  UsedPercent: number;
}

export interface HistoryEntry {
  Id: number;
  Action: string;
  TargetEnv: string;
  DetailsJson: string;
  Success: boolean;
  ErrorMessage: string;
  CreatedAt: string;
}

export interface CleanableItem {
  Name: string;
  Path: string;
  Description: string;
  SizeBytes: number;
  Selected: boolean;
  EnvKey: string;
  Category: string;
  RiskLevel: string;
}

export interface AppSettings {
  AutoScanOnStartup: boolean;
  ConfirmBeforeMigration: boolean;
  Theme: 'dark' | 'system';
  CustomScanPaths: string[];
}

export interface MigrationProgress {
  Step: string;
  StepIndex: number;
  TotalSteps: number;
  Percent: number;
  Message: string;
  EnvKey: string;
}

export interface MigrationResult {
  success: boolean;
  message: string;
  bytesMoved?: number;
  durationMs?: number;
}

export interface MetricSnapshot {
  Id: number;
  MetricKey: string;
  TargetKey: string;
  ValueBytes: number;
  CapturedAt: string;
}

export type VersionSource = 'devman' | 'external' | 'version_manager';

export interface AvailableVersion {
  Version: string;
  Stable: boolean;
  ReleaseDate: string;
  Arch: string;
  DownloadUrl: string;
  Checksum: string;
}

export interface ToolVersionCatalog {
  ToolKey: string;
  Versions: AvailableVersion[];
  FetchedAt: string;
  SourceUrl: string;
}

export interface ManagedVersion {
  Id: number;
  ToolKey: string;
  Version: string;
  InstallPath: string;
  BinPath: string;
  Source: VersionSource;
  IsDefault: boolean;
  IsActive: boolean;
  CanDelete: boolean;
  DeletePolicy: string;
  DetectedAt: string;
}

export interface VersionManagerConflict {
  ToolKey: string;
  Manager: string;
  Evidence: string;
  Detected: boolean;
}

export interface VersionInstallPlan {
  ToolKey: string;
  Version: string;
  TargetDir: string;
  DownloadUrl: string;
  ArchiveName: string;
  ExtractedDir: string;
  WillOverwrite: boolean;
  ResolverReason: string;
  EnvironmentChanges: Record<string, string>;
}

export interface VersionOperationResult {
  Success: boolean;
  Message: string;
  ToolKey: string;
  Version: string;
  AffectedPaths: string[];
  AffectedEnvironment: Record<string, string>;
  RollbackAvailable: boolean;
  VerificationCommand: string;
  VerificationOutput: string;
}

export interface ToolVersionState {
  ToolKey: string;
  Name: string;
  LocalVersions: ManagedVersion[];
  CurrentDefault?: ManagedVersion;
  ActiveCommand: string;
  PathConflict: string;
  ManagerConflict?: VersionManagerConflict;
  LastInstallPlan?: VersionInstallPlan;
}
