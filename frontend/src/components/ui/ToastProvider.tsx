import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import { Toast, ToastProps } from './Toast';

type OmitId<T> = Omit<T, 'id' | 'onDismiss'>;

interface ToastContextType {
  toast: (props: OmitId<ToastProps>) => void;
  success: (title: string, message?: React.ReactNode) => void;
  error: (title: string, message?: React.ReactNode) => void;
  info: (title: string, message?: React.ReactNode) => void;
  warning: (title: string, message?: React.ReactNode) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export const ToastProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [toasts, setToasts] = useState<ToastProps[]>([]);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const addToast = useCallback((props: OmitId<ToastProps>) => {
    const id = Math.random().toString(36).substring(2, 9);
    setToasts((prev) => [...prev, { ...props, id, onDismiss: removeToast }]);
  }, [removeToast]);

  const success = useCallback((title: string, message?: React.ReactNode) => {
    addToast({ title, message, variant: 'success' });
  }, [addToast]);

  const error = useCallback((title: string, message?: React.ReactNode) => {
    addToast({ title, message, variant: 'error' });
  }, [addToast]);

  const info = useCallback((title: string, message?: React.ReactNode) => {
    addToast({ title, message, variant: 'info' });
  }, [addToast]);

  const warning = useCallback((title: string, message?: React.ReactNode) => {
    addToast({ title, message, variant: 'warning' });
  }, [addToast]);

  return (
    <ToastContext.Provider value={{ toast: addToast, success, error, info, warning }}>
      {children}
      <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2 pointer-events-none w-full max-w-sm">
        {toasts.map((t) => (
          <Toast key={t.id} {...t} />
        ))}
      </div>
    </ToastContext.Provider>
  );
};

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};
