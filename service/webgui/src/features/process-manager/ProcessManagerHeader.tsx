import { RefreshCcw } from "lucide-react"
import { Button } from "@/components/ui/button"

type ProcessManagerHeaderProps = {
  token: string
  error: string | null
  onRefresh: () => void
  onOpenCommandPalette: () => void
}

export function ProcessManagerHeader({ token, error, onRefresh, onOpenCommandPalette }: ProcessManagerHeaderProps) {
  return (
    <header className="rounded-xl border border-slate-100/10 bg-slate-950/40 p-5 backdrop-blur">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <img src="/icon.svg" alt="JPM" className="h-11 w-11" />
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">JPM Process Control</h1>
            <p className="text-sm text-slate-300">Monitor running apps and control lifecycle from the browser.</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            className="border-slate-100/20 bg-slate-900/40 text-slate-100 hover:bg-slate-800"
            onClick={onOpenCommandPalette}
          >
            <kbd className="pointer-events-none inline-flex h-5 select-none items-center rounded border border-slate-100/30 bg-slate-900 px-1.5 font-mono text-xs font-medium text-slate-300">
              ~
            </kbd>
            Commands
          </Button>
          <Button
            variant="outline"
            className="border-slate-100/20 bg-slate-900/40 text-slate-100 hover:bg-slate-800"
            onClick={onRefresh}
          >
            <RefreshCcw className="size-4" />
            Refresh
          </Button>
        </div>
      </div>

      {!token && (
        <p className="mt-3 text-sm text-amber-300">
          Token not found in URL hash. Add #token=... to enable authenticated requests.
        </p>
      )}

      {error && <p className="mt-3 text-sm text-red-300">{error}</p>}
    </header>
  )
}
