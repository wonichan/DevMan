import { useEffect, useState } from 'react';
import { ScanAll, GetDiskInfo } from '../bindings/go/main/App';
import type { EnvSummary, DiskInfo } from '../devman-types';
import { PageHeader } from '../components/ui/PageHeader';
import { Button } from '../components/ui/Button';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { StatusBadge } from '../components/ui/StatusBadge';
import { ProgressBar } from '../components/ui/ProgressBar';
import { EmptyState } from '../components/ui/EmptyState';
import { useToast } from '../hooks/useToast';
import { RefreshIcon, CheckIcon, WarningIcon, InfoIcon, DashboardIcon } from '../components/icons';

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
  const [loading, setLoading] = useState(true);
  const { success, error } = useToast();

  const load = async () => {
    setLoading(true);
    try {
      const s = await ScanAll();
      setSummaries(s || []);
      const d = await GetDiskInfo();
      setDisks(d || []);
      success('刷新成功', '环境数据已更新');
    } catch (e: unknown) {
      console.error(e);
      error('加载失败', e instanceof Error ? e.message : String(e));
    }
    setLoading(false);
  };

  useEffect(() => {
    // Initial load without toast to prevent spam
    const initLoad = async () => {
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
    initLoad();
  }, []);

  const totalEnvSize = summaries.reduce((sum, s) => sum + (s.TotalSize || 0), 0);
  const cDisk = disks.find(d => d.Letter === 'C:') || disks[0];
  const cUsedPercent = cDisk ? Math.round(((cDisk.TotalBytes - cDisk.FreeBytes) / cDisk.TotalBytes) * 100) : 0;

  const diskStatus = cUsedPercent > 90 ? 'danger' : cUsedPercent > 70 ? 'warning' : 'healthy';
  const diskStatusLabel = cUsedPercent > 90 ? '危险' : cUsedPercent > 70 ? '建议清理' : '良好';

  return (
    <div className="animate-in fade-in duration-300">
      <PageHeader 
        title="Dashboard" 
        description="总览你的开发环境状态"
        actions={
          <Button onClick={load} disabled={loading} isLoading={loading} className="flex items-center gap-2">
            <RefreshIcon className="w-4 h-4" />
            刷新环境数据
          </Button>
        }
      />

      {loading && summaries.length === 0 && disks.length === 0 ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-5 mb-6">
          <SurfaceCard className="p-6 h-32 animate-pulse" />
          <SurfaceCard className="p-6 h-32 animate-pulse" />
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-5 mb-6">
          {/* C盘健康 */}
          <SurfaceCard className="p-6">
            <div className="flex items-start justify-between mb-4">
              <span className="text-sm text-slate-400 flex items-center gap-2">
                <InfoIcon className="w-4 h-4" />
                C盘健康状态
              </span>
              <StatusBadge status={diskStatus}>
                {diskStatusLabel}
              </StatusBadge>
            </div>
            <div className="text-5xl font-bold text-slate-100 mb-2">{cUsedPercent}%</div>
            <p className="text-sm text-slate-400">
              {cDisk ? formatBytes(cDisk.TotalBytes - cDisk.FreeBytes) + ' / ' + formatBytes(cDisk.TotalBytes) : '--'}
            </p>
          </SurfaceCard>

          {/* 已管理环境 */}
          <SurfaceCard className="p-6">
            <div className="flex items-start justify-between mb-4">
              <span className="text-sm text-slate-400 flex items-center gap-2">
                <DashboardIcon className="w-4 h-4" />
                已管理环境
              </span>
              <StatusBadge status="healthy">
                正常
              </StatusBadge>
            </div>
            <div className="text-5xl font-bold text-emerald-400 mb-2">{summaries.length}</div>
            <p className="text-sm text-slate-400">
              总占用 {formatBytes(totalEnvSize)}
            </p>
          </SurfaceCard>
        </div>
      )}

      {/* 空间占用排行 */}
      <SurfaceCard className="p-6 mb-6">
        <h3 className="text-sm text-slate-400 mb-4 flex items-center gap-2">
          <InfoIcon className="w-4 h-4" />
          环境空间占用排行
        </h3>
        
        {loading && summaries.length === 0 ? (
          <div className="space-y-4">
            {[1, 2, 3].map(i => (
              <div key={i} className="flex items-center gap-4 animate-pulse">
                <div className="w-20 h-5 bg-slate-700 rounded" />
                <div className="flex-1 h-3 bg-slate-700 rounded-full" />
                <div className="w-20 h-5 bg-slate-700 rounded" />
              </div>
            ))}
          </div>
        ) : summaries.length > 0 ? (
          <div className="space-y-4">
            {[...summaries]
              .sort((a, b) => (b.TotalSize || 0) - (a.TotalSize || 0))
              .map((s) => {
                const maxSize = Math.max(...summaries.map(x => x.TotalSize || 0), 1);
                const pct = ((s.TotalSize || 0) / maxSize) * 100;
                return (
                  <div key={s.Env.Key} className="flex items-center gap-4">
                    <span className="w-24 truncate text-sm text-slate-200 font-medium" title={s.Env.Name}>{s.Env.Name}</span>
                    <div className="flex-1">
                      <ProgressBar value={pct} variant="accent" />
                    </div>
                    <span className="w-20 text-right text-sm text-slate-400">{formatBytes(s.TotalSize || 0)}</span>
                  </div>
                );
              })}
          </div>
        ) : (
          <EmptyState 
            icon={<DashboardIcon />}
            title="暂无环境数据"
            description="点击右上角的刷新按钮开始扫描"
          />
        )}
      </SurfaceCard>
    </div>
  );
}
