import { useEffect, useState } from 'react';
import { GetEnvs } from '../api/app';
import { PageHeader } from '../components/ui/PageHeader';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { Button } from '../components/ui/Button';
import { EmptyState } from '../components/ui/EmptyState';
import { StatusBadge } from '../components/ui/StatusBadge';
import { RefreshIcon, SearchIcon } from '../components/icons';
import { useToast } from '../hooks/useToast';
import type { Env } from '../devman-types';

export default function Environments() {
  const [envs, setEnvs] = useState<Env[]>([]);
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(false);
  const { error } = useToast();

  const load = async () => {
    setLoading(true);
    try {
      const data = await GetEnvs();
      setEnvs(data || []);
    } catch (e: unknown) {
      error('加载失败', e instanceof Error ? e.message : String(e));
    }
    setLoading(false);
  };

  useEffect(() => {
    load();
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
        <Button variant="secondary" onClick={load} isLoading={loading}>
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
                <span className={`text-xs px-2 py-1 rounded-md ${env.IsManaged ? 'bg-emerald-500/10 text-emerald-400' : 'bg-slate-700 text-slate-400'}`}>
                  {env.IsManaged ? '已管理' : '未管理'}
                </span>
              </div>
            </div>
          </SurfaceCard>
        ))}
      </div>

      {filtered.length === 0 && (
        <EmptyState 
          icon={<SearchIcon className="w-6 h-6" />}
          title="未找到环境"
          description={search ? '尝试其他关键词搜索' : '暂无已扫描的环境数据'}
        />
      )}
    </div>
  );
}
