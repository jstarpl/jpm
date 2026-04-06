import { SquareTerminal } from "lucide-react"
import { Button } from "@/components/ui/button"
import type { Process, ProcessAction } from "./types"
import { formatUptime, statusClasses } from "./utils"

type ProcessTableProps = {
  processes: Process[]
  isLoading: boolean
  busyActionKey: string | null
  onOpenDetails: (processId: string) => void
  onOpenTerminal: (processId: string) => void
  onAction: (action: ProcessAction, processId: string) => void
}

export function ProcessTable({
  processes,
  isLoading,
  busyActionKey,
  onOpenDetails,
  onOpenTerminal,
  onAction,
}: ProcessTableProps) {
  return (
    <section className="rounded-xl border border-slate-100/10 bg-slate-950/40 p-2 backdrop-blur">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[840px] border-collapse text-left text-sm">
          <thead className="text-xs uppercase tracking-wide text-slate-300/80">
            <tr>
              <th className="px-4 py-3 font-medium">Name</th>
              <th className="px-4 py-3 font-medium">ID</th>
              <th className="px-4 py-3 font-medium">Exec</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Uptime</th>
              <th className="px-4 py-3 font-medium">Starts</th>
              <th className="px-4 py-3 font-medium">Exit</th>
              <th className="px-4 py-3 font-medium text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && (
              <tr>
                <td colSpan={8} className="px-4 py-6 text-center text-slate-300">
                  Loading processes...
                </td>
              </tr>
            )}

            {!isLoading && processes.length === 0 && (
              <tr>
                <td colSpan={8} className="px-4 py-6 text-center text-slate-300">
                  No processes are currently tracked.
                </td>
              </tr>
            )}

            {processes.map((process) => {
              const stopKey = `stop:${process.id}`
              const restartKey = `restart:${process.id}`
              const removeKey = `remove:${process.id}`

              return (
                <tr key={process.id} className="border-t border-slate-100/10">
                  <td className="px-4 py-3 align-middle font-medium">
                    {process.name || "(unnamed)"}
                    {process.namespace && <span className="ml-2 text-xs text-slate-400">[{process.namespace}]</span>}
                  </td>
                  <td className="px-4 py-3 align-middle text-slate-300">{process.id}</td>
                  <td className="px-4 py-3 align-middle text-slate-300">{process.exec}</td>
                  <td className="px-4 py-3 align-middle">
                    <span className={`rounded-md px-2 py-1 text-xs font-medium uppercase ${statusClasses(process.status)}`}>
                      {process.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 align-middle text-slate-300">{formatUptime(process.uptime)}</td>
                  <td className="px-4 py-3 align-middle text-slate-300">{process.startCount ?? 0}</td>
                  <td className="px-4 py-3 align-middle text-slate-300">{process.exitCode ?? 0}</td>
                  <td className="px-4 py-3 align-middle">
                    <div className="flex flex-wrap justify-end gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        className="border-slate-100/20 bg-transparent"
                        onClick={() => onOpenDetails(process.id)}
                      >
                        Details
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        className="border-slate-100/20 bg-transparent"
                        onClick={() => onOpenTerminal(process.id)}
                      >
                        <SquareTerminal className="size-4" />
                        Terminal
                      </Button>
                      <Button
                        size="sm"
                        variant="secondary"
                        disabled={busyActionKey === restartKey}
                        onClick={() => onAction("restart", process.id)}
                      >
                        Restart
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        className="border-amber-500/40 text-amber-200 hover:bg-amber-500/10"
                        disabled={busyActionKey === stopKey}
                        onClick={() => onAction("stop", process.id)}
                      >
                        Stop
                      </Button>
                      <Button
                        size="sm"
                        variant="destructive"
                        disabled={busyActionKey === removeKey}
                        onClick={() => onAction("remove", process.id)}
                      >
                        Remove
                      </Button>
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </section>
  )
}
