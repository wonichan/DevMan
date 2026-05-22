// @ts-ignore
const goMainApp = window.go?.main?.App;

export function ScanAll(): Promise<any[]> {
  return goMainApp?.ScanAll() ?? Promise.resolve([]);
}

export function GetEnvs(): Promise<any[]> {
  return goMainApp?.GetEnvs() ?? Promise.resolve([]);
}

export function GetEnvSummary(key: string): Promise<any> {
  return goMainApp?.GetEnvSummary(key) ?? Promise.resolve(null);
}

export function GetDiskInfo(): Promise<any[]> {
  return goMainApp?.GetDiskInfo() ?? Promise.resolve([]);
}

export function GetHistory(limit: number): Promise<any[]> {
  return goMainApp?.GetHistory(limit) ?? Promise.resolve([]);
}

export function Migrate(envID: number, targetDir: string, useJunction: boolean): Promise<any> {
  return goMainApp?.Migrate(envID, targetDir, useJunction) ?? Promise.resolve({ success: false, message: 'not ready' });
}

export function AnalyzeCleanable(): Promise<any[]> {
  return goMainApp?.AnalyzeCleanable() ?? Promise.resolve([]);
}

export function CleanItems(items: any[]): Promise<number> {
  return goMainApp?.CleanItems(items) ?? Promise.resolve(0);
}
