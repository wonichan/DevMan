import { useEffect, useState, useRef, useCallback } from 'react';
import { GetEnvs } from '../api/app';
import { SearchIcon, CommandIcon } from './icons';
import { ManagementBadge } from './ui/ManagementBadge';
import type { Env } from '../devman-types';

type Page = 'dashboard' | 'environments' | 'migration' | 'cleaner' | 'versions' | 'settings';

interface SearchItem {
  id: string;
  label: string;
  subtitle?: string;
  type: 'page' | 'env' | 'action';
  icon: string;
  managed?: boolean;
  action: () => void;
}

interface Props {
  onNavigate: (page: Page) => void;
}

const pageItems: SearchItem[] = [
  { id: 'page-dashboard', label: '总览', subtitle: 'Dashboard', type: 'page', icon: '📊', action: () => {} },
  { id: 'page-environments', label: '环境管理', subtitle: 'Environments', type: 'page', icon: '🖥️', action: () => {} },
  { id: 'page-migration', label: '迁移向导', subtitle: 'Migration', type: 'page', icon: '📦', action: () => {} },
  { id: 'page-cleaner', label: '缓存清理', subtitle: 'Cleaner', type: 'page', icon: '🧹', action: () => {} },
  { id: 'page-versions', label: '版本管理', subtitle: 'Versions', type: 'page', icon: '🏷️', action: () => {} },
  { id: 'page-settings', label: '设置', subtitle: 'Settings', type: 'page', icon: '⚙️', action: () => {} },
];

const actionItems: SearchItem[] = [
  { id: 'action-scan', label: '扫描环境', subtitle: 'ScanAll', type: 'action', icon: '🔍', action: () => {} },
  { id: 'action-clean', label: '分析可清理项', subtitle: 'AnalyzeCleanable', type: 'action', icon: '🧹', action: () => {} },
];

function isTypingTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false;
  const tagName = target.tagName.toLowerCase();
  return tagName === 'input' || tagName === 'textarea' || tagName === 'select' || target.isContentEditable;
}

