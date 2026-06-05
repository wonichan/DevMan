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

declare global {
  interface Window {
    go?: unknown;
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

export function EventsOn(eventName: string, callback: (...data: unknown[]) => void): () => void {
  if (window.runtime?.EventsOn) {
    return _EventsOn(eventName, callback);
  }
  return () => {};
}
