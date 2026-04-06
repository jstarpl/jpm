import { ProcessDetailsSheet } from "./ProcessDetailsSheet"
import { ProcessManagerHeader } from "./ProcessManagerHeader"
import { ProcessTable } from "./ProcessTable"
import { TerminalDialog } from "./TerminalDialog"
import { useProcessManager } from "./useProcessManager"

export function ProcessManagerPage() {
  const {
    token,
    processes,
    isLoading,
    error,
    busyActionKey,
    selectedProcess,
    terminalProcess,
    fetchProcesses,
    runAction,
    setSelectedProcessId,
    setTerminalProcessId,
  } = useProcessManager()

  return (
    <div className="min-h-svh bg-black text-slate-100">
      <main className="mx-auto flex w-full max-w-7xl flex-col gap-6 px-4 py-8 sm:px-6 lg:px-8">
        <ProcessManagerHeader token={token} error={error} onRefresh={fetchProcesses} />

        <ProcessTable
          processes={processes}
          isLoading={isLoading}
          busyActionKey={busyActionKey}
          onOpenDetails={setSelectedProcessId}
          onOpenTerminal={setTerminalProcessId}
          onAction={runAction}
        />
      </main>

      <ProcessDetailsSheet process={selectedProcess} onClose={() => setSelectedProcessId(null)} />

      <TerminalDialog
        processId={terminalProcess?.id ?? null}
        token={token}
        onOpenChange={(open) => {
          if (!open) {
            setTerminalProcessId(null)
          }
        }}
      />
    </div>
  )
}
