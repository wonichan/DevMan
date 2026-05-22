import React from 'react';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className = '', variant = 'secondary', size = 'md', isLoading, children, disabled, ...props }, ref) => {
    let variantClasses = '';
    switch (variant) {
      case 'primary':
        variantClasses = 'bg-emerald-600 hover:bg-emerald-500 text-white border border-emerald-500/50 shadow-[0_0_10px_rgba(16,185,129,0.2)]';
        break;
      case 'secondary':
        variantClasses = 'bg-[#334155] hover:bg-[#475569] text-slate-100 border border-[#475569]';
        break;
      case 'danger':
        variantClasses = 'bg-red-600 hover:bg-red-500 text-white border border-red-500/50 shadow-[0_0_10px_rgba(239,68,68,0.2)]';
        break;
      case 'ghost':
        variantClasses = 'bg-transparent hover:bg-[#334155] text-slate-300';
        break;
    }

    let sizeClasses = '';
    switch (size) {
      case 'sm':
        sizeClasses = 'text-xs px-3 py-1.5 rounded-lg';
        break;
      case 'md':
        sizeClasses = 'text-sm px-4 py-2 rounded-xl';
        break;
      case 'lg':
        sizeClasses = 'text-base px-6 py-3 rounded-xl';
        break;
    }

    const disabledClasses = (disabled || isLoading) ? 'opacity-50 cursor-not-allowed pointer-events-none' : '';
    const baseClasses = 'inline-flex items-center justify-center font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-emerald-500/50';

    return (
      <button
        ref={ref}
        disabled={disabled || isLoading}
        className={`${baseClasses} ${variantClasses} ${sizeClasses} ${disabledClasses} ${className}`}
        {...props}
      >
        {isLoading && (
          <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-current" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
        )}
        {children}
      </button>
    );
  }
);
Button.displayName = 'Button';
