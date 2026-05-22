import { useState } from 'react';
import Sidebar from './components/Sidebar';
import Dashboard from './pages/Dashboard';
import Environments from './pages/Environments';
import Migration from './pages/Migration';
import Cleaner from './pages/Cleaner';
import Versions from './pages/Versions';
import Settings from './pages/Settings';

type Page = 'dashboard' | 'environments' | 'migration' | 'cleaner' | 'versions' | 'settings';

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('dashboard');

  return (
    <div className="h-screen w-screen flex bg-devman-bg text-devman-text-primary overflow-hidden">
      <Sidebar active={currentPage} onNavigate={setCurrentPage} />
      <main className="flex-1 p-6 overflow-y-auto">
        {currentPage === 'dashboard' && <Dashboard />}
        {currentPage === 'environments' && <Environments />}
        {currentPage === 'migration' && <Migration />}
        {currentPage === 'cleaner' && <Cleaner />}
        {currentPage === 'versions' && <Versions />}
        {currentPage === 'settings' && <Settings />}
      </main>
    </div>
  );
}

export default App;
