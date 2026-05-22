type Page = 'dashboard' | 'environments' | 'migration' | 'cleaner' | 'versions' | 'settings';

interface Props {
  active: Page;
  onNavigate: (page: Page) => void;
}

const items: { key: Page; icon: string; label: string }[] = [
  { key: 'dashboard', icon: '📊', label: '总览' },
  { key: 'environments', icon: '🔧', label: '环境' },
  { key: 'migration', icon: '📦', label: '迁移' },
  { key: 'cleaner', icon: '🧹', label: '清理' },
  { key: 'versions', icon: '🔄', label: '版本' },
  { key: 'settings', icon: '⚙️', label: '设置' },
];

export default function Sidebar({ active, onNavigate }: Props) {
  return (
    <aside className="w-60 flex flex-col bg-devman-panel border-r border-devman-border/30">
      <div className="p-6">
        <div className="flex items-center gap-3 mb-1">
          <span className="text-2xl">🖥️</span>
          <span className="text-lg font-bold text-devman-text-primary">DevMan</span>
        </div>
        <p className="text-xs text-devman-text-muted">开发环境统一管理</p>
      </div>

      <nav className="flex-1 px-3 space-y-1">
        {items.map((item) => {
          const isActive = active === item.key;
          return (
            <button
              key={item.key}
              onClick={() => onNavigate(item.key)}
              className={`
                w-full flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all
                ${isActive
                  ? 'bg-devman-accent/10 text-devman-accent border border-devman-accent/20'
                  : 'text-devman-text-muted hover:bg-devman-panel-raised hover:text-devman-text-primary'
                }
              `}
            >
              <span className="text-lg">{item.icon}</span>
              <span>{item.label}</span>
            </button>
          );
        })}
      </nav>

      <div className="p-4 text-xs text-devman-text-muted/60 font-mono">
        v0.1.0
      </div>
    </aside>
  );
}
