import React from 'react';

export interface StatusBadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  status: 'healthy' | 'warning' | 'danger' | 'info' | 'active' | 'default';
  label?: React.ReactNode;
}

export const StatusBadge = React.forwardRef<HTMLSpanElement, StatusBadgeProps>(
  ({ className = '', status, label, children, ...props }, ref) => {
    let statusClasses = '';
    
    switch (status) {
      case 'healthy':
      case 'active':
        statusClasses = 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
        break;
      case 'warning':
        statusClasses = 'bg-amber-500/10 text-amber-400 border-amber-500/20';
        break;
      case 'danger':
        statusClasses = 'bg-red-500/10 text-red-400 border-red-500/20';
        break;
      case 'info':
        statusClasses = 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20';
        break;
      case 'default':
      default:
        statusClasses = 'bg-slate-500/10 text-slate-400 border-slate-500/20';
        break;
    }

    return (
      <span
        ref={ref}
        className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium border ${statusClasses} ${className}`}
        {...props}
      >
        {label || children}
      </span>
    );
  }
);
StatusBadge.displayName = 'StatusBadge';
