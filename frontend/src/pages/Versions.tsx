import { useEffect, useMemo, useState } from 'react';
import { ListToolVersions, PreviewVersionInstall } from '../api/app';
import { Button } from '../components/ui/Button';
import { EmptyState } from '../components/ui/EmptyState';
import { PageHeader } from '../components/ui/PageHeader';
import { StatusBadge } from '../components/ui/StatusBadge';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { RefreshIcon, SearchIcon, VersionsIcon, WarningIcon } from '../components/icons';
import { useToast } from '../hooks/useToast';
import type { ToolVersionState, VersionInstallPlan } from '../devman-types';
import { conflictLabel, versionSourceLabel } from '../utils/version-management';

const preferredToolOrder = ['go', 'node', 'bun', 'flutter'];

function sortTools(tools: ToolVersionState[]) {
  return [...tools].sort((a, b) => {
    const aIndex = preferredToolOrder.indexOf(a.ToolKey);
    const bIndex = preferredToolOrder.indexOf(b.ToolKey);
    if (aIndex !== -1 && bIndex !== -1) return aIndex - bIndex;
    if (aIndex !== -1) return -1;
    if (bIndex !== -1) return 1;
    return a.Name.localeCompare(b.Name);
  });
}

function formatToolName(tool: ToolVersionState) {
  if (tool.ToolKey === 'node') return 'Node.js';
  if (tool.ToolKey === 'go') return 'Go';
  if (tool.ToolKey === 'bun') return 'Bun';
  if (tool.ToolKey === 'flutter') return 'Flutter';
  return tool.Name || tool.ToolKey;
}

