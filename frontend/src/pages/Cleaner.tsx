import { useState } from 'react';
import { AnalyzeCleanable, CleanItems } from '../bindings/go/main/App';
import { PageHeader } from '../components/ui/PageHeader';
import { SurfaceCard } from '../components/ui/SurfaceCard';
import { Button } from '../components/ui/Button';
import { EmptyState } from '../components/ui/EmptyState';
import { useConfirm } from '../hooks/useConfirm';
import { useToast } from '../hooks/useToast';
import { SearchIcon, TrashIcon, WarningIcon } from '../components/icons';
import type { CleanableItem } from '../devman-types';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

export default function Cleaner() {
  const [items, setItems] = useState<CleanableItem[]>([]);
  const [scanning, setScanning] = useState(false);
  const [cleaning, setCleaning] = useState(false);
  
  const { confirm } = useConfirm();
  const toast = useToast();

  const analyze = async () => {
    setScanning(true);
    try {
      const data = await AnalyzeCleanable();
      setItems(data || []);
      toast.success('分析完成', `找到 ${data?.length || 0} 个可清理项`);
    } catch (e) {
      console.error(e);
      toast.error('分析失败', e instanceof Error ? e.message : String(e));
    }
    setScanning(false);
  };

  const toggleItem = (idx: number) => {
    setItems(prev => prev.map((item, i) =>
      i === idx ? { ...item, Selected: !item.Selected } : item
    ));
  };

  const selectAll = (checked: boolean) => {
    setItems(prev => prev.map(item => ({ ...item, Selected: checked })));
  };

  const clean = async () => {
    const selected = items.filter(i => i.Selected);
    if (selected.length === 0) return;
    
    const selectedSize = selected.reduce((sum, i) => sum + (i.SizeBytes || 0), 0);
    
    const isConfirmed = await confirm({
      title: '确认清理',
      description: `即将清理 ${selected.length} 个项目，释放 ${formatBytes(selectedSize)} 空间。此操作不可恢复。`,
      confirmText: '立即清理',
      cancelText: '取消',
      variant: 'danger'
    });

    if (!isConfirmed) return;

    setCleaning(true);
    try {
      const freed = await CleanItems(selected);
      toast.success('清理完成', `成功清理 ${selected.length} 项，释放 ${formatBytes(Number(freed))} 空间`);
      await analyze();
    } catch (e) {
      console.error(e);
      toast.error('清理失败', e instanceof Error ? e.message : String(e));
    }
    setCleaning(false);
  };

  const totalSize = items.reduce((sum, i) => sum + (i.SizeBytes || 0), 0);
  const selectedSize = items.filter(i => i.Selected).reduce((sum, i) => sum + (i.SizeBytes || 0), 0);
  const selectedCount = items.filter(i => i.Selected).length;

  return (
    <div>
      <PageHeader 
        title="缓存清理" 
        description="安全清理各类开发环境缓存，释放磁盘空间"
        actions={
          <Button 
            variant="primary" 
            onClick={analyze} 
            isLoading={scanning}
          >
            <SearchIcon className="mr-2 h-4 w-4" />
            分析可清理项
          </Button>
        }
      />

      {/* 汇总卡片 */}
      <SurfaceCard className="mb-6" variant="raised">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-slate-400 mb-1">可释放空间</p>
            <p className="text-3xl font-bold text-emerald-400">{formatBytes(totalSize)}</p>
            <p className="text-xs text-slate-400 mt-1">共 {items.length} 项可清理</p>
          </div>
          <Button
            variant="danger"
            onClick={clean}
            disabled={selectedCount === 0}
            isLoading={cleaning}
          >
            <TrashIcon className="mr-2 h-4 w-4" />
            一键清理 ({selectedCount})
          </Button>
        </div>
        {selectedCount > 0 && (
          <p className="text-xs text-slate-400 mt-3">
            已选择 {selectedCount} 项，共 {formatBytes(selectedSize)}
          </p>
        )}
      </SurfaceCard>

      {/* 列表 */}
      {items.length > 0 && (
        <div className="mb-4 pl-1">
          <label className="flex items-center gap-2 text-sm text-slate-400 cursor-pointer hover:text-slate-200 transition-colors w-fit">
            <input
              type="checkbox"
              checked={items.every(i => i.Selected)}
              onChange={(e) => selectAll(e.target.checked)}
              className="w-4 h-4 rounded border-[#475569] bg-[#0f172a] text-emerald-500 focus:ring-emerald-500/50"
            />
            全选 / 取消全选
          </label>
        </div>
      )}

      <div className="space-y-3">
        {items.map((item, idx) => (
          <SurfaceCard key={idx} className="flex items-center gap-4 py-4">
            <input
              type="checkbox"
              checked={item.Selected}
              onChange={() => toggleItem(idx)}
              className="w-4 h-4 rounded border-[#475569] bg-[#0f172a] text-emerald-500 focus:ring-emerald-500/50 flex-shrink-0 cursor-pointer"
            />
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-sm font-bold text-slate-200">{item.Name}</span>
              </div>
              <p className="text-xs text-slate-400 truncate mt-0.5 font-mono">{item.Path}</p>
              <p className="text-xs text-slate-500 mt-1">{item.Description}</p>
            </div>
            <span className="text-sm font-bold text-cyan-400 flex-shrink-0">{formatBytes(item.SizeBytes)}</span>
          </SurfaceCard>
        ))}
      </div>

      {items.length === 0 && !scanning && (
        <EmptyState 
          icon={<SearchIcon className="w-6 h-6" />}
          title="点击「分析可清理项」开始扫描"
          description="系统会检查各类开发环境的缓存目录"
        />
      )}

      <p className="text-xs text-slate-500 mt-6 text-center">
        <WarningIcon className="mr-1 inline h-3.5 w-3.5 align-[-2px]" />
        只有安全可恢复的缓存会被列出。清理前会自动创建快照。
      </p>
    </div>
  );
}
