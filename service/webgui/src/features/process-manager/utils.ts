import type { ProcessStatus } from "./types"

export function readTokenFromHash(hash: string): string {
  const normalized = hash.startsWith("#") ? hash.slice(1) : hash
  const params = new URLSearchParams(normalized)
  return params.get("token") ?? ""
}

export function formatUptime(ms?: number): string {
  if (!ms || ms <= 0) {
    return "-"
  }

  const seconds = Math.floor(ms / 1000)
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60

  return `${h.toString().padStart(2, "0")}:${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`
}

export function statusClasses(status: ProcessStatus): string {
  switch (status) {
    case "running":
      return "bg-emerald-500/15 text-emerald-300 ring-1 ring-emerald-500/30"
    case "starting":
      return "bg-amber-500/15 text-amber-300 ring-1 ring-amber-500/30"
    case "stopping":
      return "bg-orange-500/15 text-orange-300 ring-1 ring-orange-500/30"
    case "failed":
      return "bg-red-500/15 text-red-300 ring-1 ring-red-500/30"
    case "respawn":
      return "bg-violet-500/15 text-violet-300 ring-1 ring-violet-500/30"
    default:
      return "bg-slate-500/15 text-slate-300 ring-1 ring-slate-500/30"
  }
}
