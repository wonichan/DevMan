import { useEffect, useState } from 'react';
import { GetEnvs, GetEnvSummary } from '../bindings/go/main/App';
import { PageHeader } from '../components/ui/PageHeader';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { StatusBadge } from '../components/ui/StatusBadge';
import { EmptyState } from '../components/ui/EmptyState';
import { SearchIcon } from '../components/icons';
import type { EnvSummary } from '../devman-types';

export default function Versions() {
  const [envs, setEnvs] = useState<EnvSummary[]>([]);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const list = await GetEnvs();
      const summaries: EnvSummary[] = [];
      for (const e of list || []) {
        const s = await GetEnvSummary(e.Key);
        if (s) summaries.push(s);
      }
      setEnvs(summaries);
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div>
      <PageHeader 
        title="版本管理" 
        description="查看和管理已安装的版本，设置默认版本"
      />

      <div className="space-y-4">
        {envs.map(env => (
          <SurfaceCard key={env.Env.Key}>
            <div className="flex items-center gap-3 mb-4">
              <span className="text-2xl">{env.Env.Icon}</span>
              <div>
                <h3 className="font-bold text-slate-200">{env.Env.Name}</h3>
                <p className="text-xs text-slate-400">{env.Instances.length} 个版本已安装</p>
              </div>
            </div>

            <div className="space-y-2">
              {env.Instances.map(inst => (
                <div
                  key={inst.Id}
                  className={`flex items-center justify-between px-4 py-3 rounded-xl ${inst.IsDefault ? 'bg-[#0f172a] border border-emerald-500/20 shadow-[0_0_15px_rgba(16,185,129,0.05)]' : 'bg-[#1e293b] border border-[#334155]'}`}
                >
                  <div>
                    <p className="text-sm font-bold text-slate-200 font-mono">{inst.Version}</p>
                    <p className="text-xs text-slate-400 font-mono mt-0.5">{inst.InstallPath}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    {inst.IsDefault && (
                      <StatusBadge status="active" label="默认" />
                    )}
                    {inst.IsActive && (
                      <StatusBadge status="healthy" label="活跃" />
                    )}
                    <span className="text-xs text-slate-400 ml-2">{inst.Source}</span>
                  </div>
                </div>
              ))}
              {env.Instances.length === 0 && (
                <p className="text-sm text-slate-400 text-center py-4">未检测到安装版本</p>
              )}
            </div>
          </SurfaceCard>
        ))}

        {envs.length === 0 && (
          <EmptyState 
            icon={<SearchIcon className="w-6 h-6" />}
            title="暂无环境数据"
            description="请先前往「总览」页面点击「刷新环境数据」"
          />
        )}
      </div>
    </div>
  );
}
