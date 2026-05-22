import type { ComponentType } from 'react';
import { DashboardIcon, EnvironmentsIcon, MigrationIcon, CleanerIcon, VersionsIcon, SettingsIcon, CheckIcon } from './icons';
import type { IconProps } from './icons';

type Page = 'dashboard' | 'environments' | 'migration' | 'cleaner' | 'versions' | 'settings';

interface Props {
  active: Page;
  onNavigate: (page: Page) => void;
}

const items: { key: Page; icon: ComponentType<IconProps>; label: string }[] = [
  { key: 'dashboard', icon: DashboardIcon, label: '总览' },
  { key: 'environments', icon: EnvironmentsIcon, label: '环境' },
  { key: 'migration', icon: MigrationIcon, label: '迁移' },
  { key: 'cleaner', icon: CleanerIcon, label: '清理' },
  { key: 'versions', icon: VersionsIcon, label: '版本' },
  { key: 'settings', icon: SettingsIcon, label: '设置' },
];

export default function Sidebar({ active, onNavigate }: Props) {
  return (
    <aside className="w-64 flex flex-col bg-[#020617]/50 backdrop-blur-xl border-r border-[#1e293b]">
      <div className="p-6 pb-8">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20">
            <CheckIcon className="w-5 h-5 text-emerald-400" />
          </div>
          <span className="text-lg font-bold text-slate-100 tracking-tight">DevMan</span>
        </div>
      </div>

      <nav className="flex-1 px-3 space-y-1">
        {items.map((item) => {
          const isActive = active === item.key;
          const Icon = item.icon;
          return (
            <button
              key={item.key}
              onClick={() => onNavigate(item.key)}
              className={`
                group relative w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors
                ${isActive
                  ? 'bg-[#1e293b] text-emerald-400'
                  : 'text-slate-400 hover:bg-[#0f172a] hover:text-slate-200'
                }
              `}
            >
              {isActive && (
                <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-5 bg-emerald-500 rounded-r-full" />
              )}
              <Icon className="w-5 h-5 flex-shrink-0" />
              <span>{item.label}</span>
            </button>
          );
        })}
      </nav>

      <div className="p-4">
        <div className="px-3 py-2 rounded-lg bg-[#0f172a] border border-[#1e293b] flex items-center justify-between">
          <span className="text-xs text-slate-400">Version</span>
          <span className="text-xs font-mono text-emerald-500/80">v0.1.0</span>
        </div>
      </div>
    </aside>
  );
}
