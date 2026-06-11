import { useEffect, useMemo, useState } from 'react';
import {
  ScanAll,
  FetchOfficialVersions,
  InstallVersion,
  ListToolVersions,
  PreviewVersionInstall,
  SwitchVersion,
  UninstallVersion,
} from '../api/app';
import { Button } from '../components/ui/Button';
import { EmptyState } from '../components/ui/EmptyState';
import { PageHeader } from '../components/ui/PageHeader';
import { StatusBadge } from '../components/ui/StatusBadge';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { RefreshIcon, SearchIcon, TrashIcon, VersionsIcon, WarningIcon } from '../components/icons';
import { useToast } from '../hooks/useToast';
import { usePageActions } from '../hooks/usePageActions';
import type {
  AvailableVersion,
  ManagedVersion,
  ToolVersionState,
  VersionInstallPlan,
  VersionOperationResult,
} from '../devman-types';
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

function formatDate(value?: string) {
  if (!value) return '未知日期';
  return value.slice(0, 10);
}

function affectedEnvironmentEntries(result: VersionOperationResult) {
  return Object.entries(result.AffectedEnvironment || {});
}

function versionCanBeUninstalled(version: ManagedVersion) {
  if (version.IsDefault || version.IsActive) return false;
  if (version.Source === 'version_manager') return false;
  if (version.Source === 'external') return true;
  return version.CanDelete || version.DeletePolicy === 'direct';
}

