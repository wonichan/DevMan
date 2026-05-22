import React from 'react';

export interface SurfaceCardProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'raised' | 'interactive' | 'selected' | 'danger';
}

export const SurfaceCard = React.forwardRef<HTMLDivElement, SurfaceCardProps>(
  ({ className = '', variant = 'default', children, ...props }, ref) => {
    let variantClasses = 'bg-[#1e293b]/80 border-[#334155]';
    
    switch (variant) {
      case 'raised':
        variantClasses = 'bg-[#1e293b] border-[#475569] shadow-lg';
        break;
      case 'interactive':
        variantClasses = 'bg-[#1e293b]/60 border-[#334155] hover:bg-[#1e293b] hover:border-[#475569] transition-colors cursor-pointer';
        break;
      case 'selected':
        variantClasses = 'bg-[#0f172a] border-emerald-500/50 shadow-[0_0_15px_rgba(16,185,129,0.15)] ring-1 ring-emerald-500/20';
        break;
      case 'danger':
        variantClasses = 'bg-red-950/20 border-red-900/50';
        break;
    }

    return (
      <div
        ref={ref}
        className={`rounded-xl border backdrop-blur-sm p-5 ${variantClasses} ${className}`}
        {...props}
      >
        {children}
      </div>
    );
  }
);
SurfaceCard.displayName = 'SurfaceCard';
