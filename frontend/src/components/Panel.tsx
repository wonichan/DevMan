import { ReactNode, MouseEventHandler } from 'react';
import { SurfaceCard } from './ui/SurfaceCard';

interface Props {
  children: ReactNode;
  className?: string;
  onClick?: MouseEventHandler<HTMLDivElement>;
}

export default function Panel({ children, className = '', onClick }: Props) {
  return (
    <SurfaceCard
      onClick={onClick}
      variant={onClick ? 'interactive' : 'default'}
      className={className}
    >
      {children}
    </SurfaceCard>
  );
}
