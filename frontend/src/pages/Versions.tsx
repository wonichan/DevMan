import { useEffect, useState } from 'react';
import { GetEnvs, GetEnvSummary } from '../bindings/go/main/App';
import Panel from '../components/Panel';
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
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-devman-text-primary">📋 版本管理</h1>
        <p className="text-sm text-devman-text-muted mt-1">查看和管理已安装的版本，设置默认版本</p>
      </div>

      <div className="space-y-4">
        {envs.map(env => (
          <Panel key={env.Env.Key} className="p-5">
            <div className="flex items-center gap-3 mb-4">
              <span className="text-2xl">{env.Env.Icon}</span>
              <div>
                <h3 className="font-bold text-devman-text-primary">{env.Env.Name}</h3>
                <p className="text-xs text-devman-text-muted">{env.Instances.length} 个版本已安装</p>
              </div>
            </div>

            <div className="space-y-2">
              {env.Instances.map(inst => (
                <div
                  key={inst.Id}
                  className={`flex items-center justify-between px-4 py-3 rounded-xl ${inst.IsDefault ? 'bg-devman-accent/10 border border-devman-accent/20' : 'bg-devman-panel-raised border border-devman-border/30'}`}
                >
                  <div>
                    <p className="text-sm font-bold text-devman-text-primary">{inst.Version}</p>
                    <p className="text-xs text-devman-text-muted">{inst.InstallPath}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    {inst.IsDefault && (
                      <span className="px-2 py-1 bg-devman-accent/20 text-devman-accent rounded-lg text-xs font-bold">
                        默认
                      </span>
                    )}
                    {inst.IsActive && (
                      <span className="px-2 py-1 bg-green-500/15 text-green-400 rounded-lg text-xs font-bold">
                        活跃
                      </span>
                    )}
                    <span className="text-xs text-devman-text-muted">{inst.Source}</span>
                  </div>
                </div>
              ))}
              {env.Instances.length === 0 && (
                <p className="text-sm text-devman-text-muted text-center py-4">未检测到安装版本</p>
              )}
            </div>
          </Panel>
        ))}

        {envs.length === 0 && (
          <div className="text-center py-20 text-devman-text-muted">
            <p className="text-lg mb-2">🔍 暂无环境数据</p>
            <p className="text-sm">请先前往「总览」页面点击「刷新环境数据」</p>
          </div>
        )}
      </div>
    </div>
  );
}
