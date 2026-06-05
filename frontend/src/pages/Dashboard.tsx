import { useEffect, useState, useCallback } from 'react';
import { ScanAll, GetDiskInfo, GetEnvs, GetEnvSummary, GetSettings, GetMetricSnapshots } from '../api/app';
import type { EnvSummary, DiskInfo, MetricSnapshot } from '../devman-types';
import { PageHeader } from '../components/ui/PageHeader';
import { Button } from '../components/ui/Button';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { StatusBadge } from '../components/ui/StatusBadge';
import { ManagementBadge } from '../components/ui/ManagementBadge';
import { ProgressBar } from '../components/ui/ProgressBar';
import { EmptyState } from '../components/ui/EmptyState';
import { useToast } from '../hooks/useToast';
import { usePageActions } from '../hooks/usePageActions';
import { RefreshIcon, CheckIcon, WarningIcon, InfoIcon, DashboardIcon, TrendingUpIcon } from '../components/icons';
import { sortEnvSummariesByManagement } from '../utils/environment-management';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

interface TrendPoint {
  label: string;
  value: number;
}

function formatTrendLabel(capturedAt: string): string {
  const d = new Date(capturedAt);
  const mm = String(d.getMonth() + 1).padStart(2, '0');
  const dd = String(d.getDate()).padStart(2, '0');
  const hh = String(d.getHours()).padStart(2, '0');
  const min = String(d.getMinutes()).padStart(2, '0');
  return `${mm}-${dd} ${hh}:${min}`;
}

function buildTrendData(snapshots: MetricSnapshot[]): TrendPoint[] {
  const points = snapshots
    .map((s) => ({ label: formatTrendLabel(s.CapturedAt), value: s.ValueBytes || 0 }))
    .reverse();
  return points.slice(-10);
}

function TrendChart({ data }: { data: TrendPoint[] }) {
  if (data.length < 2) return null;
  const width = 280;
  const height = 80;
  const padding = 4;
  const maxValue = Math.max(...data.map((d) => d.value), 1);
  const minValue = Math.min(...data.map((d) => d.value), 0);
  const range = maxValue - minValue || 1;

  const points = data.map((d, i) => {
    const x = padding + (i / (data.length - 1)) * (width - padding * 2);
    const y = height - padding - ((d.value - minValue) / range) * (height - padding * 2);
    return { x, y };
  });

  const pathD = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ');
  const areaD = `${pathD} L ${points[points.length - 1].x} ${height} L ${points[0].x} ${height} Z`;

  return (
    <svg width="100%" viewBox={`0 0 ${width} ${height}`} className="overflow-visible">
      <defs>
        <linearGradient id="trendGradient" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="rgba(34,197,94,0.25)" />
          <stop offset="100%" stopColor="rgba(34,197,94,0)" />
        </linearGradient>
      </defs>
      <path d={areaD} fill="url(#trendGradient)" />
      <path d={pathD} fill="none" stroke="#22c55e" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      {points.map((p, i) => (
        <circle key={i} cx={p.x} cy={p.y} r="3" fill="#0f172a" stroke="#22c55e" strokeWidth="2" />
      ))}
    </svg>
  );
}

