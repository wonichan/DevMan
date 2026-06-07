import type { ManagedVersion, VersionManagerConflict } from '../devman-types';

export function versionSourceLabel(source: ManagedVersion['Source']): string {
  switch (source) {
    case 'devman':
      return 'DevMan';
    case 'version_manager':
      return '版本管理器';
    case 'external':
    default:
      return '外部';
  }
}

export function conflictLabel(conflict?: VersionManagerConflict): string {
  if (!conflict?.Detected) return '';
  return `${conflict.Manager}: ${conflict.Evidence}`;
}
