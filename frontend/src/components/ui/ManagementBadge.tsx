interface ManagementBadgeProps {
  managed: boolean;
  className?: string;
}

export function ManagementBadge({ managed, className = '' }: ManagementBadgeProps) {
  return (
    <span
      className={`text-xs px-2 py-0.5 rounded-md ${
        managed ? 'bg-emerald-500/10 text-emerald-400' : 'bg-slate-700 text-slate-400'
      } ${className}`}
    >
      {managed ? 'Managed' : 'Unmanaged'}
    </span>
  );
}
