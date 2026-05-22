import { useState } from 'react';
import { PageHeader } from '../components/ui/PageHeader';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { Button } from '../components/ui/Button';
import { SettingsIcon } from '../components/icons';

export default function Settings() {
  const [autoScan, setAutoScan] = useState(true);
  const [confirmDelete, setConfirmDelete] = useState(true);

  return (
    <div>
      <PageHeader 
        title="设置" 
        description="管理 DevMan 的偏好和配置"
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        {/* 通用设置 */}
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
                aria-checked={autoScan}
                aria-label="启动时自动扫描"
                onClick={() => setAutoScan(!autoScan)}
                className={`w-11 h-6 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-emerald-500/50 ${autoScan ? 'bg-emerald-500' : 'bg-slate-700'}`}
              >
                <div className={`w-5 h-5 bg-white rounded-full shadow transition-transform ${autoScan ? 'translate-x-5' : 'translate-x-0.5'}`} />
              </button>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-200">迁移前确认</p>
                <p className="text-xs text-slate-400">执行迁移前显示确认对话框</p>
              </div>
              <button
                role="switch"
                aria-checked={confirmDelete}
                aria-label="迁移前确认"
                onClick={() => setConfirmDelete(!confirmDelete)}
                className={`w-11 h-6 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-emerald-500/50 ${confirmDelete ? 'bg-emerald-500' : 'bg-slate-700'}`}
              >
                <div className={`w-5 h-5 bg-white rounded-full shadow transition-transform ${confirmDelete ? 'translate-x-5' : 'translate-x-0.5'}`} />
              </button>
            </div>
          </div>
        </SurfaceCard>

        {/* 数据管理 */}
        <SurfaceCard>
          <h3 className="font-bold text-slate-200 mb-4">数据管理</h3>
          <div className="space-y-4">
            <div>
              <p className="text-sm font-medium text-slate-200 mb-1">数据库路径</p>
              <p className="text-xs text-slate-400 font-mono bg-[#0f172a] rounded-lg px-3 py-2 border border-[#334155]">.devman.db</p>
              <p className="text-xs text-slate-400 mt-1">便携模式：数据库存放在与程序同一目录下</p>
            </div>
            <div>
              <p className="text-sm font-medium text-slate-200 mb-1">历史记录</p>
              <p className="text-xs text-slate-400 mb-2">保留最近 100 条迁移和清理操作记录</p>
              <Button variant="secondary" size="sm">
                导出历史记录
              </Button>
            </div>
          </div>
        </SurfaceCard>

        {/* 关于 */}
        <SurfaceCard className="lg:col-span-2">
          <h3 className="font-bold text-slate-200 mb-4">关于 DevMan</h3>
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-[#1e293b] border border-[#334155] rounded-2xl flex items-center justify-center text-xl">
              <SettingsIcon className="w-6 h-6 text-emerald-400" />
            </div>
            <div>
              <p className="text-sm font-bold text-slate-200">DevMan v0.1.0</p>
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
