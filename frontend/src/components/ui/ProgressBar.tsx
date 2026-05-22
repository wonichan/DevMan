import React from 'react';

export interface ProgressBarProps extends React.HTMLAttributes<HTMLDivElement> {
  value: number;
  variant?: 'accent' | 'info' | 'warning' | 'danger';
}

export const ProgressBar = React.forwardRef<HTMLDivElement, ProgressBarProps>(
  ({ className = '', value, variant = 'accent', ...props }, ref) => {
    const clampedValue = Math.min(100, Math.max(0, value));
    
    let colorClasses = '';
    switch (variant) {
      case 'accent':
        colorClasses = 'bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.5)]';
        break;
      case 'info':
        colorClasses = 'bg-cyan-500 shadow-[0_0_10px_rgba(6,182,212,0.5)]';
        break;
      case 'warning':
        colorClasses = 'bg-amber-500 shadow-[0_0_10px_rgba(245,158,11,0.5)]';
        break;
      case 'danger':
        colorClasses = 'bg-red-500 shadow-[0_0_10px_rgba(239,68,68,0.5)]';
        break;
    }

    return (
      <div 
        ref={ref}
        className={`w-full bg-[#0f172a] rounded-full h-2.5 overflow-hidden border border-[#1e293b] ${className}`}
        {...props}
      >
        <div 
          className={`h-full rounded-full transition-all duration-500 ease-out ${colorClasses}`}
          style={{ width: `${clampedValue}%` }}
        />
      </div>
    );
  }
);
ProgressBar.displayName = 'ProgressBar';
