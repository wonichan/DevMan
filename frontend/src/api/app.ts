import {
  ScanAll as _ScanAll,
  GetEnvs as _GetEnvs,
  GetEnvSummary as _GetEnvSummary,
  ManageEnv as _ManageEnv,
  UnmanageEnv as _UnmanageEnv,
  GetDiskInfo as _GetDiskInfo,
  GetHistory as _GetHistory,
  Migrate as _Migrate,
  AnalyzeCleanable as _AnalyzeCleanable,
  CleanItems as _CleanItems,
  GetSettings as _GetSettings,
  SaveSettings as _SaveSettings,
  GetMetricSnapshots as _GetMetricSnapshots,
} from '../../wailsjs/go/main/App';
import { EventsOn as _EventsOn } from '../../wailsjs/runtime/runtime';

import type {
  EnvSummary,
  Env,
  DiskInfo,
  HistoryEntry,
  CleanableItem,
  AppSettings,
  MigrationResult,
  MetricSnapshot,
  ToolVersionCatalog,
  ToolVersionState,
  VersionInstallPlan,
  VersionManagerConflict,
  VersionOperationResult,
} from '../devman-types';

export function ScanAll(): Promise<EnvSummary[]> {
  return _ScanAll();
}

export function GetEnvs(): Promise<Env[]> {
  return _GetEnvs();
}

export function GetEnvSummary(key: string): Promise<EnvSummary | null> {
  return _GetEnvSummary(key);
}

export function ManageEnv(key: string): Promise<Env> {
  if (!hasWailsBridge()) return Promise.reject(new Error('Environment management API is not ready'));
  return _ManageEnv(key) as Promise<Env>;
}

export function UnmanageEnv(key: string): Promise<Env> {
  if (!hasWailsBridge()) return Promise.reject(new Error('Environment management API is not ready'));
  return _UnmanageEnv(key) as Promise<Env>;
}

export function GetDiskInfo(): Promise<DiskInfo[]> {
  return _GetDiskInfo();
}

export function GetHistory(limit: number): Promise<HistoryEntry[]> {
  return _GetHistory(limit);
}

export function Migrate(envID: number, targetDir: string, useJunction: boolean): Promise<MigrationResult> {
  return _Migrate(envID, targetDir, useJunction);
}

export function AnalyzeCleanable(): Promise<CleanableItem[]> {
  return _AnalyzeCleanable();
}

export function CleanItems(items: CleanableItem[]): Promise<number> {
  return _CleanItems(items);
}

interface WailsRuntime {
  EventsOn(eventName: string, callback: (...data: unknown[]) => void): () => void;
}

interface AppBridge {
  FetchOfficialVersions?: (toolKey: string) => Promise<ToolVersionCatalog>;
  InstallVersion?: (toolKey: string, version: string, targetDir: string) => Promise<VersionOperationResult>;
  ListToolVersions?: () => Promise<ToolVersionState[]>;
  PreviewVersionInstall?: (toolKey: string, version: string) => Promise<VersionInstallPlan>;
  DetectVersionManager?: (toolKey: string) => Promise<VersionManagerConflict | null>;
  SwitchVersion?: (toolKey: string, instanceId: number) => Promise<VersionOperationResult>;
  UninstallVersion?: (instanceId: number, forceDeleteExternal: boolean) => Promise<VersionOperationResult>;
}

declare global {
  interface Window {
    go?: any;
    runtime?: WailsRuntime;
  }
}

const defaultSettings: AppSettings = {
  AutoScanOnStartup: false,
  ConfirmBeforeMigration: true,
  Theme: 'dark',
  CustomScanPaths: [],
};

function hasWailsBridge(): boolean {
  return typeof window !== 'undefined' && Boolean(window.go);
}

function appBridge(): AppBridge | undefined {
  if (typeof window === 'undefined') return undefined;
  return (window.go as { main?: { App?: AppBridge } } | undefined)?.main?.App;
}

export function GetSettings(): Promise<AppSettings> {
  if (!hasWailsBridge()) return Promise.resolve(defaultSettings);
  return _GetSettings() as Promise<AppSettings>;
}

export function SaveSettings(settings: AppSettings): Promise<void> {
  if (!hasWailsBridge()) return Promise.reject(new Error('Settings API is not ready'));
  return _SaveSettings(settings);
}

export function GetMetricSnapshots(metricKey: string, targetKey: string, limit: number): Promise<MetricSnapshot[]> {
  if (!hasWailsBridge()) return Promise.resolve([]);
  return _GetMetricSnapshots(metricKey, targetKey, limit) as Promise<MetricSnapshot[]>;
}

export function ListToolVersions(): Promise<ToolVersionState[]> {
  return appBridge()?.ListToolVersions?.() ?? Promise.resolve([]);
}

export function FetchOfficialVersions(toolKey: string): Promise<ToolVersionCatalog> {
  const fetchOfficialVersions = appBridge()?.FetchOfficialVersions;
  if (!fetchOfficialVersions) return Promise.reject(new Error('Official version API is not ready'));
  return fetchOfficialVersions(toolKey);
}

export function PreviewVersionInstall(toolKey: string, version: string): Promise<VersionInstallPlan> {
  const preview = appBridge()?.PreviewVersionInstall;
  if (!preview) return Promise.reject(new Error('Version management API is not ready'));
  return preview(toolKey, version);
}

export function InstallVersion(toolKey: string, version: string, targetDir: string): Promise<VersionOperationResult> {
  const installVersion = appBridge()?.InstallVersion;
  if (!installVersion) return Promise.reject(new Error('Install API is not ready'));
  return installVersion(toolKey, version, targetDir);
}

export function DetectVersionManager(toolKey: string): Promise<VersionManagerConflict | null> {
  return appBridge()?.DetectVersionManager?.(toolKey) ?? Promise.resolve(null);
}

export function SwitchVersion(toolKey: string, instanceId: number): Promise<VersionOperationResult> {
  const switchVersion = appBridge()?.SwitchVersion;
  if (!switchVersion) return Promise.reject(new Error('Switch API is not ready'));
  return switchVersion(toolKey, instanceId);
}

export function UninstallVersion(instanceId: number, forceDeleteExternal: boolean): Promise<VersionOperationResult> {
  const uninstallVersion = appBridge()?.UninstallVersion;
  if (!uninstallVersion) return Promise.reject(new Error('Uninstall API is not ready'));
  return uninstallVersion(instanceId, forceDeleteExternal);
}

export function EventsOn(eventName: string, callback: (...data: unknown[]) => void): () => void {
  if (window.runtime?.EventsOn) {
    return _EventsOn(eventName, callback);
  }
  return () => {};
}