export default function Versions() {
  const [tools, setTools] = useState<ToolVersionState[]>([]);
  const [activeToolKey, setActiveToolKey] = useState('go');
  const [loading, setLoading] = useState(false);
  const [previewVersion, setPreviewVersion] = useState('');
  const [installPlan, setInstallPlan] = useState<VersionInstallPlan | null>(null);
  const { error, success } = useToast();

  const orderedTools = useMemo(() => sortTools(tools), [tools]);
  const activeTool = orderedTools.find((tool) => tool.ToolKey === activeToolKey);
  const activeConflict = conflictLabel(activeTool?.ManagerConflict);
  const hasManagerConflict = Boolean(activeConflict);

  const loadData = async () => {
    setLoading(true);
    try {
      const data = await ListToolVersions();
      const list = data || [];
      setTools(list);
      setActiveToolKey((currentKey) => {
        if (list.some((tool) => tool.ToolKey === currentKey)) return currentKey;
        return sortTools(list)[0]?.ToolKey || currentKey;
      });
    } catch (e: unknown) {
      error('加载版本数据失败', e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleSelectTool = (toolKey: string) => {
    setActiveToolKey(toolKey);
    setInstallPlan(null);
  };

  const handlePreviewInstall = async () => {
    const version = previewVersion.trim();
    if (!version) return;

    setInstallPlan(null);
    try {
      const plan = await PreviewVersionInstall(activeToolKey, version);
      setInstallPlan(plan);
      success('安装预览已生成', plan.ResolverReason || '已解析目标安装路径');
    } catch (e: unknown) {
      setInstallPlan(null);
      error('生成安装预览失败', e instanceof Error ? e.message : String(e));
    }
  };

  return (
    <div>
      <PageHeader
        title="版本管理"
        description="查询本地工具版本，预览安装计划，并为后续安装、切换、删除本地版本提供入口。"
        actions={
          <Button variant="secondary" onClick={loadData} isLoading={loading}>
            <RefreshIcon className="mr-2 h-4 w-4" />
            刷新
          </Button>
        }
      />

      {orderedTools.length > 0 && (
        <div className="mb-6 flex flex-wrap gap-3">
          {orderedTools.map((tool) => {
            const isActive = tool.ToolKey === activeToolKey;
            return (
              <button
                key={tool.ToolKey}
                type="button"
                onClick={() => handleSelectTool(tool.ToolKey)}
                className={`rounded-xl border px-4 py-2 text-sm font-medium transition-all ${
                  isActive
                    ? 'border-emerald-500/50 bg-emerald-500/10 text-emerald-300 shadow-[0_0_12px_rgba(16,185,129,0.12)]'
                    : 'border-[#334155] bg-[#1e293b]/70 text-slate-300 hover:border-[#475569] hover:bg-[#1e293b]'
                }`}
              >
                {formatToolName(tool)}
              </button>
            );
          })}
        </div>
      )}

      {orderedTools.length === 0 && !loading && (
        <EmptyState
          icon={<SearchIcon className="h-6 w-6" />}
          title="暂无版本数据"
          description="未发现可管理的 Go、Node.js、Bun 或 Flutter 本地版本。"
        />
      )}

      {activeTool && (
        <div className="grid grid-cols-1 gap-5 xl:grid-cols-[minmax(0,1fr)_380px]">
          <div className="space-y-5">
            <SurfaceCard>
              <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-3">
                    <VersionsIcon className="h-6 w-6 text-emerald-400" />
                    <h2 className="text-lg font-bold text-slate-100">{formatToolName(activeTool)}</h2>
                    {loading && <StatusBadge status="info" label="加载中" />}
                    {hasManagerConflict ? (
                      <StatusBadge status="warning" label="存在管理器冲突" />
                    ) : (
                      <StatusBadge status="healthy" label="可预览安装" />
                    )}
                  </div>
                  <p className="mt-2 text-sm text-slate-400">
                    当前默认版本：
                    <span className="font-mono text-slate-200">
                      {activeTool.CurrentDefault?.Version || '未设置'}
                    </span>
                  </p>
                  {activeTool.ActiveCommand && (
                    <p className="mt-1 truncate text-xs text-slate-500">
                      活跃命令：<span className="font-mono">{activeTool.ActiveCommand}</span>
                    </p>
                  )}
                  {activeTool.PathConflict && (
                    <p className="mt-1 truncate text-xs text-amber-300">
                      PATH 提示：<span className="font-mono">{activeTool.PathConflict}</span>
                    </p>
                  )}
                </div>
                <div className="rounded-lg border border-[#334155] bg-[#0f172a] px-4 py-3 text-right">
                  <p className="text-[10px] uppercase text-slate-500">Local Versions</p>
                  <p className="text-2xl font-bold text-slate-100">{activeTool.LocalVersions.length}</p>
                </div>
              </div>

              {hasManagerConflict && (
                <div className="mt-4 flex gap-3 rounded-lg border border-amber-500/20 bg-amber-500/10 p-3 text-sm text-amber-200">
                  <WarningIcon className="mt-0.5 h-4 w-4 shrink-0" />
                  <p>{activeConflict}</p>
                </div>
              )}
            </SurfaceCard>

            <SurfaceCard>
              <div className="mb-4 flex items-center justify-between gap-4">
                <div>
                  <h3 className="text-sm font-bold text-slate-200">本地版本</h3>
                  <p className="mt-1 text-xs text-slate-400">显示已检测到的安装路径和来源。</p>
                </div>
              </div>

              <div className="space-y-2">
                {activeTool.LocalVersions.map((version) => (
                  <div
                    key={`${version.ToolKey}-${version.Version}-${version.InstallPath}`}
                    className={`rounded-lg border px-4 py-3 ${
                      version.IsDefault
                        ? 'border-emerald-500/30 bg-emerald-500/10'
                        : 'border-[#334155] bg-[#0f172a]'
                    }`}
                  >
                    <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                      <div className="min-w-0">
                        <p className="font-mono text-sm font-bold text-slate-100">{version.Version}</p>
                        <p className="mt-1 truncate font-mono text-xs text-slate-400">{version.InstallPath}</p>
                      </div>
                      <div className="flex shrink-0 flex-wrap items-center gap-2">
                        <StatusBadge status="default" label={versionSourceLabel(version.Source)} />
                        {version.IsDefault && <StatusBadge status="active" label="默认" />}
                        {version.IsActive && <StatusBadge status="healthy" label="活跃" />}
                      </div>
                    </div>
                  </div>
                ))}

                {activeTool.LocalVersions.length === 0 && (
                  <div className="rounded-lg border border-dashed border-[#334155] bg-[#0f172a] px-4 py-6 text-center">
                    <p className="text-sm text-slate-400">暂未检测到本地版本。</p>
                  </div>
                )}
              </div>
            </SurfaceCard>
          </div>

          <SurfaceCard className="h-fit">
            <div className="mb-4">
              <h3 className="text-sm font-bold text-slate-200">安装预览</h3>
              <p className="mt-1 text-xs text-slate-400">输入目标版本，先查看 DevMan 将安装到哪里。</p>
            </div>

            <div className="space-y-3">
              <input
                type="text"
                value={previewVersion}
                onChange={(event) => setPreviewVersion(event.target.value)}
                placeholder="例如 1.25.0 / 22.11.0"
                className="w-full rounded-xl border border-[#334155] bg-[#0f172a] px-4 py-2.5 font-mono text-sm text-slate-200 placeholder:text-slate-500 focus:border-emerald-500/50 focus:outline-none focus:ring-2 focus:ring-emerald-500/50"
              />
              <Button
                className="w-full"
                variant="primary"
                onClick={handlePreviewInstall}
                disabled={!previewVersion.trim() || hasManagerConflict}
              >
                预览安装
              </Button>

              {hasManagerConflict && (
                <p className="text-xs text-amber-300">检测到外部版本管理器后，安装预览已暂停。</p>
              )}
            </div>

            {installPlan && (
              <div className="mt-5 space-y-3 rounded-lg border border-[#334155] bg-[#0f172a] p-4">
                <div>
                  <p className="text-[10px] uppercase text-slate-500">Target</p>
                  <p className="mt-1 break-all font-mono text-xs text-slate-200">{installPlan.TargetDir}</p>
                </div>
                <div>
                  <p className="text-[10px] uppercase text-slate-500">Reason</p>
                  <p className="mt-1 text-sm text-slate-300">{installPlan.ResolverReason || '已生成安装计划'}</p>
                </div>
                {installPlan.WillOverwrite && <StatusBadge status="warning" label="将覆盖已有目录" />}
              </div>
            )}
          </SurfaceCard>
        </div>
      )}
    </div>
  );
}
