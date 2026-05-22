import React from 'react';

export interface PageHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string;
  description?: React.ReactNode;
  actions?: React.ReactNode;
}

export const PageHeader = React.forwardRef<HTMLDivElement, PageHeaderProps>(
  ({ className = '', title, description, actions, ...props }, ref) => {
    return (
      <div 
        ref={ref}
        className={`flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6 ${className}`}
        {...props}
      >
        <div>
          <h1 className="text-2xl font-bold text-slate-100 tracking-tight">{title}</h1>
          {description && (
            <p className="text-sm text-slate-400 mt-1">{description}</p>
          )}
        </div>
        {actions && (
          <div className="flex items-center gap-3">
            {actions}
          </div>
        )}
      </div>
    );
  }
);
PageHeader.displayName = 'PageHeader';
