import { createContext, useContext, useRef, useCallback, useState, useMemo, useEffect } from 'react';
import type { ReactNode } from 'react';

export type PageAction = 'refresh';

export interface PageActions {
  refresh?: () => void | Promise<void>;
}

interface PageActionRegistry {
  register: (page: string, getActions: () => PageActions) => void;
  unregister: (page: string) => void;
  invoke: (page: string, action: PageAction) => boolean;
}

const PageActionContext = createContext<PageActionRegistry | undefined>(undefined);

export function PageActionProvider({ children }: { children: ReactNode }) {
  const actionsMap = useRef<Map<string, () => PageActions>>(new Map());
  const [, setTick] = useState(0);

  const register = useCallback((page: string, getActions: () => PageActions) => {
    actionsMap.current.set(page, getActions);
    setTick((t) => t + 1);
  }, []);

  const unregister = useCallback((page: string) => {
    actionsMap.current.delete(page);
    setTick((t) => t + 1);
  }, []);

  const invoke = useCallback((page: string, action: PageAction): boolean => {
    const getActions = actionsMap.current.get(page);
    if (!getActions) return false;
    const actions = getActions();
    if (action === 'refresh' && actions.refresh) {
      actions.refresh();
      return true;
    }
    return false;
  }, []);

  const registry = useMemo(() => ({ register, unregister, invoke }), [register, unregister, invoke]);

  return (
    <PageActionContext.Provider value={registry}>
      {children}
    </PageActionContext.Provider>
  );
}

export function usePageActionRegistry(): PageActionRegistry {
  const ctx = useContext(PageActionContext);
  if (!ctx) {
    throw new Error('usePageActionRegistry must be used within PageActionProvider');
  }
  return ctx;
}

export function usePageActions(page: string, actions: PageActions) {
  const registry = usePageActionRegistry();
  const actionsRef = useRef(actions);
  actionsRef.current = actions;

  useEffect(() => {
    registry.register(page, () => actionsRef.current);
    return () => registry.unregister(page);
  }, [registry, page]);
}

export { PageActionContext };
export type { PageActionRegistry };
