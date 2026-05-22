import { useEffect, useState } from 'react';
import { GetEnvs } from '../bindings/go/main/App';
import { PageHeader } from '../components/ui/PageHeader';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { Button } from '../components/ui/Button';
import { EmptyState } from '../components/ui/EmptyState';
import { StatusBadge } from '../components/ui/StatusBadge';
import { RefreshIcon, SearchIcon } from '../components/icons';
import type { Env } from '../devman-types';

export default function Environments() {
  const [envs, setEnvs] = useState<Env[]>([]);
  const [search, setSearch] = useState('');

  useEffect(() => {
    GetEnvs().then(setEnvs).catch(console.error);
  }, []);

  const filtered = envs.filter(e =>
    e.Name.toLowerCase().includes(search.toLowerCase()) ||
    e.Key.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div>
      <PageHeader 
        title="环境管理" 
        description="查看和管理已安装的开发环境"
      />

      <div className="flex items-center gap-4 mb-6">
        <div className="flex-1 relative">
          <input
            type="text"
            placeholder="搜索环境..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full px-4 py-2.5 bg-[#1e293b]/80 border border-[#334155] rounded-xl text-sm text-slate-200 placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500/50 transition-all"
          />
        </div>
        <Button variant="secondary" onClick={() => GetEnvs().then(setEnvs).catch(console.error)}>
          <RefreshIcon className="mr-2 h-4 w-4" />
          刷新
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        {filtered.map((env) => (
          <SurfaceCard key={env.Key} variant="interactive">
            <div className="flex items-start justify-between mb-3">
              <div className="flex items-center gap-3">
                <span className="text-2xl">{env.Icon}</span>
                <div>
                  <h3 className="text-base font-bold text-slate-200">{env.Name}</h3>
                  <p className="text-xs text-slate-400">{env.Category}</p>
                </div>
              </div>
              <StatusBadge status="healthy" label="✓ 正常" />
            </div>
            <p className="text-xs text-slate-400 mb-4 h-8 overflow-hidden line-clamp-2">{env.Description}</p>
            <div className="flex items-center justify-between">
              <span className="text-xs text-slate-500">{env.Website}</span>
              <div className="flex gap-2">
                <Button variant="secondary" size="sm">
                  迁移
                </Button>
                <Button variant="secondary" size="sm">
                  版本
                </Button>
              </div>
            </div>
          </SurfaceCard>
        ))}
      </div>

      {filtered.length === 0 && (
        <EmptyState 
          icon={<SearchIcon className="w-6 h-6" />}
          title="未找到匹配的环境"
          description="请先点击「刷新」扫描系统环境，或尝试更换搜索词"
        />
      )}
    </div>
  );
}
