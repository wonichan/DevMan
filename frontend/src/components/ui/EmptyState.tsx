import React from 'react';
import { SurfaceCard } from './SurfaceCard';

export interface EmptyStateProps extends React.HTMLAttributes<HTMLDivElement> {
  icon?: React.ReactNode;
  title: string;
  description?: React.ReactNode;
  action?: React.ReactNode;
}

export const EmptyState = React.forwardRef<HTMLDivElement, EmptyStateProps>(
  ({ className = '', icon, title, description, action, ...props }, ref) => {
    return (
      <SurfaceCard 
        ref={ref}
        className={`flex flex-col items-center justify-center p-10 text-center ${className}`}
        {...props}
      >
        {icon && (
          <div className="w-12 h-12 rounded-full bg-[#1e293b] flex items-center justify-center text-slate-400 mb-4 border border-[#334155]">
            {icon}
          </div>
        )}
        <h3 className="text-lg font-medium text-slate-200">{title}</h3>
        {description && (
          <p className="text-sm text-slate-400 mt-2 max-w-sm">{description}</p>
        )}
        {action && (
          <div className="mt-6">
            {action}
          </div>
        )}
      </SurfaceCard>
    );
  }
);
EmptyState.displayName = 'EmptyState';
