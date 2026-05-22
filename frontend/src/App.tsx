import { useState, useEffect } from 'react';
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
import { GetSettings } from './api/app';

export type Page = 'dashboard' | 'environments' | 'migration' | 'cleaner' | 'versions' | 'settings';

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('dashboard');

  useEffect(() => {
    GetSettings()
      .then((s) => {
        document.documentElement.dataset.theme = s.Theme;
      })
      .catch(() => {
        document.documentElement.dataset.theme = 'dark';
      });
  }, []);

  return (
    <ToastProvider>
      <ConfirmProvider>
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
      </ConfirmProvider>
    </ToastProvider>
  );
}

export default App;
