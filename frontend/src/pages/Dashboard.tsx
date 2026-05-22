import { useEffect, useState } from 'react';
import { ScanAll, GetDiskInfo } from '../bindings/go/main/App';
import Panel from '../components/Panel';
import type { EnvSummary, DiskInfo } from '../devman-types';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

export default function Dashboard() {
  const [summaries, setSummaries] = useState<EnvSummary[]>([]);
  const [disks, setDisks] = useState<DiskInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const load = async () => {
    setLoading(true);
    try {
      const s = await ScanAll();
      setSummaries(s || []);
      const d = await GetDiskInfo();
      setDisks(d || []);
    } catch (e) {
      console.error(e);
    }
    setLoading(false);
  };

  useEffect(() => {
    load();
  }, []);

  const totalEnvSize = summaries.reduce((sum, s) => sum + (s.TotalSize || 0), 0);
  const cDisk = disks.find(d => d.Letter === 'C:') || disks[0];
  const cUsedPercent = cDisk ? Math.round(((cDisk.TotalBytes - cDisk.FreeBytes) / cDisk.TotalBytes) * 100) : 0;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-devman-text-primary">📊 Dashboard</h1>
          <p className="text-sm text-devman-text-muted mt-1">总览你的开发环境状态</p>
        </div>
        <button
          onClick={load}
          disabled={loading}
          className="px-4 py-2 bg-devman-panel-raised border border-devman-border rounded-xl text-sm text-devman-text-primary hover:bg-devman-border/20 transition-colors disabled:opacity-50"
        >
          {loading ? '扫描中...' : '🔄 刷新环境数据'}
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5 mb-6">
        {/* C盘健康 */}
        <Panel className="p-6">
          <div className="flex items-start justify-between mb-4">
            <span className="text-sm text-devman-text-muted">💾 C盘健康状态</span>
            <span className={`text-xs px-2 py-1 rounded-full font-medium ${
              cUsedPercent > 90 ? 'bg-red-500/10 text-red-400' :
              cUsedPercent > 70 ? 'bg-yellow-500/10 text-yellow-400' :
              'bg-green-500/10 text-green-400'
            }`}>
              {cUsedPercent > 90 ? '🔴 危险' : cUsedPercent > 70 ? '🟡 建议清理' : '🟢 良好'}
            </span>
          </div>
          <div className="text-5xl font-bold text-devman-text-primary mb-2">{cUsedPercent}%</div>
          <p className="text-sm text-devman-text-muted">
            {cDisk ? formatBytes(cDisk.TotalBytes - cDisk.FreeBytes) + ' / ' + formatBytes(cDisk.TotalBytes) : '--'}
          </p>
        </Panel>

        {/* 已管理环境 */}
        <Panel className="p-6">
          <div className="flex items-start justify-between mb-4">
            <span className="text-sm text-devman-text-muted">🔧 已管理环境</span>
            <span className="text-xs px-2 py-1 rounded-full bg-green-500/10 text-green-400 font-medium">
              ✓ 正常
            </span>
          </div>
          <div className="text-5xl font-bold text-devman-accent mb-2">{summaries.length}</div>
          <p className="text-sm text-devman-text-muted">
            总占用 {formatBytes(totalEnvSize)}
          </p>
        </Panel>
      </div>

      {/* 空间占用排行 */}
      <Panel className="p-6 mb-6">
        <h3 className="text-sm text-devman-text-muted mb-4">📊 环境空间占用排行</h3>
        <div className="space-y-3">
          {[...summaries]
            .sort((a, b) => (b.TotalSize || 0) - (a.TotalSize || 0))
            .map((s) => {
              const maxSize = Math.max(...summaries.map(x => x.TotalSize || 0), 1);
              const pct = ((s.TotalSize || 0) / maxSize) * 100;
              return (
                <div key={s.Env.Key} className="flex items-center gap-4">
                  <span className="w-20 text-sm text-devman-text-primary font-medium">{s.Env.Name}</span>
                  <div className="flex-1 h-3 bg-devman-panel-raised rounded-full overflow-hidden">
                    <div
                      className="h-full rounded-full bg-gradient-to-r from-devman-accent to-devman-info transition-all"
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                  <span className="w-20 text-right text-sm text-devman-text-muted">{formatBytes(s.TotalSize || 0)}</span>
                </div>
              );
            })}
        </div>
      </Panel>
    </div>
  );
}
