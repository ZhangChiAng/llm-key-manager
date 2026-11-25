/** Must match the Status constants in internal/probe. */
export type ProbeStatus =
  | 'available'
  | 'authentication_failed'
  | 'insufficient_balance'
  | 'permission_denied'
  | 'rate_limited'
  | 'failed'
  | 'unsupported'

/** Includes transient and not-yet-checked UI states, which have no result label. */
export type ProbeDisplayState = ProbeStatus | 'probing' | ''

/** User-facing conclusions shared by all probe result displays. */
export const probeStatusLabels: Record<ProbeStatus, string> = {
  available: '可用',
  authentication_failed: '认证失败',
  insufficient_balance: '余额不足',
  permission_denied: '权限受限',
  rate_limited: '限流',
  failed: '检测失败',
  unsupported: '不支持检测',
}

/** Leaves pending and absent results blank; probing uses a separate spinner. */
export function probeStatusLabel(state: ProbeDisplayState): string {
  return state === '' || state === 'probing' ? '' : probeStatusLabels[state]
}
