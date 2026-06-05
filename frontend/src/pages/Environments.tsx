import { useEffect, useState } from 'react';
import { GetEnvs, ManageEnv, UnmanageEnv } from '../api/app';
import { PageHeader } from '../components/ui/PageHeader';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { Button } from '../components/ui/Button';
import { EmptyState } from '../components/ui/EmptyState';
import { StatusBadge } from '../components/ui/StatusBadge';
import { ManagementBadge } from '../components/ui/ManagementBadge';
import { RefreshIcon, SearchIcon } from '../components/icons';
import { useToast } from '../hooks/useToast';
import type { Env } from '../devman-types';

export default function Environments() {
  const [envs, setEnvs] = useState<Env[]>([]);
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(false);
  const [pendingKey, setPendingKey] = useState<string | null>(null);
  const { error, success } = useToast();

  const load = async () => {
    setLoading(true);
    try {
      const data = await GetEnvs();
      setEnvs(data || []);
    } catch (e: unknown) {
      error('Load failed', e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const updateManagement = async (env: Env) => {
    setPendingKey(env.Key);
    try {
      const updated = env.IsManaged ? await UnmanageEnv(env.Key) : await ManageEnv(env.Key);
      setEnvs((current) => current.map((item) => (item.Key === updated.Key ? updated : item)));
      success(updated.IsManaged ? 'Environment managed' : 'Environment unmanaged', updated.Name);
    } catch (e: unknown) {
      error(env.IsManaged ? 'Unmanage failed' : 'Manage failed', e instanceof Error ? e.message : String(e));
    } finally {
      setPendingKey(null);
    }
  };

  const filtered = envs.filter((env) => {
    const term = search.toLowerCase();
    return env.Name.toLowerCase().includes(term) || env.Key.toLowerCase().includes(term);
  });
  const managedCount = envs.filter((env) => env.IsManaged).length;
  const unmanagedCount = Math.max(0, envs.length - managedCount);

  return (
    <div>
      <PageHeader
        title="Environment Management"
        description="View scanned development environments and track approved tools."
      />

      <SurfaceCard className="mb-6">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h2 className="text-sm font-bold text-slate-200">Tracking summary</h2>
            <p className="text-xs text-slate-400 mt-1">
              Managed means DevMan can track this detected tool after user approval. It does not mean DevMan installed, upgrades, or owns the tool.
            </p>
          </div>
          <div className="grid grid-cols-3 gap-3 shrink-0">
            <div className="min-w-[84px] rounded-lg border border-[#334155] bg-[#0f172a] px-3 py-2 text-center">
              <p className="text-[10px] uppercase text-slate-500">Total</p>
              <p className="text-lg font-bold text-slate-100">{envs.length}</p>
            </div>
            <div className="min-w-[84px] rounded-lg border border-emerald-500/20 bg-emerald-500/10 px-3 py-2 text-center">
              <p className="text-[10px] uppercase text-emerald-400/80">Managed</p>
              <p className="text-lg font-bold text-emerald-400">{managedCount}</p>
            </div>
            <div className="min-w-[84px] rounded-lg border border-[#334155] bg-slate-700/40 px-3 py-2 text-center">
              <p className="text-[10px] uppercase text-slate-400">Unmanaged</p>
              <p className="text-lg font-bold text-slate-200">{unmanagedCount}</p>
            </div>
          </div>
        </div>
      </SurfaceCard>

      <div className="flex items-center gap-4 mb-6">
        <div className="flex-1 relative">
          <input
            type="text"
            placeholder="Search environments..."
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            className="w-full px-4 py-2.5 bg-[#1e293b]/80 border border-[#334155] rounded-xl text-sm text-slate-200 placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500/50 transition-all"
          />
        </div>
        <Button variant="secondary" onClick={load} isLoading={loading}>
          <RefreshIcon className="mr-2 h-4 w-4" />
          Refresh
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        {filtered.map((env) => {
          const isPending = pendingKey === env.Key;
          return (
            <SurfaceCard key={env.Key} variant="interactive">
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3 min-w-0">
                  <span className="text-2xl shrink-0">{env.Icon}</span>
                  <div className="min-w-0">
                    <h3 className="text-base font-bold text-slate-200 truncate">{env.Name}</h3>
                    <p className="text-xs text-slate-400 truncate">{env.Category}</p>
                  </div>
                </div>
                <StatusBadge status="healthy" label="Healthy" />
              </div>

              <p className="text-xs text-slate-400 mb-4 h-8 overflow-hidden line-clamp-2">{env.Description}</p>

              <div className="flex items-center justify-between gap-4">
                <span className="text-xs text-slate-500 truncate">{env.Website}</span>
                <div className="flex items-center gap-2 shrink-0">
                  <ManagementBadge managed={env.IsManaged} className="min-w-[86px] text-center py-1" />
                  <Button
                    size="sm"
                    variant={env.IsManaged ? 'ghost' : 'secondary'}
                    className="w-[92px]"
                    onClick={() => updateManagement(env)}
                    isLoading={isPending}
                    disabled={pendingKey !== null && !isPending}
                  >
                    {env.IsManaged ? 'Unmanage' : 'Manage'}
                  </Button>
                </div>
              </div>
            </SurfaceCard>
          );
        })}
      </div>

      {filtered.length === 0 && (
        <EmptyState
          icon={<SearchIcon className="w-6 h-6" />}
          title="No environments found"
          description={search ? 'Try another search term.' : 'No scanned environment data is available.'}
        />
      )}
    </div>
  );
}
