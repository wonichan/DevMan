import React, { createContext, useContext, useState, useCallback, useEffect, useRef, ReactNode } from 'react';
import { SurfaceCard } from './SurfaceCard';
import { Button } from './Button';

export interface ConfirmOptions {
  title: string;
  description: React.ReactNode;
  confirmText?: string;
  cancelText?: string;
  variant?: 'danger' | 'default';
}

interface ConfirmContextType {
  confirm: (options: ConfirmOptions) => Promise<boolean>;
}

const ConfirmContext = createContext<ConfirmContextType | undefined>(undefined);

export const ConfirmProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [options, setOptions] = useState<ConfirmOptions | null>(null);
  const resolverRef = useRef<((value: boolean) => void) | null>(null);
  const triggerRef = useRef<HTMLElement | null>(null);

  const confirm = useCallback((opts: ConfirmOptions) => {
    if (resolverRef.current) {
      resolverRef.current(false);
    }

    triggerRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    setOptions(opts);
    setIsOpen(true);
    return new Promise<boolean>((resolve) => {
      resolverRef.current = resolve;
    });
  }, []);

  const close = useCallback((value: boolean) => {
    setIsOpen(false);
    setOptions(null);

    if (resolverRef.current) {
      resolverRef.current(value);
      resolverRef.current = null;
    }

    triggerRef.current?.focus();
    triggerRef.current = null;
  }, []);

  const handleConfirm = useCallback(() => {
    close(true);
  }, [close]);

  const handleCancel = useCallback(() => {
    close(false);
  }, [close]);

  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        handleCancel();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleCancel, isOpen]);

  return (
    <ConfirmContext.Provider value={{ confirm }}>
      {children}
      {isOpen && options && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
          <SurfaceCard className="w-full max-w-md animate-in zoom-in-95 duration-200" variant="raised">
            <div role="dialog" aria-modal="true" aria-labelledby="devman-confirm-title">
            <h3 id="devman-confirm-title" className="text-xl font-bold text-slate-100 mb-2">{options.title}</h3>
            <div className="text-slate-300 text-sm mb-6">
              {options.description}
            </div>
            <div className="flex justify-end gap-3">
              <Button variant="ghost" onClick={handleCancel} autoFocus>
                {options.cancelText || 'Cancel'}
              </Button>
              <Button 
                variant={options.variant === 'danger' ? 'danger' : 'primary'} 
                onClick={handleConfirm}
              >
                {options.confirmText || 'Confirm'}
              </Button>
            </div>
            </div>
          </SurfaceCard>
        </div>
      )}
    </ConfirmContext.Provider>
  );
};

export const useConfirm = () => {
  const context = useContext(ConfirmContext);
  if (!context) {
    throw new Error('useConfirm must be used within a ConfirmProvider');
  }
  return context;
};
