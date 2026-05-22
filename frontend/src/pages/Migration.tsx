import { useEffect, useState } from 'react';
import { GetEnvs, GetEnvSummary, Migrate, GetDiskInfo } from '../bindings/go/main/App';
import Panel from '../components/Panel';
import type { EnvSummary, DiskInfo } from '../devman-types';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

export default function Migration() {
  const [step, setStep] = useState(1);
  const [envs, setEnvs] = useState<EnvSummary[]>([]);
  const [selectedEnv, setSelectedEnv] = useState<EnvSummary | null>(null);
  const [targetDir, setTargetDir] = useState('D:\\Dev');
  const [disks, setDisks] = useState<DiskInfo[]>([]);
  const [migrating, setMigrating] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [logs, setLogs] = useState<string[]>([]);

  useEffect(() => {
    loadEnvs();
    loadDisks();
  }, []);

  const loadEnvs = async () => {
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

  const loadDisks = async () => {
    try {
      const d = await GetDiskInfo();
      setDisks(d || []);
    } catch (e) {
      console.error(e);
    }
  };

  const handleMigrate = async () => {
    if (!selectedEnv) return;
    setMigrating(true);
    setLogs(['开始迁移...']);
    try {
      const res = await Migrate(selectedEnv.Env.Id, targetDir, false);
      setResult(res);
      setLogs(prev => [...prev, res.Message || '迁移完成']);
    } catch (e: any) {
      setLogs(prev => [...prev, '错误: ' + (e.message || String(e))]);
    }
    setMigrating(false);
  };

  const cDisk = disks.find(d => d.Letter === 'C:') || disks[0];
  const targetDisk = targetDir[0] + ':';
  const tDisk = disks.find(d => d.Letter === targetDisk);

  const steps = [
    { num: 1, label: '选择环境' },
    { num: 2, label: '选择目标' },
    { num: 3, label: '预览确认' },
    { num: 4, label: '执行迁移' },
  ];

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-devman-text-primary">📦 迁移向导</h1>
        <p className="text-sm text-devman-text-muted mt-1">将开发环境安全迁移到其他磁盘，释放 C 盘空间</p>
      </div>

      {/* 步骤指示器 */}
      <div className="flex items-center mb-8">
        {steps.map((s, idx) => (
          <div key={s.num} className="flex items-center">
            <div className={`
              w-9 h-9 rounded-full flex items-center justify-center text-sm font-bold
              ${step >= s.num
                ? 'bg-devman-accent/20 text-devman-accent border border-devman-accent/40'
                : 'bg-devman-panel-raised text-devman-text-muted border border-devman-border'
              }
            `}>
              {step > s.num ? '✓' : s.num}
            </div>
            <span className={`ml-2 text-sm font-medium ${step >= s.num ? 'text-devman-text-primary' : 'text-devman-text-muted'}`}>
              {s.label}
            </span>
            {idx < steps.length - 1 && (
              <div className={`w-16 h-0.5 mx-3 ${step > s.num ? 'bg-devman-accent/40' : 'bg-devman-border'}`} />
            )}
          </div>
        ))}
      </div>

      {/* Step 1: 选择环境 */}
      {step === 1 && (
        <div>
          <h3 className="text-lg font-bold text-devman-text-primary mb-4">选择要迁移的环境</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {envs.map((env) => (
              <Panel
                key={env.Env.Key}
                className={`p-5 cursor-pointer transition-all ${
                  selectedEnv?.Env.Key === env.Env.Key
                    ? 'border-devman-accent/50 bg-devman-accent/5'
                    : 'hover:border-devman-border-strong'
                }`}
                onClick={() => setSelectedEnv(env)}
              >
                <div className="flex items-center gap-3">
                  <span className="text-2xl">{env.Env.Icon}</span>
                  <div className="flex-1">
                    <h4 className="font-bold text-devman-text-primary">{env.Env.Name}</h4>
                    <p className="text-xs text-devman-text-muted">
                      {env.Instances[0]?.InstallPath || '未知路径'}
                    </p>
                  </div>
                  <span className="text-sm font-mono text-devman-info">{formatBytes(env.TotalSize)}</span>
                </div>
              </Panel>
            ))}
          </div>
          {envs.length === 0 && (
            <div className="text-center py-12 text-devman-text-muted">
              <p className="text-lg mb-2">🔍 暂无环境数据</p>
              <p className="text-sm">请先前往「总览」页面点击「刷新环境数据」</p>
            </div>
          )}
          <div className="mt-6 flex justify-end">
            <button
              disabled={!selectedEnv}
              onClick={() => setStep(2)}
              className="px-6 py-2.5 bg-devman-accent/15 text-devman-accent rounded-xl text-sm font-bold hover:bg-devman-accent/25 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
            >
              下一步 →
            </button>
          </div>
        </div>
      )}

      {/* Step 2: 选择目标 */}
      {step === 2 && selectedEnv && (
        <div>
          <h3 className="text-lg font-bold text-devman-text-primary mb-4">选择目标位置</h3>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-5 mb-6">
            <Panel className="p-5">
              <p className="text-sm text-devman-text-muted mb-2">已选环境</p>
              <div className="flex items-center gap-3 mb-3">
                <span className="text-2xl">{selectedEnv.Env.Icon}</span>
                <div>
                  <h4 className="font-bold text-devman-text-primary">{selectedEnv.Env.Name}</h4>
                  <p className="text-xs text-devman-text-muted">{selectedEnv.Instances[0]?.InstallPath}</p>
                </div>
              </div>
              <p className="text-2xl font-bold text-devman-info">{formatBytes(selectedEnv.TotalSize)}</p>
            </Panel>

            <Panel className="p-5">
              <p className="text-sm text-devman-text-muted mb-2">目标位置</p>
              <input
                type="text"
                value={targetDir}
                onChange={(e) => setTargetDir(e.target.value)}
                className="w-full px-4 py-2.5 bg-devman-panel-raised border border-devman-border rounded-xl text-sm text-devman-text-primary focus:outline-none focus:border-devman-accent/50 mb-3"
              />
              <p className="text-xs text-devman-text-muted">
                最终路径: {targetDir}\\{selectedEnv.Env.Name.toLowerCase().replace(/\./g, '')}
              </p>
            </Panel>
          </div>

          {/* 磁盘预览 */}
          <Panel className="p-5 mb-6">
            <p className="text-sm text-devman-text-muted mb-4">磁盘空间预览</p>
            <div className="space-y-3">
              {cDisk && (
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-devman-text-primary">C 盘</span>
                    <span className="text-devman-text-muted">
                      {Math.round(((cDisk.TotalBytes - cDisk.FreeBytes) / cDisk.TotalBytes) * 100)}% 已用
                    </span>
                  </div>
                  <div className="h-3 bg-devman-panel-raised rounded-full overflow-hidden">
                    <div
                      className="h-full rounded-full bg-yellow-500/60"
                      style={{ width: `${((cDisk.TotalBytes - cDisk.FreeBytes) / cDisk.TotalBytes) * 100}%` }}
                    />
                  </div>
                </div>
              )}
              {tDisk && (
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-devman-text-primary">{targetDisk} 盘</span>
                    <span className="text-devman-text-muted">
                      {Math.round(((tDisk.TotalBytes - tDisk.FreeBytes) / tDisk.TotalBytes) * 100)}% 已用
                    </span>
                  </div>
                  <div className="h-3 bg-devman-panel-raised rounded-full overflow-hidden">
                    <div
                      className="h-full rounded-full bg-green-500/60"
                      style={{ width: `${((tDisk.TotalBytes - tDisk.FreeBytes) / tDisk.TotalBytes) * 100}%` }}
                    />
                  </div>
                </div>
              )}
            </div>
            {selectedEnv && cDisk && (
              <p className="text-xs text-devman-text-muted mt-3">
                迁移后 C 盘将释放 {formatBytes(selectedEnv.TotalSize)} 空间
              </p>
            )}
          </Panel>

          <div className="flex justify-between">
            <button
              onClick={() => setStep(1)}
              className="px-6 py-2.5 bg-devman-panel-raised border border-devman-border rounded-xl text-sm text-devman-text-primary hover:bg-devman-border/20 transition-colors"
            >
              ← 上一步
            </button>
            <button
              onClick={() => setStep(3)}
              className="px-6 py-2.5 bg-devman-accent/15 text-devman-accent rounded-xl text-sm font-bold hover:bg-devman-accent/25 transition-colors"
            >
              下一步 →
            </button>
          </div>
        </div>
      )}

      {/* Step 3: 预览确认 */}
      {step === 3 && selectedEnv && (
        <div>
          <h3 className="text-lg font-bold text-devman-text-primary mb-4">确认迁移操作</h3>
          <Panel className="p-5 mb-6">
            <div className="space-y-3 text-sm">
              <div className="flex justify-between py-2 border-b border-devman-border/30">
                <span className="text-devman-text-muted">环境</span>
                <span className="text-devman-text-primary font-medium">{selectedEnv.Env.Name} {selectedEnv.Instances[0]?.Version}</span>
              </div>
              <div className="flex justify-between py-2 border-b border-devman-border/30">
                <span className="text-devman-text-muted">源路径</span>
                <span className="text-devman-text-primary font-medium">{selectedEnv.Instances[0]?.InstallPath}</span>
              </div>
              <div className="flex justify-between py-2 border-b border-devman-border/30">
                <span className="text-devman-text-muted">目标路径</span>
                <span className="text-devman-text-primary font-medium">{targetDir}\\{selectedEnv.Env.Name.toLowerCase().replace(/\./g, '')}</span>
              </div>
              <div className="flex justify-between py-2 border-b border-devman-border/30">
                <span className="text-devman-text-muted">迁移大小</span>
                <span className="text-devman-info font-bold">{formatBytes(selectedEnv.TotalSize)}</span>
              </div>
              <div className="flex justify-between py-2">
                <span className="text-devman-text-muted">策略</span>
                <span className="text-devman-text-primary font-medium">复制 → 验证 → 提交 → 删除源</span>
              </div>
            </div>
          </Panel>

          <div className="bg-yellow-500/10 border border-yellow-500/20 rounded-xl p-4 mb-6">
            <p className="text-sm text-yellow-400">
              ⚠️ 迁移前会自动创建配置快照。如果迁移失败，可以通过快照恢复原始状态。
              迁移过程中请勿关闭程序。
            </p>
          </div>

          <div className="flex justify-between">
            <button
              onClick={() => setStep(2)}
              className="px-6 py-2.5 bg-devman-panel-raised border border-devman-border rounded-xl text-sm text-devman-text-primary hover:bg-devman-border/20 transition-colors"
            >
              ← 上一步
            </button>
            <button
              onClick={() => { setStep(4); handleMigrate(); }}
              disabled={migrating}
              className="px-6 py-2.5 bg-devman-accent/15 text-devman-accent rounded-xl text-sm font-bold hover:bg-devman-accent/25 disabled:opacity-50 transition-colors"
            >
              {migrating ? '迁移中...' : '确认迁移'}
            </button>
          </div>
        </div>
      )}

      {/* Step 4: 执行结果 */}
      {step === 4 && (
        <div>
          <h3 className="text-lg font-bold text-devman-text-primary mb-4">迁移结果</h3>

          {migrating && (
            <Panel className="p-5 mb-6">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-5 h-5 border-2 border-devman-accent border-t-transparent rounded-full animate-spin" />
                <span className="text-devman-text-primary">正在迁移...</span>
              </div>
              <div className="h-2 bg-devman-panel-raised rounded-full overflow-hidden">
                <div className="h-full bg-devman-accent rounded-full animate-pulse w-3/4" />
              </div>
            </Panel>
          )}

          {!migrating && result && (
            <Panel className={`p-5 mb-6 ${result.Success ? 'border-green-500/30' : 'border-red-500/30'}`}>
              <div className="flex items-center gap-3 mb-3">
                <span className={`text-2xl ${result.Success ? 'text-green-400' : 'text-red-400'}`}>
                  {result.Success ? '✅' : '❌'}
                </span>
                <h4 className={`text-lg font-bold ${result.Success ? 'text-green-400' : 'text-red-400'}`}>
                  {result.Success ? '迁移成功' : '迁移失败'}
                </h4>
              </div>
              <p className="text-sm text-devman-text-primary mb-2">{result.Message}</p>
              {result.Success && (
                <div className="flex gap-6 text-sm mt-3">
                  <div>
                    <span className="text-devman-text-muted">移动大小: </span>
                    <span className="text-devman-info font-bold">{formatBytes(result.BytesMoved || 0)}</span>
                  </div>
                  <div>
                    <span className="text-devman-text-muted">耗时: </span>
                    <span className="text-devman-text-primary font-medium">{(result.DurationMs || 0) / 1000}s</span>
                  </div>
                </div>
              )}
            </Panel>
          )}

          {logs.length > 0 && (
            <Panel className="p-4 mb-6 bg-devman-bg-deep">
              <p className="text-xs text-devman-text-muted mb-2">操作日志</p>
              <div className="space-y-1 font-mono text-xs">
                {logs.map((log, i) => (
                  <p key={i} className="text-devman-text-muted">{log}</p>
                ))}
              </div>
            </Panel>
          )}

          <div className="flex justify-between">
            <button
              onClick={() => { setStep(1); setSelectedEnv(null); setResult(null); setLogs([]); }}
              className="px-6 py-2.5 bg-devman-panel-raised border border-devman-border rounded-xl text-sm text-devman-text-primary hover:bg-devman-border/20 transition-colors"
            >
              返回首页
            </button>
            {!migrating && result?.Success && (
              <button
                onClick={() => loadEnvs()}
                className="px-6 py-2.5 bg-devman-accent/15 text-devman-accent rounded-xl text-sm font-bold hover:bg-devman-accent/25 transition-colors"
              >
                🔄 刷新环境列表
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
