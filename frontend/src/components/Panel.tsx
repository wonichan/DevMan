import { ReactNode, MouseEventHandler } from 'react';

interface Props {
  children: ReactNode;
  className?: string;
  onClick?: MouseEventHandler<HTMLDivElement>;
}

export default function Panel({ children, className = '', onClick }: Props) {
  return (
    <div
      onClick={onClick}
      className={`
        bg-devman-panel border border-devman-border/30 rounded-[22px]
        shadow-lg shadow-black/10
        ${className}
      `}
    >
      {children}
    </div>
  );
}
