import { useState, useEffect, useCallback } from 'react';
import Sidebar from './components/Sidebar';
import GlobalSearch from './components/GlobalSearch';
import Dashboard from './pages/Dashboard';
import Environments from './pages/Environments';
import Migration from './pages/Migration';
import Cleaner from './pages/Cleaner';
import Versions from './pages/Versions';
import Settings from './pages/Settings';
import { ToastProvider } from './components/ui/ToastProvider';
import { ConfirmProvider } from './components/ui/ConfirmDialog';
import { PageActionProvider, usePageActionRegistry } from './hooks/usePageActions';
import { GetSettings } from './api/app';

export type Page = 'dashboard' | 'environments' | 'migration' | 'cleaner' | 'versions' | 'settings';

const pageOrder: Page[] = ['dashboard', 'environments', 'migration', 'cleaner', 'versions', 'settings'];

function isTypingTarget(el: EventTarget | null): boolean {
  if (!(el instanceof HTMLElement)) return false;
  const tag = el.tagName.toLowerCase();
  if (tag === 'input' || tag === 'textarea' || tag === 'select') return true;
  if (el.isContentEditable) return true;
  return false;
}

function AppContent() {
  const [currentPage, setCurrentPage] = useState<Page>('dashboard');
  const { invoke } = usePageActionRegistry();

  const handleRefresh = useCallback(() => {
    invoke(currentPage, 'refresh');
  }, [currentPage, invoke]);

  useEffect(() => {
    GetSettings()
      .then((s) => {
        document.documentElement.dataset.theme = s.Theme;
      })
      .catch(() => {
        document.documentElement.dataset.theme = 'dark';
      });
  }, []);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!e.ctrlKey && !e.metaKey) return;

      if (e.key === 'k') {
        return;
      }

      if (e.key === 'r') {
        if (isTypingTarget(e.target)) return;
        e.preventDefault();
        handleRefresh();
        return;
      }

      const digit = parseInt(e.key, 10);
      if (!Number.isNaN(digit) && digit >= 1 && digit <= 6) {
        if (isTypingTarget(e.target)) return;
        e.preventDefault();
        setCurrentPage(pageOrder[digit - 1]);
        return;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleRefresh]);

  return (
    <div className="h-screen w-screen flex app-atmosphere overflow-hidden">
      <Sidebar active={currentPage} onNavigate={setCurrentPage} />
      <main className="flex-1 overflow-y-auto">
        <div className="max-w-5xl mx-auto p-8 pb-24">
          <div className="flex justify-end mb-4">
            <GlobalSearch onNavigate={setCurrentPage} />
          </div>
          {currentPage === 'dashboard' && <Dashboard />}
          {currentPage === 'environments' && <Environments />}
          {currentPage === 'migration' && <Migration />}
          {currentPage === 'cleaner' && <Cleaner />}
          {currentPage === 'versions' && <Versions />}
          {currentPage === 'settings' && <Settings />}
        </div>
      </main>
    </div>
  );
}

function App() {
  return (
    <ToastProvider>
      <ConfirmProvider>
        <PageActionProvider>
          <AppContent />
        </PageActionProvider>
      </ConfirmProvider>
    </ToastProvider>
  );
}

export default App;
