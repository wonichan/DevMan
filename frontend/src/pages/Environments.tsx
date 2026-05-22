import { useEffect, useState } from 'react';
import { GetEnvs } from '../bindings/go/main/App';
import Panel from '../components/Panel';
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
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-devman-text-primary">🔧 环境管理</h1>
          <p className="text-sm text-devman-text-muted mt-1">查看和管理已安装的开发环境</p>
        </div>
      </div>

      <div className="flex items-center gap-4 mb-6">
        <div className="flex-1 relative">
          <input
            type="text"
            placeholder="搜索环境..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full px-4 py-2.5 bg-devman-panel-raised border border-devman-border rounded-[16px] text-sm text-devman-text-primary placeholder:text-devman-text-muted/50 focus:outline-none focus:border-devman-accent/50"
          />
        </div>
        <button className="px-4 py-2.5 bg-devman-panel-raised border border-devman-border rounded-xl text-sm text-devman-text-primary hover:bg-devman-border/20 transition-colors">
          🔄 刷新
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        {filtered.map((env) => (
          <Panel key={env.Key} className="p-5 hover:border-devman-accent/20 transition-colors">
            <div className="flex items-start justify-between mb-3">
              <div className="flex items-center gap-3">
                <span className="text-2xl">{env.Icon}</span>
                <div>
                  <h3 className="text-base font-bold text-devman-text-primary">{env.Name}</h3>
                  <p className="text-xs text-devman-text-muted">{env.Category}</p>
                </div>
              </div>
              <span className="text-xs px-2 py-1 rounded-full bg-green-500/10 text-green-400 font-medium">
                ✓ 正常
              </span>
            </div>
            <p className="text-xs text-devman-text-muted mb-4">{env.Description}</p>
            <div className="flex items-center justify-between">
              <span className="text-xs text-devman-text-muted">{env.Website}</span>
              <div className="flex gap-2">
                <button className="px-3 py-1.5 bg-devman-panel-raised border border-devman-border rounded-lg text-xs text-devman-text-primary hover:bg-devman-border/20 transition-colors">
                  迁移
                </button>
                <button className="px-3 py-1.5 bg-devman-panel-raised border border-devman-border rounded-lg text-xs text-devman-text-primary hover:bg-devman-border/20 transition-colors">
                  版本
                </button>
              </div>
            </div>
          </Panel>
        ))}
      </div>

      {filtered.length === 0 && (
        <div className="text-center py-20 text-devman-text-muted">
          <p className="text-lg mb-2">🔍 未找到匹配的环境</p>
          <p className="text-sm">请先点击「刷新」扫描系统环境</p>
        </div>
      )}
    </div>
  );
}
