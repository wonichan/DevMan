import React, { useEffect } from 'react';
import { CheckIcon, WarningIcon, InfoIcon, CloseIcon } from '../icons';

export type ToastVariant = 'success' | 'error' | 'info' | 'warning';

export interface ToastProps {
  id: string;
  title: string;
  message?: React.ReactNode;
  variant?: ToastVariant;
  duration?: number;
  onDismiss: (id: string) => void;
}

export const Toast: React.FC<ToastProps> = ({
  id,
  title,
  message,
  variant = 'info',
  duration = 5000,
  onDismiss
}) => {
  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        onDismiss(id);
      }, duration);
      return () => clearTimeout(timer);
    }
  }, [id, duration, onDismiss]);

  let icon = null;
  let variantClasses = '';

  switch (variant) {
    case 'success':
      icon = <CheckIcon className="text-emerald-400" width={20} height={20} />;
      variantClasses = 'border-emerald-500/30 bg-emerald-950/20';
      break;
    case 'error':
      icon = <WarningIcon className="text-red-400" width={20} height={20} />;
      variantClasses = 'border-red-500/30 bg-red-950/20';
      break;
    case 'warning':
      icon = <WarningIcon className="text-amber-400" width={20} height={20} />;
      variantClasses = 'border-amber-500/30 bg-amber-950/20';
      break;
    case 'info':
    default:
      icon = <InfoIcon className="text-cyan-400" width={20} height={20} />;
      variantClasses = 'border-cyan-500/30 bg-cyan-950/20';
      break;
  }

  return (
    <div className={`pointer-events-auto flex w-full max-w-md rounded-lg shadow-lg border backdrop-blur-md overflow-hidden ${variantClasses} transition-all`}>
      <div className="flex-shrink-0 flex items-start justify-center p-4">
        {icon}
      </div>
      <div className="ml-2 flex-1 pt-4 pb-4 pr-2">
        <p className="text-sm font-medium text-slate-100">{title}</p>
        {message && <p className="mt-1 text-sm text-slate-300">{message}</p>}
      </div>
      <div className="flex-shrink-0 flex">
        <button
          onClick={() => onDismiss(id)}
          className="inline-flex rounded-md p-2 text-slate-400 hover:text-slate-200 focus:outline-none"
        >
          <span className="sr-only">Close</span>
          <CloseIcon width={20} height={20} />
        </button>
      </div>
    </div>
  );
};
