import { useEffect, useState } from 'react';
import { PageHeader } from '../components/ui/PageHeader';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { Button } from '../components/ui/Button';
import { SettingsIcon, TrashIcon, PlusIcon } from '../components/icons';
import { useToast } from '../hooks/useToast';
import { useConfirm } from '../hooks/useConfirm';
import { GetSettings, SaveSettings } from '../api/app';
import type { AppSettings } from '../devman-types';

export default function Settings() {
  const [settings, setSettings] = useState<AppSettings>({
    AutoScanOnStartup: true,
    ConfirmBeforeMigration: true,
    Theme: 'dark',
    CustomScanPaths: [],
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [newPath, setNewPath] = useState('');
  const { success, error } = useToast();
  const { confirm } = useConfirm();

  useEffect(() => {
    GetSettings()
      .then((s) => {
        setSettings(s);
        applyTheme(s.Theme);
      })
      .catch((e) => {
        error('加载设置失败', e instanceof Error ? e.message : String(e));
      })
      .finally(() => setLoading(false));
  }, []);

  const applyTheme = (theme: string) => {
    document.documentElement.dataset.theme = theme;
  };

  const update = <K extends keyof AppSettings>(key: K, value: AppSettings[K]) => {
    setSettings((prev) => ({ ...prev, [key]: value }));
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await SaveSettings(settings);
      success('保存成功', '设置已更新');
      applyTheme(settings.Theme);
    } catch (e) {
      error('保存失败', e instanceof Error ? e.message : String(e));
    }
    setSaving(false);
  };

  const addPath = () => {
    const trimmed = newPath.trim();
    if (!trimmed) return;
    if (settings.CustomScanPaths.includes(trimmed)) {
      error('路径已存在', trimmed);
      return;
    }
    setSettings((prev) => ({
      ...prev,
      CustomScanPaths: [...prev.CustomScanPaths, trimmed],
    }));
    setNewPath('');
  };

  const removePath = async (path: string) => {
    const ok = await confirm({
      title: '删除自定义扫描路径',
      description: `确定要删除路径 ${path} 吗？`,
      confirmText: '删除',
      cancelText: '取消',
      variant: 'danger',
    });
    if (!ok) return;
    setSettings((prev) => ({
      ...prev,
      CustomScanPaths: prev.CustomScanPaths.filter((p) => p !== path),
    }));
  };

  if (loading) {
    return (
      <div>
        <PageHeader title="设置" description="管理 DevMan 的偏好和配置" />
        <SurfaceCard>
          <div className="flex items-center justify-center py-12">
            <div className="w-6 h-6 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
            <span className="ml-3 text-sm text-slate-400">加载设置中...</span>
          </div>
        </SurfaceCard>
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        title="设置"
        description="管理 DevMan 的偏好和配置"
        actions={
          <Button variant="primary" onClick={handleSave} isLoading={saving}>
            保存设置
          </Button>
        }
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <SurfaceCard>
          <h3 className="font-bold text-slate-200 mb-4">通用设置</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-200">启动时自动扫描</p>
                <p className="text-xs text-slate-400">打开应用时自动检测开发环境变化</p>
              </div>
              <button
                role="switch"
                aria-checked={settings.AutoScanOnStartup}
                aria-label="启动时自动扫描"
                onClick={() => update('AutoScanOnStartup', !settings.AutoScanOnStartup)}
                className={`w-11 h-6 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-emerald-500/50 ${settings.AutoScanOnStartup ? 'bg-emerald-500' : 'bg-slate-700'}`}
              >
                <div className={`w-5 h-5 bg-white rounded-full shadow transition-transform ${settings.AutoScanOnStartup ? 'translate-x-5' : 'translate-x-0.5'}`} />
              </button>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-200">迁移前确认</p>
                <p className="text-xs text-slate-400">执行迁移前显示确认对话框</p>
              </div>
              <button
                role="switch"
                aria-checked={settings.ConfirmBeforeMigration}
                aria-label="迁移前确认"
                onClick={() => update('ConfirmBeforeMigration', !settings.ConfirmBeforeMigration)}
                className={`w-11 h-6 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-emerald-500/50 ${settings.ConfirmBeforeMigration ? 'bg-emerald-500' : 'bg-slate-700'}`}
              >
                <div className={`w-5 h-5 bg-white rounded-full shadow transition-transform ${settings.ConfirmBeforeMigration ? 'translate-x-5' : 'translate-x-0.5'}`} />
              </button>
            </div>
            <div>
              <p className="text-sm font-medium text-slate-200 mb-2">主题</p>
              <div className="flex gap-2">
                {(['dark', 'system'] as const).map((t) => (
                  <button
                    key={t}
                    onClick={() => update('Theme', t)}
                    className={`px-3 py-1.5 rounded-lg text-sm border transition-colors ${
                      settings.Theme === t
                        ? 'bg-emerald-500/20 border-emerald-500/50 text-emerald-400'
                        : 'bg-[#0f172a] border-[#334155] text-slate-400 hover:text-slate-200'
                    }`}
                  >
                    {t === 'dark' ? '深色' : '跟随系统'}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </SurfaceCard>

        <SurfaceCard>
          <h3 className="font-bold text-slate-200 mb-4">自定义扫描路径</h3>
          <div className="space-y-3">
            <div className="flex gap-2">
              <input
                type="text"
                value={newPath}
                onChange={(e) => setNewPath(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') addPath();
                }}
                placeholder="输入路径并回车添加..."
                className="flex-1 px-4 py-2 bg-[#0f172a] border border-[#334155] rounded-xl text-sm text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/20 transition-all"
              />
              <Button variant="secondary" size="sm" onClick={addPath}>
                <PlusIcon className="w-4 h-4" />
              </Button>
            </div>
            {settings.CustomScanPaths.length === 0 && (
              <p className="text-xs text-slate-500">暂无自定义扫描路径</p>
            )}
            <div className="space-y-2 max-h-48 overflow-y-auto">
              {settings.CustomScanPaths.map((path) => (
                <div
                  key={path}
                  className="flex items-center justify-between px-3 py-2 bg-[#0f172a] border border-[#334155] rounded-lg"
                >
                  <span className="text-sm text-slate-300 font-mono truncate">{path}</span>
                  <button
                    onClick={() => removePath(path)}
                    className="p-1 rounded hover:bg-red-500/20 text-slate-400 hover:text-red-400 transition-colors"
                    aria-label={`删除 ${path}`}
                  >
                    <TrashIcon className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          </div>
        </SurfaceCard>

        <SurfaceCard className="lg:col-span-2">
          <h3 className="font-bold text-slate-200 mb-4">关于 DevMan</h3>
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-[#1e293b] border border-[#334155] rounded-2xl flex items-center justify-center text-xl">
              <SettingsIcon className="w-6 h-6 text-emerald-400" />
            </div>
            <div>
              <p className="text-sm font-bold text-slate-200">DevMan v0.2.0</p>
              <p className="text-xs text-slate-400">Windows 开发环境管理器</p>
              <p className="text-xs text-slate-400 mt-1">
                基于 Wails v2 + Go + React + Tailwind CSS 构建
              </p>
            </div>
          </div>
        </SurfaceCard>
      </div>
    </div>
  );
}
