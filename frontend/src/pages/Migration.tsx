import { useEffect, useState } from 'react';
import { GetEnvs, GetEnvSummary, Migrate, GetDiskInfo } from '../bindings/go/main/App';
import type { EnvSummary, DiskInfo } from '../devman-types';
import { PageHeader } from '../components/ui/PageHeader';
import { Button } from '../components/ui/Button';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { ProgressBar } from '../components/ui/ProgressBar';
import { useToast } from '../hooks/useToast';
import { useConfirm } from '../hooks/useConfirm';
import { 
  CheckIcon, 
  CloseIcon, 
  RefreshIcon, 
  WarningIcon, 
  InfoIcon, 
  ArrowRightIcon, 
  ArrowLeftIcon,
  MigrationIcon
} from '../components/icons';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

interface MigrationResult {
  success: boolean;
  message: string;
  bytesMoved?: number;
  durationMs?: number;
}

export default function Migration() {
  const [step, setStep] = useState(1);
  const [envs, setEnvs] = useState<EnvSummary[]>([]);
  const [selectedEnv, setSelectedEnv] = useState<EnvSummary | null>(null);
  const [targetDir, setTargetDir] = useState('D:\\Dev');
  const [disks, setDisks] = useState<DiskInfo[]>([]);
  const [migrating, setMigrating] = useState(false);
  const [result, setResult] = useState<MigrationResult | null>(null);
  const [logs, setLogs] = useState<string[]>([]);

  const { success, error } = useToast();
  const { confirm } = useConfirm();

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

  const startMigration = async () => {
    if (!selectedEnv) return;
    
    const isConfirmed = await confirm({
      title: '确认执行迁移',
      description: `即将迁移 ${selectedEnv.Env.Name} 到 ${targetDir}。此操作会移动文件并更新环境变量。`,
      confirmText: '立即迁移',
      cancelText: '取消',
      variant: 'danger'
    });

    if (!isConfirmed) return;

    setStep(4);
    setMigrating(true);
    setLogs(['开始迁移...']);
    
    try {
      const res = await Migrate(selectedEnv.Env.Id, targetDir, false) as MigrationResult;
      setResult(res);
      setLogs(prev => [...prev, res.message || '迁移完成']);
      
      if (res.success) {
        success('迁移成功', res.message);
      } else {
        error('迁移失败', res.message);
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      setLogs(prev => [...prev, '错误: ' + msg]);
      error('迁移错误', msg);
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
    <div className="animate-in fade-in duration-300">
      <PageHeader 
        title="迁移向导" 
        description="将开发环境安全迁移到其他磁盘，释放 C 盘空间"
      />

      {/* 步骤指示器 */}
      <div className="flex items-center mb-8 bg-[#1e293b]/40 p-4 rounded-xl border border-[#334155]">
        {steps.map((s, idx) => (
          <div key={s.num} className="flex items-center">
            <div className={`
              w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold transition-colors
              ${step > s.num
                ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40'
                : step === s.num
                  ? 'bg-emerald-600 text-white shadow-[0_0_10px_rgba(16,185,129,0.3)]'
                  : 'bg-[#0f172a] text-slate-500 border border-[#334155]'
              }
            `}>
              {step > s.num ? <CheckIcon className="w-4 h-4" /> : s.num}
            </div>
            <span className={`ml-3 text-sm font-medium ${
              step >= s.num ? 'text-slate-200' : 'text-slate-500'
            }`}>
              {s.label}
            </span>
            {idx < steps.length - 1 && (
              <div className={`w-12 h-[2px] mx-4 transition-colors ${
                step > s.num ? 'bg-emerald-500/40' : 'bg-[#334155]'
              }`} />
            )}
          </div>
        ))}
      </div>

      {/* Step 1: 选择环境 */}
      {step === 1 && (
        <div className="animate-in fade-in slide-in-from-right-4 duration-300">
          <h3 className="text-lg font-bold text-slate-200 mb-4">选择要迁移的环境</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {envs.map((env) => (
              <SurfaceCard
                key={env.Env.Key}
                variant={selectedEnv?.Env.Key === env.Env.Key ? 'selected' : 'interactive'}
                className="p-5"
                role="button"
                tabIndex={0}
                onClick={() => setSelectedEnv(env)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' || event.key === ' ') {
                    event.preventDefault();
                    setSelectedEnv(env);
                  }
                }}
              >
                <div className="flex items-center gap-4">
                  <span className="text-3xl opacity-80">{env.Env.Icon}</span>
                  <div className="flex-1 min-w-0">
                    <h4 className="font-bold text-slate-100 truncate">{env.Env.Name}</h4>
                    <p className="text-xs text-slate-400 truncate mt-1">
                      {env.Instances[0]?.InstallPath || '未知路径'}
                    </p>
                  </div>
                  <span className="text-sm font-mono text-cyan-400 bg-cyan-400/10 px-2 py-1 rounded-md">
                    {formatBytes(env.TotalSize)}
                  </span>
                </div>
              </SurfaceCard>
            ))}
          </div>
          
          {envs.length === 0 && (
            <SurfaceCard className="text-center py-12 flex flex-col items-center">
              <MigrationIcon className="w-12 h-12 text-slate-500 mb-4 opacity-50" />
              <p className="text-lg font-medium text-slate-300 mb-2">暂无环境数据</p>
              <p className="text-sm text-slate-400">请先前往「总览」页面扫描环境数据</p>
            </SurfaceCard>
          )}
          
          <div className="mt-8 flex justify-end">
            <Button
              variant="primary"
              disabled={!selectedEnv}
              onClick={() => setStep(2)}
              className="flex items-center gap-2"
            >
              下一步
              <ArrowRightIcon className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}

      {/* Step 2: 选择目标 */}
      {step === 2 && selectedEnv && (
        <div className="animate-in fade-in slide-in-from-right-4 duration-300">
          <h3 className="text-lg font-bold text-slate-200 mb-4">选择目标位置</h3>
          
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-5 mb-6">
            <SurfaceCard className="p-6">
              <p className="text-sm text-slate-400 mb-4 flex items-center gap-2">
                <InfoIcon className="w-4 h-4" />
                已选环境
              </p>
              <div className="flex items-center gap-4 mb-4 bg-[#0f172a] p-4 rounded-xl border border-[#334155]">
                <span className="text-3xl">{selectedEnv.Env.Icon}</span>
                <div className="flex-1 min-w-0">
                  <h4 className="font-bold text-slate-100 truncate">{selectedEnv.Env.Name}</h4>
                  <p className="text-xs text-slate-400 truncate mt-1">{selectedEnv.Instances[0]?.InstallPath}</p>
                </div>
              </div>
              <div className="flex items-end gap-2">
                <p className="text-3xl font-bold text-cyan-400">{formatBytes(selectedEnv.TotalSize)}</p>
                <span className="text-sm text-slate-500 mb-1">待迁移</span>
              </div>
            </SurfaceCard>

            <SurfaceCard className="p-6">
              <p className="text-sm text-slate-400 mb-4 flex items-center gap-2">
                <MigrationIcon className="w-4 h-4" />
                目标位置
              </p>
              <input
                type="text"
                value={targetDir}
                onChange={(e) => setTargetDir(e.target.value)}
                className="w-full px-4 py-3 bg-[#0f172a] border border-[#334155] rounded-xl text-sm text-slate-200 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/20 mb-4 transition-all"
              />
              <div className="bg-emerald-500/10 border border-emerald-500/20 p-3 rounded-lg flex items-start gap-2">
                <InfoIcon className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                <p className="text-xs text-emerald-300/80 leading-relaxed">
                  最终路径: <span className="text-emerald-300 font-mono bg-emerald-500/20 px-1 rounded">{targetDir}\\{selectedEnv.Env.Name.toLowerCase().replace(/\./g, '')}</span>
                </p>
              </div>
            </SurfaceCard>
          </div>

          {/* 磁盘预览 */}
          <SurfaceCard className="p-6 mb-8">
            <p className="text-sm text-slate-400 mb-5 flex items-center gap-2">
              <InfoIcon className="w-4 h-4" />
              磁盘空间预览
            </p>
            <div className="space-y-6">
              {cDisk && (
                <div>
                  <div className="flex justify-between text-sm mb-2">
                    <span className="text-slate-200 font-medium">C 盘 (源)</span>
                    <span className="text-slate-400">
                      {Math.round(((cDisk.TotalBytes - cDisk.FreeBytes) / cDisk.TotalBytes) * 100)}% 已用
                    </span>
                  </div>
                  <ProgressBar 
                    value={((cDisk.TotalBytes - cDisk.FreeBytes) / cDisk.TotalBytes) * 100} 
                    variant="warning" 
                  />
                </div>
              )}
              {tDisk && (
                <div>
                  <div className="flex justify-between text-sm mb-2">
                    <span className="text-slate-200 font-medium">{targetDisk} 盘 (目标)</span>
                    <span className="text-slate-400">
                      {Math.round(((tDisk.TotalBytes - tDisk.FreeBytes) / tDisk.TotalBytes) * 100)}% 已用
                    </span>
                  </div>
                  <ProgressBar 
                    value={((tDisk.TotalBytes - tDisk.FreeBytes) / tDisk.TotalBytes) * 100} 
                    variant="accent" 
                  />
                </div>
              )}
            </div>
            {selectedEnv && cDisk && (
              <div className="mt-5 pt-4 border-t border-[#334155] flex items-center gap-2 text-sm text-slate-400">
                <CheckIcon className="w-4 h-4 text-emerald-400" />
                迁移后 C 盘将释放 <strong className="text-emerald-400">{formatBytes(selectedEnv.TotalSize)}</strong> 空间
              </div>
            )}
          </SurfaceCard>

          <div className="flex justify-between">
            <Button variant="secondary" onClick={() => setStep(1)} className="flex items-center gap-2">
              <ArrowLeftIcon className="w-4 h-4" />
              上一步
            </Button>
            <Button variant="primary" onClick={() => setStep(3)} className="flex items-center gap-2">
              下一步
              <ArrowRightIcon className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}

      {/* Step 3: 预览确认 */}
      {step === 3 && selectedEnv && (
        <div className="animate-in fade-in slide-in-from-right-4 duration-300">
          <h3 className="text-lg font-bold text-slate-200 mb-4">确认迁移操作</h3>
          
          <SurfaceCard className="p-0 overflow-hidden mb-6 border-[#334155]">
            <div className="grid grid-cols-3 divide-x divide-[#334155] bg-[#0f172a]">
              <div className="p-4 flex flex-col items-center justify-center text-center">
                <span className="text-xs text-slate-500 uppercase tracking-wider mb-1">环境</span>
                <span className="text-sm font-bold text-slate-200">{selectedEnv.Env.Name} {selectedEnv.Instances[0]?.Version}</span>
              </div>
              <div className="p-4 flex flex-col items-center justify-center text-center">
                <span className="text-xs text-slate-500 uppercase tracking-wider mb-1">迁移大小</span>
                <span className="text-sm font-bold text-cyan-400">{formatBytes(selectedEnv.TotalSize)}</span>
              </div>
              <div className="p-4 flex flex-col items-center justify-center text-center">
                <span className="text-xs text-slate-500 uppercase tracking-wider mb-1">策略</span>
                <span className="text-sm font-bold text-emerald-400">验证 → 提交</span>
              </div>
            </div>
            
            <div className="p-6 space-y-4">
              <div className="flex flex-col gap-1">
                <span className="text-xs text-slate-500 uppercase tracking-wider">源路径</span>
                <span className="text-sm text-slate-200 font-mono bg-[#0f172a] p-2 rounded border border-[#334155]">{selectedEnv.Instances[0]?.InstallPath}</span>
              </div>
              
              <div className="flex justify-center text-slate-500 py-1">
                <ArrowRightIcon className="w-5 h-5 rotate-90" />
              </div>
              
              <div className="flex flex-col gap-1">
                <span className="text-xs text-slate-500 uppercase tracking-wider">目标路径</span>
                <span className="text-sm text-emerald-300 font-mono bg-emerald-500/10 p-2 rounded border border-emerald-500/20">{targetDir}\\{selectedEnv.Env.Name.toLowerCase().replace(/\./g, '')}</span>
              </div>
            </div>
          </SurfaceCard>

          <div className="bg-amber-500/10 border border-amber-500/20 rounded-xl p-4 mb-8 flex gap-3 items-start">
            <WarningIcon className="w-5 h-5 text-amber-400 shrink-0 mt-0.5" />
            <p className="text-sm text-amber-300/90 leading-relaxed">
              <strong>不可中断操作：</strong> 迁移前会自动创建配置快照。如果迁移失败，可以通过快照恢复原始状态。
              迁移过程中请勿关闭程序或切断电源。
            </p>
          </div>

          <div className="flex justify-between">
            <Button variant="secondary" onClick={() => setStep(2)} className="flex items-center gap-2">
              <ArrowLeftIcon className="w-4 h-4" />
              上一步
            </Button>
            <Button 
              variant="primary" 
              onClick={startMigration} 
              disabled={migrating}
              className="flex items-center gap-2 px-8"
            >
              确认并开始迁移
              <CheckIcon className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}

      {/* Step 4: 执行结果 */}
      {step === 4 && (
        <div className="animate-in fade-in zoom-in-95 duration-300">
          <h3 className="text-lg font-bold text-slate-200 mb-4">迁移结果</h3>

          {migrating && (
            <SurfaceCard className="p-8 mb-6 text-center">
              <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-emerald-500/10 mb-4">
                <div className="w-8 h-8 border-3 border-emerald-500 border-t-transparent rounded-full animate-spin" />
              </div>
              <h4 className="text-lg font-bold text-slate-200 mb-2">正在执行环境迁移</h4>
              <p className="text-sm text-slate-400 mb-6">这可能需要几分钟时间，请勿关闭程序...</p>
              
              <div className="w-3/4 mx-auto max-w-md h-2 bg-[#0f172a] rounded-full overflow-hidden border border-[#334155]">
                <div className="h-full bg-emerald-500 rounded-full animate-pulse w-3/4 shadow-[0_0_10px_rgba(16,185,129,0.5)]" />
              </div>
            </SurfaceCard>
          )}

          {!migrating && result && (
            <SurfaceCard className={`p-6 mb-6 border ${result.success ? 'border-emerald-500/50 bg-emerald-500/5' : 'border-red-500/50 bg-red-500/5'}`}>
              <div className="flex items-start gap-4">
                <div className={`p-3 rounded-full ${result.success ? 'bg-emerald-500/20 text-emerald-400' : 'bg-red-500/20 text-red-400'}`}>
                  {result.success ? <CheckIcon className="w-6 h-6" /> : <CloseIcon className="w-6 h-6" />}
                </div>
                <div>
                  <h4 className={`text-lg font-bold mb-1 ${result.success ? 'text-emerald-400' : 'text-red-400'}`}>
                    {result.success ? '迁移成功' : '迁移失败'}
                  </h4>
                  <p className="text-sm text-slate-300 mb-4">{result.message}</p>
                  
                  {result.success && (
                    <div className="flex flex-wrap gap-4 text-sm bg-[#0f172a]/50 p-3 rounded-lg border border-[#334155]/50">
                      <div className="flex items-center gap-2">
                        <span className="text-slate-500">移动大小:</span>
                        <span className="text-cyan-400 font-bold">{formatBytes(result.bytesMoved || 0)}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="text-slate-500">耗时:</span>
                        <span className="text-slate-200 font-medium">{(result.durationMs || 0) / 1000}s</span>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </SurfaceCard>
          )}

          {logs.length > 0 && (
            <SurfaceCard className="p-0 mb-8 overflow-hidden border-[#334155]">
              <div className="bg-[#0f172a] px-4 py-2 border-b border-[#334155] flex items-center gap-2">
                <InfoIcon className="w-4 h-4 text-slate-400" />
                <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">执行日志</span>
              </div>
              <div className="p-4 bg-[#1e293b]/50 max-h-60 overflow-y-auto font-mono text-xs space-y-1.5">
                {logs.map((log, i) => (
                  <div key={i} className="flex gap-3 text-slate-300">
                    <span className="text-slate-500 select-none">[{String(i + 1).padStart(2, '0')}]</span>
                    <span className={log.includes('错误') ? 'text-red-400' : log.includes('成功') || log.includes('完成') ? 'text-emerald-400' : ''}>
                      {log}
                    </span>
                  </div>
                ))}
              </div>
            </SurfaceCard>
          )}

          <div className="flex justify-between">
            <Button
              variant="secondary"
              onClick={() => { setStep(1); setSelectedEnv(null); setResult(null); setLogs([]); }}
            >
              返回重新选择
            </Button>
            {!migrating && result?.success && (
              <Button
                variant="primary"
                onClick={() => {
                  loadEnvs();
                  setStep(1); 
                  setSelectedEnv(null); 
                  setResult(null); 
                  setLogs([]);
                }}
                className="flex items-center gap-2"
              >
                <RefreshIcon className="w-4 h-4" />
                完成并刷新
              </Button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
