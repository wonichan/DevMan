import {
  ScanAll as _ScanAll,
  GetEnvs as _GetEnvs,
  GetEnvSummary as _GetEnvSummary,
  GetDiskInfo as _GetDiskInfo,
  GetHistory as _GetHistory,
  Migrate as _Migrate,
  AnalyzeCleanable as _AnalyzeCleanable,
  CleanItems as _CleanItems,
} from '../bindings/go/main/App';

import type {
  EnvSummary,
  Env,
  DiskInfo,
  HistoryEntry,
  CleanableItem,
  AppSettings,
  MigrationResult,
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

interface WailsGoApp {
  ScanAll(): Promise<EnvSummary[]>;
  GetEnvs(): Promise<Env[]>;
  GetEnvSummary(key: string): Promise<EnvSummary | null>;
  GetDiskInfo(): Promise<DiskInfo[]>;
  GetHistory(limit: number): Promise<HistoryEntry[]>;
  Migrate(envID: number, targetDir: string, useJunction: boolean): Promise<MigrationResult>;
  AnalyzeCleanable(): Promise<CleanableItem[]>;
  CleanItems(items: CleanableItem[]): Promise<number>;
  GetSettings?(): Promise<AppSettings>;
  SaveSettings?(settings: AppSettings): Promise<void>;
}

interface WailsGoMain {
  App?: WailsGoApp;
}

interface WailsGo {
  main?: WailsGoMain;
}

interface WailsRuntime {
  EventsOn(eventName: string, callback: (...data: unknown[]) => void): () => void;
}

declare global {
  interface Window {
    go?: WailsGo;
    runtime?: WailsRuntime;
  }
}

export function GetSettings(): Promise<AppSettings> {
  const fn = window.go?.main?.App?.GetSettings;
  if (!fn) {
    return Promise.resolve({
      AutoScanOnStartup: true,
      ConfirmBeforeMigration: true,
      Theme: 'dark',
      CustomScanPaths: [],
    });
  }
  return fn();
}

export function SaveSettings(settings: AppSettings): Promise<void> {
  const fn = window.go?.main?.App?.SaveSettings;
  if (!fn) {
    return Promise.reject(new Error('设置 API 尚未就绪'));
  }
  return fn(settings);
}

export function EventsOn(eventName: string, callback: (...data: unknown[]) => void): () => void {
  if (window.runtime?.EventsOn) {
    return window.runtime.EventsOn(eventName, callback);
  }
  return () => {};
}