export default function GlobalSearch({ onNavigate }: Props) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [envs, setEnvs] = useState<Env[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        if (isTypingTarget(e.target)) return;
        e.preventDefault();
        setOpen((prev) => !prev);
      }
      if (e.key === 'Escape') {
        setOpen(false);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  useEffect(() => {
    if (open) {
      setQuery('');
      setSelectedIndex(0);
      GetEnvs().then(setEnvs).catch(() => {});
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [open]);

  const buildItems = useCallback((): SearchItem[] => {
    const envItems: SearchItem[] = [...envs]
      .sort((a, b) => {
        if (a.IsManaged !== b.IsManaged) return a.IsManaged ? -1 : 1;
        return a.Name.localeCompare(b.Name);
      })
      .map((env) => ({
        id: `env-${env.Key}`,
        label: env.Name,
        subtitle: env.Key,
        type: 'env',
        icon: env.Icon,
        managed: env.IsManaged,
        action: () => onNavigate('environments'),
      }));

    const pages = pageItems.map((p) => ({
      ...p,
      action: () => {
        const pageMap: Record<string, Page> = {
          'page-dashboard': 'dashboard',
          'page-environments': 'environments',
          'page-migration': 'migration',
          'page-cleaner': 'cleaner',
          'page-versions': 'versions',
          'page-settings': 'settings',
        };
        onNavigate(pageMap[p.id]);
      },
    }));

    const actions = actionItems.map((a) => ({
      ...a,
      action: () => {
        if (a.id === 'action-scan') onNavigate('dashboard');
        if (a.id === 'action-clean') onNavigate('cleaner');
      },
    }));

    const all = [...pages, ...envItems, ...actions];
    if (!query.trim()) return all;
    const q = query.toLowerCase();
    return all.filter(
      (i) =>
        i.label.toLowerCase().includes(q) ||
        (i.subtitle && i.subtitle.toLowerCase().includes(q))
    );
  }, [envs, query, onNavigate]);

  const items = buildItems();

  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (items.length === 0) return;
      setSelectedIndex((prev) => (prev + 1) % items.length);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (items.length === 0) return;
      setSelectedIndex((prev) => (prev - 1 + items.length) % items.length);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const item = items[selectedIndex];
      if (item) {
        item.action();
        setOpen(false);
      }
    }
  };

  useEffect(() => {
    if (listRef.current && items.length > 0) {
      const el = listRef.current.children[selectedIndex] as HTMLElement;
      if (el) {
        el.scrollIntoView({ block: 'nearest' });
      }
    }
  }, [selectedIndex, items.length]);

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-[#0f172a] border border-[#334155] text-slate-400 text-sm hover:text-slate-200 hover:border-[#475569] transition-colors"
      >
        <SearchIcon className="w-4 h-4" />
        <span className="hidden sm:inline">搜索...</span>
        <span className="hidden md:inline-flex items-center gap-1 text-xs text-slate-500 ml-2">
          <kbd className="px-1.5 py-0.5 rounded bg-[#1e293b] border border-[#334155] font-mono">Ctrl</kbd>
          <kbd className="px-1.5 py-0.5 rounded bg-[#1e293b] border border-[#334155] font-mono">K</kbd>
        </span>
      </button>

      {open && (
        <div
          className="fixed inset-0 z-50 flex items-start justify-center pt-[10vh] bg-black/60 backdrop-blur-sm animate-in fade-in duration-200"
          onClick={() => setOpen(false)}
        >
          <div
            className="w-full max-w-lg bg-[#0f172a] border border-[#334155] rounded-2xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center gap-3 px-4 py-3 border-b border-[#334155]">
              <SearchIcon className="w-5 h-5 text-slate-400" />
              <input
                ref={inputRef}
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="搜索页面、环境或操作..."
                className="flex-1 bg-transparent text-slate-200 text-sm placeholder:text-slate-500 focus:outline-none"
              />
              <button
                onClick={() => setOpen(false)}
                className="text-xs text-slate-500 hover:text-slate-300 transition-colors"
              >
                ESC
              </button>
            </div>

            <div ref={listRef} className="max-h-[50vh] overflow-y-auto py-2">
              {items.length === 0 && (
                <div className="px-4 py-8 text-center text-sm text-slate-500">
                  未找到匹配项
                </div>
              )}
              {items.map((item, idx) => (
                <button
                  key={item.id}
                  onClick={() => {
                    item.action();
                    setOpen(false);
                  }}
                  onMouseEnter={() => setSelectedIndex(idx)}
                  className={`w-full flex items-center gap-3 px-4 py-2.5 text-left transition-colors ${
                    idx === selectedIndex ? 'bg-[#1e293b]' : 'hover:bg-[#1e293b]/50'
                  }`}
                >
                  <span className="text-lg">{item.icon}</span>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-slate-200 truncate">{item.label}</p>
                    {item.subtitle && (
                      <p className="text-xs text-slate-500 truncate">{item.subtitle}</p>
                    )}
                  </div>
                  <span className="text-xs px-2 py-0.5 rounded bg-[#1e293b] border border-[#334155] text-slate-500 capitalize">
                    {item.type === 'page' ? '页面' : item.type === 'env' ? '环境' : '操作'}
                  </span>
                  {item.type === 'env' && <ManagementBadge managed={Boolean(item.managed)} />}
                </button>
              ))}
            </div>

            <div className="px-4 py-2 border-t border-[#334155] flex items-center gap-4 text-xs text-slate-500">
              <span className="flex items-center gap-1">
                <CommandIcon className="w-3 h-3" /> 打开
              </span>
              <span className="flex items-center gap-1">
                <kbd className="px-1 rounded bg-[#1e293b] border border-[#334155] font-mono">↑↓</kbd> 选择
              </span>
              <span className="flex items-center gap-1">
                <kbd className="px-1 rounded bg-[#1e293b] border border-[#334155] font-mono">↵</kbd> 确认
              </span>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
