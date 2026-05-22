import { useState } from 'react';
import Panel from '../components/Panel';

export default function Settings() {
  const [autoScan, setAutoScan] = useState(true);
  const [confirmDelete, setConfirmDelete] = useState(true);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-devman-text-primary">⚙️ 设置</h1>
        <p className="text-sm text-devman-text-muted mt-1">管理 DevMan 的偏好和配置</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        {/* 通用设置 */}
        <Panel className="p-5">
          <h3 className="font-bold text-devman-text-primary mb-4">通用设置</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-devman-text-primary">启动时自动扫描</p>
                <p className="text-xs text-devman-text-muted">打开应用时自动检测开发环境变化</p>
              </div>
              <button
                onClick={() => setAutoScan(!autoScan)}
                className={`w-11 h-6 rounded-full transition-colors ${autoScan ? 'bg-devman-accent' : 'bg-devman-border'}`}
              >
                <div className={`w-5 h-5 bg-white rounded-full shadow transition-transform ${autoScan ? 'translate-x-6' : 'translate-x-0.5'}`} />
              </button>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-devman-text-primary">迁移前确认</p>
                <p className="text-xs text-devman-text-muted">执行迁移前显示确认对话框</p>
              </div>
              <button
                onClick={() => setConfirmDelete(!confirmDelete)}
                className={`w-11 h-6 rounded-full transition-colors ${confirmDelete ? 'bg-devman-accent' : 'bg-devman-border'}`}
              >
                <div className={`w-5 h-5 bg-white rounded-full shadow transition-transform ${confirmDelete ? 'translate-x-6' : 'translate-x-0.5'}`} />
              </button>
            </div>
          </div>
        </Panel>

        {/* 数据管理 */}
        <Panel className="p-5">
          <h3 className="font-bold text-devman-text-primary mb-4">数据管理</h3>
          <div className="space-y-4">
            <div>
              <p className="text-sm font-medium text-devman-text-primary mb-1">数据库路径</p>
              <p className="text-xs text-devman-text-muted font-mono bg-devman-panel-raised rounded-lg px-3 py-2">.devman.db</p>
              <p className="text-xs text-devman-text-muted mt-1">便携模式：数据库存放在与程序同一目录下</p>
            </div>
            <div>
              <p className="text-sm font-medium text-devman-text-primary mb-1">历史记录</p>
              <p className="text-xs text-devman-text-muted mb-2">保留最近 100 条迁移和清理操作记录</p>
              <button className="px-4 py-2 bg-devman-panel-raised border border-devman-border rounded-xl text-xs text-devman-text-primary hover:bg-devman-border/20 transition-colors">
                导出历史记录
              </button>
            </div>
          </div>
        </Panel>

        {/* 关于 */}
        <Panel className="p-5 lg:col-span-2">
          <h3 className="font-bold text-devman-text-primary mb-4">关于 DevMan</h3>
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-devman-accent/15 rounded-2xl flex items-center justify-center text-xl">
              🧹
            </div>
            <div>
              <p className="text-sm font-bold text-devman-text-primary">DevMan v0.1.0</p>
              <p className="text-xs text-devman-text-muted">Windows 开发环境管理器</p>
              <p className="text-xs text-devman-text-muted mt-1">
                基于 Wails v2 + Go + React + Tailwind CSS 构建
              </p>
            </div>
          </div>
        </Panel>
      </div>
    </div>
  );
}
