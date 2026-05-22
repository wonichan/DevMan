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
