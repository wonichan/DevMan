import type { CleanableItem, Env, EnvSummary } from '../devman-types';

export type ManagedLookup = Record<string, boolean>;

export function buildManagedLookup(envs: Env[]): ManagedLookup {
  return Object.fromEntries((envs || []).map((env) => [env.Key, env.IsManaged]));
}

export function sortEnvSummariesByManagement(list: EnvSummary[]): EnvSummary[] {
  return [...list].sort((a, b) => {
    if (a.Env.IsManaged !== b.Env.IsManaged) {
      return a.Env.IsManaged ? -1 : 1;
    }
    return a.Env.Name.localeCompare(b.Env.Name);
  });
}

export function sortCleanableItemsByManagement(items: CleanableItem[], managedByKey: ManagedLookup): CleanableItem[] {
  return [...items].sort((a, b) => {
    const aManaged = Boolean(managedByKey[a.EnvKey]);
    const bManaged = Boolean(managedByKey[b.EnvKey]);
    if (aManaged !== bManaged) return aManaged ? -1 : 1;
    return a.Name.localeCompare(b.Name);
  });
}