export default function Versions() {
  const [tools, setTools] = useState<ToolVersionState[]>([]);
  const [activeToolKey, setActiveToolKey] = useState('go');
  const [loading, setLoading] = useState(false);
  const [previewVersion, setPreviewVersion] = useState('');
  const [installPlan, setInstallPlan] = useState<VersionInstallPlan | null>(null);
  const [officialVersions, setOfficialVersions] = useState<AvailableVersion[]>([]);
  const [operationResult, setOperationResult] = useState<VersionOperationResult | null>(null);
  const [actionKey, setActionKey] = useState<string | null>(null);
  const { error, success } = useToast();

  const orderedTools = useMemo(() => sortTools(tools || []), [tools]);
  const activeTool = orderedTools.find((tool) => tool.ToolKey === activeToolKey);
  const localVersions = activeTool && Array.isArray(activeTool.LocalVersions) ? activeTool.LocalVersions : [];
  const activeConflict = conflictLabel(activeTool?.ManagerConflict);
  const hasManagerConflict = Boolean(activeConflict);
  const canPlanInstall = Boolean(previewVersion.trim()) && !hasManagerConflict && !actionKey;

  const showOperationResult = (
    result: VersionOperationResult,
    successTitle: string,
    failureTitle: string,
  ) => {
    setOperationResult(result);
    if (result.Success) {
      success(successTitle, result.Message);
      return;
    }
    error(failureTitle, result.Message || '后端返回操作失败。');
  };

  const loadData = async (refreshScan = false) => {
    setLoading(true);
    try {
      if (refreshScan) {
        await ScanAll();
      }
      const data = await ListToolVersions();
      const list = Array.isArray(data) ? data : [];
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

  usePageActions('versions', { refresh: () => loadData(true) });

  const handleSelectTool = (toolKey: string) => {
    setActiveToolKey(toolKey);
    setPreviewVersion('');
    setInstallPlan(null);
    setOfficialVersions([]);
  };

  const handleFetchOfficial = async () => {
    setActionKey('fetch-official');
    try {
      const catalog = await FetchOfficialVersions(activeToolKey);
      setOfficialVersions(Array.isArray(catalog?.Versions) ? catalog.Versions : []);
      success('官方版本已刷新');
    } catch (e: unknown) {
      error('官方版本查询失败', e instanceof Error ? e.message : String(e));
    } finally {
      setActionKey(null);
    }
  };

  const handleSelectOfficialVersion = (version: AvailableVersion) => {
    setPreviewVersion(version.Version);
    setInstallPlan(null);
  };

  const handlePreviewInstall = async () => {
    const version = previewVersion.trim();
    if (!version) return;
    if (hasManagerConflict) {
      error('无法生成安装预览', activeConflict || '检测到外部版本管理器。');
      return;
    }

    setActionKey('preview-install');
    setInstallPlan(null);
    try {
      const plan = await PreviewVersionInstall(activeToolKey, version);
      setInstallPlan(plan);
      success('安装预览已生成', plan.ResolverReason || '已解析目标安装路径。');
    } catch (e: unknown) {
      error('生成安装预览失败', e instanceof Error ? e.message : String(e));
    } finally {
      setActionKey(null);
    }
  };

  const handleInstall = async () => {
    if (!installPlan || hasManagerConflict) return;
    setActionKey('install');
    try {
      const result = await InstallVersion(installPlan.ToolKey, installPlan.Version, installPlan.TargetDir);
      showOperationResult(result, '安装完成', '安装失败');
      setInstallPlan(null);
      await loadData();
    } catch (e: unknown) {
      error('安装失败', e instanceof Error ? e.message : String(e));
    } finally {
      setActionKey(null);
    }
  };

  const handleSwitch = async (version: ManagedVersion) => {
    if (version.IsDefault || version.IsActive || hasManagerConflict) return;
    setActionKey(`switch-${version.Id}`);
    try {
      const result = await SwitchVersion(version.ToolKey, version.Id);
      showOperationResult(result, '切换完成', '切换失败');
      await loadData();
    } catch (e: unknown) {
      error('切换失败', e instanceof Error ? e.message : String(e));
    } finally {
      setActionKey(null);
    }
  };

  const handleUninstall = async (version: ManagedVersion, force: boolean) => {
    if (version.IsDefault || version.IsActive) {
      error('无法卸载当前使用的版本', '请先切换到其他版本后再卸载。');
      return;
    }
    if (version.Source === 'version_manager') {
      error('无法删除版本管理器托管的版本', '请在对应的外部版本管理器中处理此版本。');
      return;
    }
    if (version.Source === 'devman' && !version.CanDelete && version.DeletePolicy !== 'direct') {
      error('无法卸载此版本', version.DeletePolicy || '后端策略不允许删除。');
      return;
    }

    const isExternal = version.Source !== 'devman';
    if (isExternal || force) {
      const message = isExternal
        ? `确认删除外部版本 ${version.Version}？这可能会删除 DevMan 之外安装的文件。`
        : `确认卸载 ${version.Version}？`;
      if (!window.confirm(message)) return;
    }

    setActionKey(`uninstall-${version.Id}`);
    try {
      const result = await UninstallVersion(version.Id, force);
      showOperationResult(result, '卸载完成', '卸载失败');
      await loadData();
    } catch (e: unknown) {
      error('卸载失败', e instanceof Error ? e.message : String(e));
    } finally {
      setActionKey(null);
    }
  };

  return (
    <div>
      <PageHeader
        title="版本管理"
        description="查看本地工具版本，刷新官方版本列表，并执行安装、切换或卸载操作。"
        actions={
          <Button variant="secondary" onClick={() => loadData(true)} isLoading={loading}>
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
          description="尚未发现可管理的 Go、Node.js、Bun 或 Flutter 本地版本。"
        />
      )}

      {activeTool && (
        <div className="grid grid-cols-1 gap-5 xl:grid-cols-[minmax(0,1fr)_400px]">
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
                      <StatusBadge status="healthy" label="可安装" />
                    )}
                  </div>
                  <p className="mt-2 text-sm text-slate-400">
                    当前默认版本：
                    <span className="font-mono text-slate-200">{activeTool.CurrentDefault?.Version || '未设置'}</span>
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
                  <p className="text-2xl font-bold text-slate-100">{localVersions.length}</p>
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
                  <p className="mt-1 text-xs text-slate-400">显示已检测到的安装路径、来源和当前状态。</p>
                </div>
              </div>

              <div className="space-y-2">
                {localVersions.map((version) => {
                  const switchDisabled = version.IsDefault || version.IsActive || hasManagerConflict || Boolean(actionKey);
                  const uninstallDisabled = !versionCanBeUninstalled(version) || Boolean(actionKey);
                  const isExternal = version.Source !== 'devman';
                  return (
                    <div
                      key={`${version.ToolKey}-${version.Version}-${version.InstallPath}`}
                      className={`rounded-lg border px-4 py-3 ${
                        version.IsDefault
                          ? 'border-emerald-500/30 bg-emerald-500/10'
                          : 'border-[#334155] bg-[#0f172a]'
                      }`}
                    >
                      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                        <div className="min-w-0">
                          <p className="font-mono text-sm font-bold text-slate-100">{version.Version}</p>
                          <p className="mt-1 truncate font-mono text-xs text-slate-400">{version.InstallPath}</p>
                          {!version.CanDelete && (
                            <p className="mt-1 text-xs text-amber-300">
                              删除策略：{version.DeletePolicy || '此版本当前不建议删除'}
                            </p>
                          )}
                        </div>
                        <div className="flex shrink-0 flex-wrap items-center gap-2">
                          <StatusBadge status="default" label={versionSourceLabel(version.Source)} />
                          {version.IsDefault && <StatusBadge status="active" label="默认" />}
                          {version.IsActive && <StatusBadge status="healthy" label="活跃" />}
                        </div>
                      </div>
                      <div className="mt-3 flex flex-wrap gap-2">
                        <Button
                          size="sm"
                          variant="secondary"
                          disabled={switchDisabled}
                          isLoading={actionKey === `switch-${version.Id}`}
                          onClick={() => handleSwitch(version)}
                        >
                          设为默认
                        </Button>
                        <Button
                          size="sm"
                          variant="danger"
                          disabled={uninstallDisabled}
                          isLoading={actionKey === `uninstall-${version.Id}`}
                          onClick={() => handleUninstall(version, isExternal)}
                        >
                          <TrashIcon className="mr-1 h-3.5 w-3.5" />
                          {isExternal ? '强制删除' : '卸载'}
                        </Button>
                      </div>
                    </div>
                  );
                })}

                {localVersions.length === 0 && (
                  <div className="rounded-lg border border-dashed border-[#334155] bg-[#0f172a] px-4 py-6 text-center">
                    <p className="text-sm text-slate-400">暂未检测到本地版本。</p>
                  </div>
                )}
              </div>
            </SurfaceCard>

            {operationResult && (
              <SurfaceCard>
                <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <h3 className="text-sm font-bold text-slate-200">最近操作</h3>
                    <p className="mt-1 text-xs text-slate-400">
                      {operationResult.ToolKey} {operationResult.Version}
                    </p>
                  </div>
                  <StatusBadge
                    status={operationResult.Success ? 'healthy' : 'danger'}
                    label={operationResult.Success ? '成功' : '失败'}
                  />
                </div>
                <p className="text-sm text-slate-300">{operationResult.Message || '操作已返回结果。'}</p>

                {(operationResult.AffectedPaths || []).length > 0 && (
                  <div className="mt-4">
                    <p className="text-[10px] uppercase text-slate-500">Affected Paths</p>
                    <div className="mt-2 space-y-1">
                      {(operationResult.AffectedPaths || []).map((path) => (
                        <p key={path} className="break-all font-mono text-xs text-slate-300">
                          {path}
                        </p>
                      ))}
                    </div>
                  </div>
                )}

                {affectedEnvironmentEntries(operationResult).length > 0 && (
                  <div className="mt-4">
                    <p className="text-[10px] uppercase text-slate-500">Environment</p>
                    <div className="mt-2 space-y-1">
                      {affectedEnvironmentEntries(operationResult).map(([key, value]) => (
                        <p key={key} className="break-all font-mono text-xs text-slate-300">
                          {key}={value}
                        </p>
                      ))}
                    </div>
                  </div>
                )}

                {(operationResult.VerificationCommand || operationResult.VerificationOutput) && (
                  <div className="mt-4 rounded-lg border border-[#334155] bg-[#0f172a] p-3">
                    {operationResult.VerificationCommand && (
                      <p className="break-all font-mono text-xs text-emerald-300">
                        {operationResult.VerificationCommand}
                      </p>
                    )}
                    {operationResult.VerificationOutput && (
                      <pre className="mt-2 max-h-40 overflow-auto whitespace-pre-wrap break-words text-xs text-slate-300">
                        {operationResult.VerificationOutput}
                      </pre>
                    )}
                  </div>
                )}
              </SurfaceCard>
            )}
          </div>

          <div className="space-y-5">
            <SurfaceCard className="h-fit">
              <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 className="text-sm font-bold text-slate-200">官方版本</h3>
                  <p className="mt-1 text-xs text-slate-400">拉取官方版本列表，点击版本后生成安装预览。</p>
                </div>
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={handleFetchOfficial}
                  isLoading={actionKey === 'fetch-official'}
                  disabled={Boolean(actionKey)}
                >
                  刷新官方版本
                </Button>
              </div>

              <div className="max-h-80 space-y-2 overflow-auto pr-1">
                {(officialVersions || []).slice(0, 20).map((version) => (
                  <button
                    key={`${version.Version}-${version.Arch}-${version.DownloadUrl}`}
                    type="button"
                    onClick={() => handleSelectOfficialVersion(version)}
                    className={`w-full rounded-lg border px-3 py-2 text-left transition-colors ${
                      previewVersion === version.Version
                        ? 'border-emerald-500/40 bg-emerald-500/10'
                        : 'border-[#334155] bg-[#0f172a] hover:border-[#475569]'
                    }`}
                  >
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-mono text-sm font-bold text-slate-100">{version.Version}</span>
                      {version.Arch && <StatusBadge status="default" label={version.Arch} />}
                      {version.Stable && <StatusBadge status="healthy" label="stable" />}
                    </div>
                    <p className="mt-1 text-xs text-slate-500">{formatDate(version.ReleaseDate)}</p>
                  </button>
                ))}

                {(officialVersions || []).length === 0 && (
                  <div className="rounded-lg border border-dashed border-[#334155] bg-[#0f172a] px-4 py-6 text-center">
                    <p className="text-sm text-slate-400">尚未加载官方版本。</p>
                  </div>
                )}
              </div>
            </SurfaceCard>

            <SurfaceCard className="h-fit">
              <div className="mb-4">
                <h3 className="text-sm font-bold text-slate-200">安装预览</h3>
                <p className="mt-1 text-xs text-slate-400">输入或选择目标版本，先查看 DevMan 将安装到哪里。</p>
              </div>

              <div className="space-y-3">
                <input
                  type="text"
                  value={previewVersion}
                  onChange={(event) => {
                    setPreviewVersion(event.target.value);
                    setInstallPlan(null);
                  }}
                  placeholder="例如 1.25.0 / 22.11.0"
                  className="w-full rounded-xl border border-[#334155] bg-[#0f172a] px-4 py-2.5 font-mono text-sm text-slate-200 placeholder:text-slate-500 focus:border-emerald-500/50 focus:outline-none focus:ring-2 focus:ring-emerald-500/50"
                />
                <Button
                  className="w-full"
                  variant="primary"
                  onClick={handlePreviewInstall}
                  disabled={!canPlanInstall}
                  isLoading={actionKey === 'preview-install'}
                >
                  预览安装
                </Button>

                {hasManagerConflict && (
                  <p className="text-xs text-amber-300">检测到外部版本管理器后，安装和切换操作已暂停。</p>
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
                  <Button
                    className="w-full"
                    variant="primary"
                    onClick={handleInstall}
                    disabled={hasManagerConflict || Boolean(actionKey)}
                    isLoading={actionKey === 'install'}
                  >
                    下载并安装
                  </Button>
                </div>
              )}
            </SurfaceCard>
          </div>
        </div>
      )}
    </div>
  );
}
