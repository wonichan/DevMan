import { useEffect, useState } from 'react';
import { ScanAll, GetDiskInfo, GetEnvs, GetEnvSummary, GetSettings } from '../api/app';
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

  const loadStoredSummaries = async (): Promise<EnvSummary[]> => {
    const envs = await GetEnvs();
    const summaries = await Promise.all((envs || []).map((env) => GetEnvSummary(env.Key)));
    return summaries.filter((summary): summary is EnvSummary => Boolean(summary));
  };

  const load = async (refresh = true) => {
    setLoading(true);
    try {
      const s = refresh ? await ScanAll() : await loadStoredSummaries();
      setSummaries(s || []);
      const d = await GetDiskInfo();
      setDisks(d || []);
      if (refresh) success('刷新成功', '环境数据已更新');
    } catch (e: unknown) {
      error('加载失败', e instanceof Error ? e.message : String(e));
    }
    setLoading(false);
  };

  useEffect(() => {
    const initLoad = async () => {
      setLoading(true);
      try {
        const settings = await GetSettings();
        const s = settings.AutoScanOnStartup ? await ScanAll() : await loadStoredSummaries();
        setSummaries(s || []);
        const d = await GetDiskInfo();
        setDisks(d || []);
      } catch (e: unknown) {
        error('加载失败', e instanceof Error ? e.message : String(e));
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
          <Button onClick={() => load(true)} disabled={loading} isLoading={loading}>
            <RefreshIcon className="mr-2 h-4 w-4" />
            刷新
          </Button>
        }
      />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-5 mb-6">
        <SurfaceCard variant="raised">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 rounded-xl bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20">
              <CheckIcon className="w-5 h-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-xs text-slate-400">已检测环境</p>
              <p className="text-2xl font-bold text-slate-100">{summaries.length}</p>
            </div>
          </div>
        </SurfaceCard>
        <SurfaceCard variant="raised">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 rounded-xl bg-cyan-500/10 flex items-center justify-center border border-cyan-500/20">
              <InfoIcon className="w-5 h-5 text-cyan-400" />
            </div>
            <div>
              <p className="text-xs text-slate-400">环境总占用</p>
              <p className="text-2xl font-bold text-slate-100">{formatBytes(totalEnvSize)}</p>
            </div>
          </div>
        </SurfaceCard>
        <SurfaceCard variant="raised">
          <div className="flex items-center gap-3 mb-3">
            <div className={`w-10 h-10 rounded-xl flex items-center justify-center border ${
              diskStatus === 'danger' ? 'bg-red-500/10 border-red-500/20' : 
              diskStatus === 'warning' ? 'bg-amber-500/10 border-amber-500/20' : 
              'bg-emerald-500/10 border-emerald-500/20'
            }`}>
              <WarningIcon className={`w-5 h-5 ${
                diskStatus === 'danger' ? 'text-red-400' : 
                diskStatus === 'warning' ? 'text-amber-400' : 
                'text-emerald-400'
              }`} />
            </div>
            <div>
              <p className="text-xs text-slate-400">C 盘状态</p>
              <div className="flex items-center gap-2">
                <p className="text-2xl font-bold text-slate-100">{cUsedPercent}%</p>
                <StatusBadge status={diskStatus} label={diskStatusLabel} />
              </div>
            </div>
          </div>
          {cDisk && (
            <ProgressBar value={cUsedPercent} variant={diskStatus === 'danger' ? 'danger' : diskStatus === 'warning' ? 'warning' : 'accent'} />
          )}
        </SurfaceCard>
      </div>

      <h2 className="text-lg font-bold text-slate-200 mb-4">环境概览</h2>
      {summaries.length === 0 ? (
        <EmptyState 
          icon={<DashboardIcon className="w-6 h-6" />}
          title="暂无环境数据"
          description="点击上方「刷新」按钮扫描系统环境"
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {summaries.map((summary) => (
            <SurfaceCard key={summary.Env.Key} variant="interactive">
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <span className="text-2xl">{summary.Env.Icon}</span>
                  <div>
                    <h3 className="text-base font-bold text-slate-200">{summary.Env.Name}</h3>
                    <p className="text-xs text-slate-400">{summary.Instances.length} 个版本</p>
                  </div>
                </div>
                <StatusBadge status={summary.Health === 'healthy' ? 'healthy' : 'warning'} label={summary.Health === 'healthy' ? '正常' : '需关注'} />
              </div>
              <div className="space-y-2">
                {summary.Instances.slice(0, 2).map(inst => (
                  <div key={inst.Id} className="flex items-center justify-between text-sm">
                    <span className="text-slate-300 font-mono">{inst.Version}</span>
                    <span className="text-slate-500 text-xs">{inst.InstallPath}</span>
                  </div>
                ))}
              </div>
              <div className="mt-4 pt-3 border-t border-[#334155] flex items-center justify-between">
                <span className="text-xs text-slate-500">总大小</span>
                <span className="text-sm font-bold text-cyan-400">{formatBytes(summary.TotalSize)}</span>
              </div>
            </SurfaceCard>
          ))}
        </div>
      )}
    </div>
  );
}
