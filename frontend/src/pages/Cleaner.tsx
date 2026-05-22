import { useEffect, useState } from 'react';
import { AnalyzeCleanable, CleanItems } from '../bindings/go/main/App';
import Panel from '../components/Panel';
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
  const [result, setResult] = useState<{ freed: number; count: number } | null>(null);

  const analyze = async () => {
    setScanning(true);
    setResult(null);
    try {
      const data = await AnalyzeCleanable();
      setItems(data || []);
    } catch (e) {
      console.error(e);
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
    setCleaning(true);
    setResult(null);
    try {
      const freed = await CleanItems(selected);
      setResult({ freed: Number(freed), count: selected.length });
      // Refresh list
      await analyze();
    } catch (e) {
      console.error(e);
    }
    setCleaning(false);
  };

  const totalSize = items.reduce((sum, i) => sum + (i.SizeBytes || 0), 0);
  const selectedSize = items.filter(i => i.Selected).reduce((sum, i) => sum + (i.SizeBytes || 0), 0);
  const selectedCount = items.filter(i => i.Selected).length;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-devman-text-primary">🧹 缓存清理</h1>
          <p className="text-sm text-devman-text-muted mt-1">安全清理各类开发环境缓存，释放磁盘空间</p>
        </div>
        <button
          onClick={analyze}
          disabled={scanning}
          className="px-4 py-2.5 bg-devman-accent/15 text-devman-accent rounded-xl text-sm font-bold hover:bg-devman-accent/25 disabled:opacity-50 transition-colors"
        >
          {scanning ? '分析中...' : '🔍 分析可清理项'}
        </button>
      </div>

      {/* 汇总卡片 */}
      <Panel className="p-5 mb-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-devman-text-muted mb-1">可释放空间</p>
            <p className="text-3xl font-bold text-devman-accent">{formatBytes(totalSize)}</p>
            <p className="text-xs text-devman-text-muted mt-1">共 {items.length} 项可清理</p>
          </div>
          <button
            onClick={clean}
            disabled={cleaning || selectedCount === 0}
            className="px-6 py-3 bg-devman-accent/15 text-devman-accent rounded-xl text-sm font-bold hover:bg-devman-accent/25 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          >
            {cleaning ? '清理中...' : `🚮 一键清理 (${selectedCount})`}
          </button>
        </div>
        {selectedCount > 0 && (
          <p className="text-xs text-devman-text-muted mt-3">
            已选择 {selectedCount} 项，共 {formatBytes(selectedSize)}
          </p>
        )}
      </Panel>

      {/* 结果提示 */}
      {result && (
        <div className="mb-6 bg-green-500/10 border border-green-500/20 rounded-xl p-4">
          <p className="text-sm text-green-400">
            ✅ 已清理 {result.count} 项，释放空间 {formatBytes(result.freed)}
          </p>
        </div>
      )}

      {/* 列表 */}
      {items.length > 0 && (
        <div className="mb-4">
          <label className="flex items-center gap-2 text-sm text-devman-text-muted cursor-pointer">
            <input
              type="checkbox"
              checked={items.every(i => i.Selected)}
              onChange={(e) => selectAll(e.target.checked)}
              className="w-4 h-4 rounded border-devman-border bg-devman-panel-raised text-devman-accent focus:ring-devman-accent/50"
            />
            全选 / 取消全选
          </label>
        </div>
      )}

      <div className="space-y-3">
        {items.map((item, idx) => (
          <Panel key={idx} className="p-4 flex items-center gap-4">
            <input
              type="checkbox"
              checked={item.Selected}
              onChange={() => toggleItem(idx)}
              className="w-4 h-4 rounded border-devman-border bg-devman-panel-raised text-devman-accent focus:ring-devman-accent/50 flex-shrink-0"
            />
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-sm font-bold text-devman-text-primary">{item.Name}</span>
              </div>
              <p className="text-xs text-devman-text-muted truncate">{item.Path}</p>
              <p className="text-xs text-devman-text-muted/60">{item.Description}</p>
            </div>
            <span className="text-sm font-bold text-devman-info flex-shrink-0">{formatBytes(item.SizeBytes)}</span>
          </Panel>
        ))}
      </div>

      {items.length === 0 && !scanning && (
        <div className="text-center py-20 text-devman-text-muted">
          <p className="text-lg mb-2">🔍 点击「分析可清理项」开始扫描</p>
          <p className="text-sm">系统会检查各类开发环境的缓存目录</p>
        </div>
      )}

      <p className="text-xs text-devman-text-muted/60 mt-6">
        ⚠️ 只有安全可恢复的缓存会被列出。清理前会自动创建快照。
      </p>
    </div>
  );
}