export default function Dashboard() {
  const [summaries, setSummaries] = useState<EnvSummary[]>([]);
  const [disks, setDisks] = useState<DiskInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [trendData, setTrendData] = useState<TrendPoint[]>([]);
  const [trendLoading, setTrendLoading] = useState(true);
  const { success, error } = useToast();

  const loadStoredSummaries = async (): Promise<EnvSummary[]> => {
    const envs = await GetEnvs();
    const summaries = await Promise.all((envs || []).map((env) => GetEnvSummary(env.Key)));
    return sortEnvSummariesByManagement(summaries.filter((summary): summary is EnvSummary => Boolean(summary)));
  };

  const load = async (refresh = true) => {
    setLoading(true);
    try {
      const s = refresh ? await ScanAll() : await loadStoredSummaries();
      setSummaries(sortEnvSummariesByManagement(s || []));
      const d = await GetDiskInfo();
      setDisks(d || []);
      if (refresh) {
        success('刷新成功', '环境数据已更新');
        await loadTrend();
      }
    } catch (e: unknown) {
      error('加载失败', e instanceof Error ? e.message : String(e));
    }
    setLoading(false);
  };

  const loadTrend = useCallback(async () => {
    setTrendLoading(true);
    try {
      const snapshots = await GetMetricSnapshots('env_total_size', 'all', 100);
      setTrendData(buildTrendData(snapshots || []));
    } catch {
      setTrendData([]);
    }
    setTrendLoading(false);
  }, []);

  useEffect(() => {
    const initLoad = async () => {
      setLoading(true);
      try {
        const settings = await GetSettings();
        const s = settings.AutoScanOnStartup ? await ScanAll() : await loadStoredSummaries();
        setSummaries(sortEnvSummariesByManagement(s || []));
        const d = await GetDiskInfo();
        setDisks(d || []);
      } catch (e: unknown) {
        error('加载失败', e instanceof Error ? e.message : String(e));
      }
      setLoading(false);
    };
    initLoad();
    loadTrend();
  }, []);

  usePageActions('dashboard', { refresh: () => load(true) });

  const totalEnvSize = summaries.reduce((sum, s) => sum + (s.TotalSize || 0), 0);
  const managedCount = summaries.filter((summary) => summary.Env.IsManaged).length;
  const cDisk = disks.find(d => d.Letter === 'C:') || disks[0];
  const cUsedPercent = cDisk ? Math.round(((cDisk.TotalBytes - cDisk.FreeBytes) / cDisk.TotalBytes) * 100) : 0;

  const diskStatus = cUsedPercent > 90 ? 'danger' : cUsedPercent > 70 ? 'warning' : 'healthy';
  const diskStatusLabel = cUsedPercent > 90 ? '危险' : cUsedPercent > 70 ? '建议清理' : '良好';

  const latestTrendValue = trendData.length > 0 ? trendData[trendData.length - 1].value : 0;
  const previousTrendValue = trendData.length > 1 ? trendData[trendData.length - 2].value : latestTrendValue;
  const trendDelta = latestTrendValue - previousTrendValue;

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

      <div className="grid grid-cols-1 md:grid-cols-4 gap-5 mb-6">
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
            <div className="w-10 h-10 rounded-xl bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20">
              <CheckIcon className="w-5 h-5 text-emerald-400" />
            </div>
            <div>
              <p className="text-xs text-slate-400">Managed</p>
              <p className="text-2xl font-bold text-slate-100">{managedCount}</p>
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

      <h2 className="text-lg font-bold text-slate-200 mb-4">趋势</h2>
      <SurfaceCard variant="raised" className="mb-6">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 rounded-xl bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20">
            <TrendingUpIcon className="w-5 h-5 text-emerald-400" />
          </div>
          <div>
            <p className="text-xs text-slate-400">环境总占用趋势</p>
            <div className="flex items-center gap-2">
              <p className="text-2xl font-bold text-slate-100">{formatBytes(latestTrendValue)}</p>
              {trendData.length > 1 && (
                <span className={`text-xs font-medium ${trendDelta >= 0 ? 'text-red-400' : 'text-emerald-400'}`}>
                  {trendDelta >= 0 ? '+' : ''}{formatBytes(trendDelta)}
                </span>
              )}
            </div>
          </div>
        </div>
        {trendLoading ? (
          <div className="h-20 flex items-center justify-center">
            <div className="w-6 h-6 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
          </div>
        ) : trendData.length < 2 ? (
          <EmptyState
            icon={<TrendingUpIcon className="w-6 h-6" />}
            title="暂无趋势数据"
            description="执行扫描后将自动记录环境占用趋势"
          />
        ) : (
          <div className="w-full">
            <TrendChart data={trendData} />
            <div className="flex justify-between mt-2">
              <span className="text-[10px] text-slate-500">{trendData[0].label}</span>
              <span className="text-[10px] text-slate-500">{trendData[trendData.length - 1].label}</span>
            </div>
          </div>
        )}
      </SurfaceCard>

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
                <div className="flex flex-col items-end gap-2">
                  <StatusBadge status={summary.Health === 'healthy' ? 'healthy' : 'warning'} label={summary.Health === 'healthy' ? '正常' : '需关注'} />
                  <ManagementBadge managed={summary.Env.IsManaged} />
                </div>
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
